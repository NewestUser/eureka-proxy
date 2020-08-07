package eureka

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Status string

const (
	UP      = "UP"
	DOWN    = "DOWN"
	UNKNOWN = "UNKNOWN"
)

type State struct {
	Apps *Applications `json:"applications"`
}

type Applications struct {
	XMLName      xml.Name       `json:"-" xml:"applications"`
	Version      interface{}    `json:"versions__delta" xml:"versions__delta"`
	Status       string         `json:"apps__hashcode" xml:"apps__hashcode"`
	Applications []*Application `json:"application" xml:"application"`
}

func (a *Applications) ContainsApp(ID string) (bool, *Application) {

	for _, app := range a.Applications {
		if strings.EqualFold(ID, app.Name) {
			return true, app
		}
	}

	return false, nil
}

func (a *Applications) AddApp(app ...*Application) {
	a.Applications = append(a.Applications, app...)
}

type Application struct {
	XMLName   xml.Name    `json:"-" xml:"application"`
	Name      string      `json:"name" xml:"name"`
	Instances []*Instance `json:"instance" xml:"instance"`
}

func (app *Application) ReplaceInstances(newInstances []*Instance) {
	app.Instances = newInstances
}

//---------------------------------------------------------------------

// A request that is send when a service should be registered in eureka.
type RegistrationRequest struct {
	Instance *Instance `json:"instance" xml:"instance"`
}

// Information about a single instance.
type Instance struct {
	InstanceID                    string      `json:"instanceId" xml:"instanceId"`
	HostName                      string      `json:"hostName" xml:"hostName"`
	App                           string      `json:"app" xml:"app"`
	IPAddress                     string      `json:"ipAddr" xml:"ipAddr"`
	Status                        Status      `json:"status" xml:"status"`
	OverriddenStatus              Status      `json:"overriddenstatus" xml:"overriddenstatus"`
	Port                          *Port       `json:"port,omitempty" xml:"port,omitempty"`
	SecurePort                    *Port       `json:"securePort,omitempty" xml:"securePort,omitempty"`
	CountryID                     int         `json:"countryId" xml:"countryId"`
	DataCenter                    *DataCenter `json:"dataCenterInfo" xml:"dataCenterInfo"`
	Lease                         *Lease      `json:"leaseInfo" xml:"leaseInfo"`
	MetaData                      *MetaData   `json:"metadata" xml:"metadata"`
	HomePageURL                   string      `json:"homePageUrl" xml:"homePageUrl"`
	StatusPageURL                 string      `json:"statusPageUrl" xml:"statusPageUrl"`
	HealthCheckURL                string      `json:"healthCheckUrl" xml:"healthCheckUrl"`
	VIPAddress                    string      `json:"vipAddress" xml:"vipAddress"`
	SecureVIPAddress              string      `json:"secureVipAddress" xml:"secureVipAddress"`
	IsCoordinatingDiscoveryServer string      `json:"isCoordinatingDiscoveryServer" xml:"isCoordinatingDiscoveryServer"`
	LastUpdatedTimestamp          string      `json:"lastUpdatedTimestamp" xml:"lastUpdatedTimestamp"`
	LastDirtyTimestamp            string      `json:"lastDirtyTimestamp" xml:"lastDirtyTimestamp"`
	ActionType                    string      `json:"actionType" xml:"actionType"`
}

func NewInstance(id, ip, hostName string, port int) *Instance {

	return &Instance{
		InstanceID:                    fmt.Sprintf("%s:%s:%d", strings.ToLower(hostName), strings.ToLower(id), port),
		HostName:                      ip,
		App:                           strings.ToUpper(id),
		IPAddress:                     ip,
		Status:                        UP,
		OverriddenStatus:              UNKNOWN,
		Port:                          NewPort(port),
		SecurePort:                    SecurePort(),
		CountryID:                     1,
		DataCenter:                    DefaultDataCenter(),
		Lease:                         DefaultLease(),
		MetaData:                      NewMetaData(strings.ToLower(id), port),
		HomePageURL:                   fmt.Sprintf("http://%s:%d/", ip, port),
		StatusPageURL:                 fmt.Sprintf("http://%s:%d/admin/manage/info", ip, port),
		HealthCheckURL:                fmt.Sprintf("http://%s:%d/admin/manage/health", ip, port),
		VIPAddress:                    fmt.Sprintf("%s", strings.ToLower(id)),
		SecureVIPAddress:              fmt.Sprintf("%s", strings.ToLower(id)),
		IsCoordinatingDiscoveryServer: "false",
		LastUpdatedTimestamp:          "1517243533603",
		LastDirtyTimestamp:            "1513015393398",
		ActionType:                    "ADDED",
	}
}

type Port struct {
	Number  int    `json:"$" xml:",chardata"`
	Enabled string `json:"@enabled" xml:"enabled,attr"`
}

func NewPort(number int) *Port {
	return &Port{Number: number, Enabled: "true"}
}

func SecurePort() *Port {
	return &Port{Number: 443, Enabled: "false"}
}

type DataCenter struct {
	Class string `json:"@class" xml:"class,attr"`
	Name  string `json:"name" xml:"name"`
}

func DefaultDataCenter() *DataCenter {
	return &DataCenter{
		Class: `com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo`,
		Name:  `MyOwn`,
	}
}

type Lease struct {
	RenewalIntervalInSecs int
	DurationInSecs        int
	RegistrationTimestamp int
	LastRenewalTimestamp  int
	EvictionTimestamp     int
	ServiceUpTimestamp    int
}

func DefaultLease() *Lease {
	return &Lease{
		RenewalIntervalInSecs: 30,
		DurationInSecs:        90,
		RegistrationTimestamp: 1519411412763,
		LastRenewalTimestamp:  1519747384239,
		EvictionTimestamp:     0,
		ServiceUpTimestamp:    1519411412763,
	}
}

type MetaData struct {
	InstanceID string `json:"instanceId" xml:"instanceId"`
}

func NewMetaData(ID string, port int) *MetaData {
	return &MetaData{
		InstanceID: fmt.Sprintf("%s:%d", ID, port),
	}
}
