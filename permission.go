package caddy_ask_redis

import (
	"context"
	"fmt"
	"regexp"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	DefaultRedisAddress  = "127.0.0.1:6379"
	DefaultRedisUsername = "default"
)

var (
	_ caddy.Module                = (*PermissionByRedis)(nil)
	_ caddy.Provisioner           = (*PermissionByRedis)(nil)
	_ caddy.Validator             = (*PermissionByRedis)(nil)
	_ caddyfile.Unmarshaler       = (*PermissionByRedis)(nil)
	_ caddy.CleanerUpper          = (*PermissionByRedis)(nil)
	_ caddytls.OnDemandPermission = (*PermissionByRedis)(nil)
	_ caddy.Provisioner           = (*PermissionByRedis)(nil)
)

func init() {
	caddy.RegisterModule(PermissionByRedis{})
}

type PermissionByRedis struct {
	Client   *redis.Client
	logger   *zap.Logger
	WwwRegex *regexp.Regexp

	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Prefix   string `json:"prefix"`
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
		m.logger = ctx.Logger(m)
	}

	m.WwwRegex = regexp.MustCompile(`^www\.`)

	m.logger.Info(fmt.Sprintf("Creating new Redis client %s", m.Address))

	m.Client = redis.NewClient(&redis.Options{
		Addr:     m.Address,
		Username: m.Username,
		Password: m.Password,
		DB:       0,
	})

	err := m.Client.Ping(ctx).Err()
	if err != nil {
		return err
	}

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
		case "address":
			if value != "" {
				m.Address = value
			} else {
				m.Address = DefaultRedisAddress
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
		case "prefix":
			if value != "" {
				m.Prefix = value
			} else {
				m.Prefix = ""
			}
		}
	}

	return nil
}

func (m PermissionByRedis) CertificateAllowed(ctx context.Context, name string) error {

	// remove www.
	name = m.WwwRegex.ReplaceAllString(name, "")

	// compil key and check existance
	redisKey := fmt.Sprintf("%s%s", m.Prefix, name)
	val, err := m.Client.Exists(ctx, redisKey).Result()

	if err != nil {
		return fmt.Errorf("%s: %w (error looking up %s - %s)", name, caddytls.ErrPermissionDenied, redisKey, err)
	}

	m.logger.Debug(fmt.Sprintf("Allowing certificate for %s: %d", name, val))

	if val == 1 {
		return nil
	}

	return fmt.Errorf("%s: %w (redis key %s not found)", name, caddytls.ErrPermissionDenied, redisKey)
}
