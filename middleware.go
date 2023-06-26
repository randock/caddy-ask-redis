package caddy_ask_redis

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const (
	DefaultRedisHost = "127.0.0.1"
	DefaultRedisPort = "6379"
	DefaultRedisDB   = 0
)

var (
	_ caddy.Module                = (*Middleware)(nil)
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("ask_redis", parseCaddyfileHandler)
}

type Middleware struct {
	Client *redis.Client
	logger *zap.SugaredLogger
	ctx    context.Context

	Host string `json:"host"`
	Port string `json:"port"`
	DB   int    `json:"db"`
	Key  string `json:"key"`
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.ask_redis",
		New: func() caddy.Module {
			return new(Middleware)
		},
	}
}

// Provision implements the caddy.Provisioner interface.
func (m *Middleware) Provision(ctx caddy.Context) error {
	if m.logger == nil {
		m.logger = ctx.Logger(m).Sugar()
	}

	m.ctx = ctx

	m.Client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", m.Host, m.Port),
		DB:   m.DB,
	})

	return nil
}

// ServeHTTP implements the caddy.Handler interface.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	domain := r.URL.Query().Get("domain")
	if domain != "" {
		val, _ := m.Client.SIsMember(m.ctx, m.Key, domain).Result()
		if val {
			w.WriteHeader(http.StatusOK)
			return nil
		}
	}
	w.WriteHeader(http.StatusNotFound)
	return nil
}

func (m *Middleware) Validate() error {
	if m.Key == "" {
		return fmt.Errorf("empty key")
	}

	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler. Syntax:
//
//	ask_redis <secret>
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
		case "db":
			if value != "" {
				dbParse, err := strconv.Atoi(value)
				if err == nil {
					m.DB = dbParse
				} else {
					m.DB = DefaultRedisDB
				}
			} else {
				m.DB = DefaultRedisDB
			}
		case "key":
			if value != "" {
				m.Key = value
			}
		}
	}
	return nil
}

func parseCaddyfileHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler,
	error) {
	m := new(Middleware)
	if err := m.UnmarshalCaddyfile(h.Dispenser); err != nil {
		return nil, err
	}

	return m, nil
}
