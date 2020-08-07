package fake

import (
	"fmt"
	eureka2 "github.com/newestuser/eureka-proxy/lib/eureka"
	"net/http"
	"strings"
)

// A representation of a fake application that will be injected in the list of eureka applications.
// This data structure is responsible for handling the http layer that is responsible for intercepting the traffic
// of a given eureka application.
type appCluster struct {
	ID string
	*Application
}

// Add all of the instances of the application to the fake application cluster.
func (clust *appCluster) add(app *Application) {
	if clust.ID == "" {
		clust.ID = app.ID
	}

	if !strings.EqualFold(clust.ID, app.ID) {
		panic(fmt.Errorf("cannot add an application with Id: %s to cluster with Id: %s", app.ID, clust.ID))
	}

	if clust.Application == nil {
		clust.Application = &Application{ID: app.ID}
	}

	clust.Application.AddInstances(app.Instances())
}

// Deregister a single instance from the application
func (clust *appCluster) deregister(instanceId string) (bool, *Target) {

	return clust.Application.RemoveInstance(instanceId)
}

// Check if there are any instances left, return true if there are no instances.
func (clust *appCluster) noInstances() bool {

	return clust.Application.NoInstances()
}

func (clust *appCluster) NewInstances() []*eureka2.Instance {

	return clust.Application.NewInstances()
}

func (clust *appCluster) NewEurekaApp() *eureka2.Application {
	return clust.Application.NewEurekaApp()
}

func (clust *appCluster) isRegistrationRequest(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}

	return caseInsensitiveContains(r.URL.Path, fmt.Sprintf("eureka/apps/%s", clust.ID))
}

func (clust *appCluster) successfullyRegister(w http.ResponseWriter) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (clust *appCluster) isHeartbeatRequest(r *http.Request) bool {
	if r.Method != http.MethodPut {
		return false
	}

	return caseInsensitiveContains(r.URL.Path, fmt.Sprintf("eureka/apps/%s", clust.ID))
}

func (clust *appCluster) successfulHeartbeat(w http.ResponseWriter) {

	w.WriteHeader(http.StatusOK)
}

func (clust *appCluster) isRequestingInstances(r *http.Request) bool {

	return false
}

func (clust *appCluster) returnInstances(w http.ResponseWriter) {

}

func (clust *appCluster) isDeregistrationRequest(r *http.Request) (bool, string) {
	if r.Method != http.MethodDelete {
		return false, ""
	}

	serviceUrlPath := fmt.Sprintf("eureka/apps/%s/", clust.ID)

	if caseInsensitiveContains(r.URL.Path, serviceUrlPath) {
		return true, strings.Split(r.URL.Path, serviceUrlPath)[1]
	}

	return false, ""
}

func (clust *appCluster) successfullyDeregister(w http.ResponseWriter) {

	w.WriteHeader(http.StatusNotFound)
}
