package reverse

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/newestuser/eureka-proxy/lib/logging"
	"github.com/newestuser/eureka-proxy/lib/reverse-proxy/strip"
	"github.com/rs/cors"
)

type Proxy interface {
	Start() error

	ServeHTTP(http.ResponseWriter, *http.Request)
}

func SingleRoute(route, strip string, target *url.URL) []*RouteConfig {
	return []*RouteConfig{NewRouteConfig(route, strip, target)}
}

func NewRouteConfig(route, strip string, target *url.URL) *RouteConfig {
	return &RouteConfig{Route: route, PathStrip: strip, TargetURL: target}
}

type ProxyConfig struct {
	Routes     []*RouteConfig
	Port       int
	Trace      bool
	LoggingOff bool
	EnableCORS bool
}

type RouteConfig struct {
	Route     string
	TargetURL *url.URL
	PathStrip string
}

func (r RouteConfig) String() string {
	if r.PathStrip == "" {
		return fmt.Sprintf("Route(from:'%v' to:'%v')", r.Route, r.TargetURL)
	}

	return fmt.Sprintf("Route(from:'%v' to:'%v' strip:'%v')", r.Route, r.TargetURL, r.PathStrip)
}

func NewReverseProxy(conf *ProxyConfig) (Proxy, error) {
	router := mux.NewRouter()
	logger := logging.NewLevelLogger(conf.Trace, !conf.LoggingOff)

	for _, route := range conf.Routes {
		rHandler, err := reverseHandler(logger, route)

		if err != nil {
			return nil, err
		}

		router.PathPrefix(route.Route).Handler(rHandler)
	}

	var proxyHandler http.Handler = router

	if conf.EnableCORS {
		proxyHandler = cors.AllowAll().Handler(proxyHandler)
	}

	return &reverseProxy{
		router: proxyHandler,
		logger: logger,
		conf:   conf,
	}, nil
}

type reverseProxy struct {
	router http.Handler
	logger logging.Logger
	conf   *ProxyConfig
}

func (proxy *reverseProxy) Start() error {

	proxy.logger.InfoF("Reverse proxy starting on port %d\n", proxy.conf.Port)

	for _, r := range proxy.conf.Routes {
		proxy.logger.InfoF("Proxying to %s\n", r.String())
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", proxy.conf.Port), proxy.router)
}

func (proxy *reverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy.router.ServeHTTP(w, r)
}

func reverseHandler(logger logging.Logger, c *RouteConfig) (http.Handler, error) {

	s, err := strip.New(c.PathStrip)

	if err != nil {
		return nil, err
	}

	reverseHandler := httputil.NewSingleHostReverseProxy(c.TargetURL)
	logHandler := logging.NewHandler(logger, reverseHandler)
	stripHandler := strip.NewHandler(s, logHandler)

	return stripHandler, err
}
