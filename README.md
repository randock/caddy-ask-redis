# caddy-ask-redis

[Caddy Server](https://caddyserver.com/) v2 module for use in [tls_on_demand](https://caddyserver.com/docs/automatic-https#on-demand-tls) ask property.
The module will check if requested domain is a member of the Redis set defined by `key`.

## Installation
```
xcaddy build \
    --with github.com/CruGlobal/caddy-ask-redis
```

## Usage
Caddyfile
```
# Global config
{
   	on_demand_tls {
		ask http://localhost:1111
	}
}

http://localhost:1111 {
    route {
        skip_log
        ask_redis {
            host {$REDIS_HOST}
            port {$REDIS_PORT}
            db {$REDIS_DB_INDEX}
            key certificates:domain_allowlist
        }
    }
}
```
