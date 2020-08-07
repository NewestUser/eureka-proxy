# Command line Eureka proxy

The eureka proxy will forward traffic to an environment which already has all the 
service dependencies that you need without allowing your locally running service 
to interfere with the environment.

Locally running services will communicate with each other, those that are not running 
will be retrieved from the eureka that is used to proxy against.

#### Setup
 - download and [setup](https://golang.org/doc/install) Go
 - [recompile](https://golang.org/doc/code.html#Command) [eureka-proxy](./cmd/eureka-proxy)


## Usage manual
```console 
Usage: eureka-proxy [global flags] <url>

global flags:
  -fake value
        ServiceID and Port of a dummy application which will be added to the list of registered services
        example: foo-service:8081
  -pollute
        Allow services to reach the real Eureka instance.
  -port int
        Port on which to start the proxy (default 8761)
  -strip string
        Strip or replace part of url
  -trace
        Print all HTTP communication
  -v    Print version and exit

example:
        eureka-proxy http://my-dev-environment.net:8761
```

The client can also accept a configuration file.
This can be used to register fake services that are not located on the host machine.
For an example configuration see [config.yml](eureka/cmd/eureka-proxy/config.yml).

Example usage with a configuration file:
```
eureka-proxy [global-flags] ./path/to/config.yml
```

#### Additional
If you want to proxy requests without the eureka hustle checkout [reverse-proxy](./cmd/reverse-proxy).