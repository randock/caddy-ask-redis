package caddy_ask_redis

import (
	"context"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	DefaultRedisHost     = "127.0.0.1"
	DefaultRedisPort     = "6379"
	DefaultRedisUsername = "default"
)

var (
	_ caddy.Module          = (*PermissionByRedis)(nil)
	_ caddy.Provisioner     = (*PermissionByRedis)(nil)
	_ caddy.Validator       = (*PermissionByRedis)(nil)
	_ caddyfile.Unmarshaler = (*PermissionByRedis)(nil)
	_ caddy.CleanerUpper    = (*PermissionByRedis)(nil)
	_ caddy.Provisioner     = (*PermissionByRedis)(nil)
)

func init() {
	caddy.RegisterModule(PermissionByRedis{})
}

type PermissionByRedis struct {
	Client *redis.Client
	logger *zap.SugaredLogger

	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CaddyModule returns the Caddy module information.
func (PermissionByRedis) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "tls.permission.redis",
		New: func() caddy.Module {
			return new(PermissionByRedis)
		},
	}
}

func (m *PermissionByRedis) Cleanup() error {
	// Close the Redis connection
	if m.Client != nil {
		m.Client.Close()
	}

	return nil
}

// Provision implements the caddy.Provisioner interface.
func (m *PermissionByRedis) Provision(ctx caddy.Context) error {
	if m.logger == nil {
		m.logger = ctx.Logger(m).Sugar()
	}

	m.Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", m.Host, m.Port),
		Username: m.Username,
		Password: m.Password,
		DB:       0,
	})

	return nil
}

func (m *PermissionByRedis) Validate() error {
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler. Syntax:
//
//	ask_redis <secret>
func (m *PermissionByRedis) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		key := d.Val()
		var value string

		if !d.Args(&value) {
			continue
		}

		switch key {
		case "host":
			if value != "" {
				m.Host = value
			} else {
				m.Host = DefaultRedisHost
			}
		case "port":
			if value != "" {
				m.Port = value
			} else {
				m.Port = DefaultRedisPort
			}
		case "username":
			if value != "" {
				m.Username = value
			} else {
				m.Username = DefaultRedisUsername
			}
		case "password":
			if value != "" {
				m.Password = value
			} else {
				m.Password = ""
			}
		}
	}

	return nil
}

func (p PermissionByRedis) CertificateAllowed(ctx context.Context, name string) error {
	redisKey := fmt.Sprintf("certificates/%s", name)
	val, err := p.Client.Exists(ctx, redisKey).Result()

	if err != nil {
		return fmt.Errorf("%s: error looking up %s, %w", name, redisKey, err)
	}

	if val == 1 {
		return nil
	}

	return fmt.Errorf("%s: redis key %s not found", name, redisKey)
}
