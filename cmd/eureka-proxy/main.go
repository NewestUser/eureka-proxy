package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/newestuser/eureka-proxy/lib/eureka/fake"
	"github.com/newestuser/eureka-proxy/lib/flags"
	"github.com/newestuser/eureka-proxy/lib/logging"
	"github.com/newestuser/eureka-proxy/lib/netutil"
	"github.com/newestuser/eureka-proxy/lib/reverse-proxy"
)

const version = "v1.0"

func main() {

	fs := flags.NewFlagSet("eureka-proxy")

	fs.Usage = func() {
		fmt.Println("\nUsage: eureka-proxy [global flags] <url>")
		fmt.Printf("\nglobal flags:\n")
		fs.PrintDefaults()
		fmt.Print(example)
		return
	}

	versionFlag := fs.BoolFlag("v", false, "Print version and exit")
	portFlag := fs.IntFlag("port", 8761, "Port on which to start the proxy")
	stripFlag := fs.StringFlag("strip", "", "Strip or replace part of url")
	traceFlag := fs.BoolFlag("trace", false, "Print all HTTP communication")
	fakeFlag := fs.StringArrFlag("fake", "", "ServiceID and Port of a dummy application which will be added to the list of registered services \nexample: foo-service:8081")
	polluteFlag := fs.BoolFlag("pollute", false, "Allow services to register in the real Eureka instance")

	args := fs.ParseArgs()

	if versionFlag.IsSet() {
		fmt.Println(version)
		os.Exit(0)
	}

	if args.IsEmpty() {
		fmt.Println("Specify eureka url or valid config file")
		fs.Usage()
		os.Exit(1)
	}

	arg := args.First()

	var routes []*reverse.RouteConfig = nil
	var eurekaUrl *url.URL = nil
	var fakes = make([]*fake.Application, 0)

	if isUrl, urlArg := arg.IsURL(); isUrl {
		routes = reverse.SingleRoute("/", stripFlag.Get(), urlArg)
		eurekaUrl = urlArg
	} else if isFile, bytes := arg.IsFile(); isFile {
		routes, eurekaUrl, fakes = parseYmlFile(bytes)
	} else {
		log.Fatal(fmt.Sprintf("Please provide a valid eureka URL like http://ziongw1-dev.neterra.skrill.net:8761 or a configuration file"))
	}

	c := &reverse.ProxyConfig{
		Routes:     routes,
		Port:       portFlag.Get(),
		LoggingOff: true,
	}

	proxy, err := reverse.NewReverseProxy(c)

	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to initialize proxy, err:%s\n", err.Error()))
	}

	if fakeFlag.IsSet() {
		for _, serviceAndPort := range fakeFlag.Values() {
			fakes = append(fakes, fakeApp(serviceAndPort))
		}
	}

	handler := fake.RequestHandler(fakes, polluteFlag.Get(), proxy)
	handler = loggingHandler(handler, traceFlag.Get())

	log.Printf("Reverse proxy starting on port %d\n", c.Port)
	log.Printf("Proxying to %s\n", eurekaUrl.String())

	for _, fakeApp := range fakes {
		log.Printf("Injecting %s\n\n", fakeApp)
	}

	if err = http.ListenAndServe(fmt.Sprintf(":%d", c.Port), handler); err != nil {
		log.Fatal(fmt.Sprintf("Unable to start proxy, err:%s", err.Error()))
	}
}

func fakeApp(serviceAndPort string) *fake.Application {
	serviceID, port := flags.ParseIdAndPort(serviceAndPort)

	return fake.SingleLocalApp(serviceID, port)
}

func parseYmlFile(bytes []byte) ([]*reverse.RouteConfig, *url.URL, []*fake.Application) {

	type fakeAppConfig struct {
		Id       string `yaml:"id"`
		Ip       string `yaml:"ip"`
		HostName string `yaml:"hostname"`
	}

	type routeConfig struct {
		Proxy struct {
			EurekaUrl string           `yaml:"eurekaUrl"`
			Port      string           `yaml:"port"`
			Fakes     []*fakeAppConfig `yaml:"fakes"`
		}
	}

	config := &routeConfig{}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		log.Fatalf("could not parse yaml file err: %s\n", err.Error())
	}

	if config.Proxy.EurekaUrl == "" {
		log.Fatalf("please specify a valid eurekaUrl in the yml configuration\n")
	}

	targetUrl, err := url.Parse(config.Proxy.EurekaUrl)
	if err != nil {
		log.Fatalf("please provide a valid eurekaUrl err:%s\n", err.Error())
	}

	routes := reverse.SingleRoute("/", "", targetUrl)
	fakes := make([]*fake.Application, 0)

	for _, fakeConfig := range config.Proxy.Fakes {

		serviceId, port := flags.ParseIdAndPort(fakeConfig.Id)
		ip := valOrDefault(fakeConfig.Ip, netutil.OutboundIP().String)
		host := valOrDefault(fakeConfig.HostName, defaultHost)

		fakeApp := fake.SingleInstanceApp(serviceId, serviceId, ip, host, port)
		fakes = append(fakes, fakeApp)
	}

	return routes, targetUrl, fakes
}

func defaultHost() string {
	return fmt.Sprintf("%s.EUREKA-PROXY.FAKE", netutil.Hostname())
}

func valOrDefault(val string, fallback func() string) string {
	if val == "" {
		return fallback()
	}

	return val
}

// Logging should be done before the request is intercepted by the eureka proxy because
// the request might not reach the original reverse proxy that comes in with build in logging
func loggingHandler(chain http.Handler, traceOn bool) http.Handler {

	logger := logging.NewLevelLogger(traceOn, true)
	return logging.NewHandler(logger, chain)
}

const example = `
example:
        eureka-proxy http://my-dev-environment.net:8761
`
