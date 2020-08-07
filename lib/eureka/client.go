package eureka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ArthurHlt/go-eureka-client/eureka"
)

func GetInstanceURL(host, id string) (string, error) {
	client := eureka.NewClient([]string{normalizeHost(host) + "/eureka"})
	applications, err := client.GetApplications()

	if err != nil {
		return "", fmt.Errorf("could not find application in eureka err:%s", err)
	}

	for _, registeredApp := range applications.Applications {
		if isDesiredService(registeredApp.Name, id) {
			for _, registeredInstance := range registeredApp.Instances {
				return removeSlash(registeredInstance.HomePageUrl), nil
			}
		}
	}

	return "", fmt.Errorf("no application in eureka matches the id: %v", id)
}

func RegisterInstance(eurekaHost string, instance *Instance) error {

	url := fmt.Sprintf("%s/eureka/apps/%s", normalizeHost(eurekaHost), instance.App)

	body, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("could not register application with id: %s err: %s", instance.InstanceID, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not register application wtih id: %s eureka responded with status: %d body: %s", instance.InstanceID, resp.StatusCode, string(body))
	}

	return nil
}

//Example: localhost:8761, foo-service, FOO:8081
func UnregisterApp(eurekaHost, app, hostName string) error {

	url := fmt.Sprintf("%s/eureka/apps/%s/%s", normalizeHost(eurekaHost), app, hostName)

	req, err := http.NewRequest(http.MethodDelete, url, nil)

	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	return nil
}

func normalizeHost(host string) string {

	if strings.HasPrefix(host, "http://") {
		return host
	}

	return "http://" + host
}

func isDesiredService(got, want string) bool {
	return strings.HasSuffix(strings.ToLower(got), strings.ToLower(want))
}

func removeSlash(url string) string {
	return url[0 : len(url)-1]
}
