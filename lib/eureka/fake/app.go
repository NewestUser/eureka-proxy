package fake

import (
	"fmt"
	eureka2 "github.com/newestuser/eureka-proxy/lib/eureka"
	"strings"
)

// A representation of a single application with multiple instances
type Application struct {
	ID      string
	targets map[string] *Target
}

// A representation of a single application instance
type Target struct {
	InstanceID string
	Host       string
	Port       int
	IP         string
}

func (t *Target) String() string {
	return fmt.Sprintf("Instance{id=%s, host=%s, port=%d, ip=%s}", t.InstanceID, t.Host, t.Port, t.IP)
}

// Create a single app with one instance
func SingleInstanceApp(appID, instanceID, IP, host string, port int) *Application {

	app := &Application{ID: appID}

	app.AddInstance(&Target{InstanceID: instanceID, Host: host, Port: port, IP: IP})

	return app
}

func (app *Application) String() string {

	return fmt.Sprintf("FakeApp{id=%s, instances=%s}", app.ID, app.Instances())
}

func (app *Application) NewEurekaApp() *eureka2.Application {
	return &eureka2.Application{
		Name:      strings.ToUpper(app.ID),
		Instances: app.NewInstances(),
	}
}

func (app *Application) NewInstances() []*eureka2.Instance {
	inst := make([]*eureka2.Instance, 0)

	for _, target := range app.targets {
		inst = append(inst, eureka2.NewInstance(app.ID, target.IP, target.Host, target.Port))
	}

	return inst
}

// Remove a single instance from the application. Return true if the successfully removed, false otherwise.
func (app *Application) RemoveInstance(instanceID string) (bool, *Target) {
	app.initIfNil()

	normInstanceId := strings.ToLower(instanceID)

	if target, ok := app.targets[normInstanceId]; ok {
		delete(app.targets, normInstanceId)
		return true, target
	}

	return false, nil
}

// Check if there are any instances left in the application.
func (app *Application) NoInstances() bool {
	app.initIfNil()

	return len(app.targets) == 0
}

// Add a single instance to this application.
func (app *Application) AddInstance(t *Target) {
	app.initIfNil()

	app.targets[strings.ToLower(t.InstanceID)] = t
}

// Add multiple instances to this application.
func (app *Application) AddInstances(targets []*Target) {
	for _, t := range targets {
		app.AddInstance(t)
	}
}

// Retrieve all the instances of the application.
func (app *Application) Instances() []*Target {
	app.initIfNil()

	targets := make([]*Target, len(app.targets))

	i := 0
	for _, v := range app.targets {
		targets[i] = v
	}

	return targets
}

func (app *Application) initIfNil() {
	if app.targets == nil {
		app.targets = make(map[string]*Target)
	}
}
