# caddy-ask-redis

[Caddy Server](https://caddyserver.com/) v2 module for use in [tls_on_demand](https://caddyserver.com/docs/automatic-https#on-demand-tls) ask property.
The module will check if requested domain is a member of the Redis set defined by `key`.

## Installation
```
xcaddy build \
    --with github.com/randock/caddy-ask-redis
```

## Usage
Caddyfile
```
# Global config
{
   	on_demand_tls {
		permission redis {
            host {$REDIS_HOST}
            port {$REDIS_PORT}
            username {$REDIS_USERNAME}
            password {$REDIS_PASSWORD}
        }
	}
}
```
