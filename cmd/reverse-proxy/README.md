# Command line tool for proxying requests against a single or multiple targets

#### Setup
 - download and [setup](https://golang.org/doc/install) Go
 - [recompile](https://golang.org/doc/code.html#Command) reverse-proxy

## Usage manual
```console 
Usage: reverse-proxy [global flags] <url>

global flags:
  -enable-cors
        enable CORS requests
  -port int
        proxy port (default 8080)
  -strip string
        strip or replace part of url
  -trace
        trace proxied requests
  -v    proxy version

example:
        reverse-proxy http://foo-service.net:8080
```

## Proxy to multiple targets
Example configuration in `routes.yml`:

```yml 
proxy:
  routes:
    foo-route:
      path: /foo/
      url: http://localhost:4200
      stripPrefix: false
    bar-route:
      path: /bar-api/
      url: http://bar-service.net:8080
      stripPrefix: false
```

Usage:
```console
reverse-rpoxy routes.yml
```