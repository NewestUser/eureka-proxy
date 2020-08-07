package main

import (
	"fmt"
	"github.com/newestuser/eureka-proxy/lib/flags"
	"github.com/newestuser/eureka-proxy/lib/reverse-proxy"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
)

const version = "v1.3"

func main() {

	fs := flags.NewFlagSet("reverse-proxy")

	versionFlag := fs.BoolFlag("v", false, "proxy version")
	portFlag := fs.IntFlag("port", 4400, "proxy port")
	stripFlag := fs.StringFlag("strip", "", "strip or replace part of url")
	traceFlag := fs.BoolFlag("trace", false, "trace proxied requests")
	enableCorsFlag := fs.BoolFlag("enable-cors", false, "enable CORS requests")


	fs.Usage = func() {
		fmt.Println("\nUsage: reverse-proxy [global flags] <url>")
		fmt.Printf("\nglobal flags:\n")
		fs.PrintDefaults()
		fmt.Print(example)
		return
	}

	fs.ParseArgs()

	if versionFlag.IsSet() {
		fmt.Println(version)
		os.Exit(0)
	}

	args := fs.ParseArgs()

	if args.IsEmpty() {
		fmt.Println("Specify url to proxy against or valid config file")
		fs.Usage()
		os.Exit(1)
	}

	urlOrFile := args.First()

	var routes []*reverse.RouteConfig = nil

	if isFile, bytes := urlOrFile.IsFile(); isFile {

		parsedConfig := readRouteConfiguration(bytes)
		routes = adaptRouteConfiguration(parsedConfig)

	} else if isUrl, targetUrl := urlOrFile.IsURL(); isUrl {
		routes = reverse.SingleRoute("/", stripFlag.Get(), targetUrl)

	} else {
		log.Fatal(fmt.Sprintf("Please provide a valid URL or YAML configuration as an argument."))
	}

	c := &reverse.ProxyConfig{
		Routes: routes,
		Port:   portFlag.Get(),
		Trace:  traceFlag.Get(),
		EnableCORS: enableCorsFlag.Get(),
	}

	proxy, err := reverse.NewReverseProxy(c)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to start proxy, err:%s", err.Error()))
	}

	if err = proxy.Start(); err != nil {
		log.Fatal(fmt.Sprintf("Unable to start proxy, err:%s", err.Error()))
	}
}

type routeConfig struct {
	Proxy struct {
		Routes map[string]struct {
			Path        string `yaml:"path"`
			Url         string `yaml:"url"`
			StripPrefix bool   `yaml:"stripPrefix"`
		}
	}
}

func readRouteConfiguration(fileBytes []byte) *routeConfig {

	config := &routeConfig{}
	if err := yaml.Unmarshal(fileBytes, config); err != nil {
		log.Fatalf("could not parse yaml file err: %s\n", err.Error())
	}

	return config
}

func adaptRouteConfiguration(routeConfig *routeConfig) []*reverse.RouteConfig {
	routes := make([]*reverse.RouteConfig, 0)

	for routeLabel, route := range routeConfig.Proxy.Routes {
		routeURL, routeErr := url.Parse(route.Url)

		if routeErr != nil {
			log.Fatalf("the url %s for route %s is invalid, err:%s", route.Url, routeLabel, routeErr.Error())
		}

		strip := ""
		if route.StripPrefix {
			strip = fmt.Sprintf("%s:%s", route.Path, "")
		}

		routes = append(routes, reverse.SingleRoute(route.Path, strip, routeURL)...)
	}

	return routes
}

const example = `
example:
        reverse-proxy http://ziongw1-dev.neterra.skrill.net:8888
`