package fake

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	eureka2 "github.com/newestuser/eureka-proxy/lib/eureka"
	"github.com/newestuser/eureka-proxy/lib/httputil"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func RequestHandler(fakeApps []*Application, pollute bool, chain http.Handler) http.Handler {

	fakes := make(map[string]*appCluster)

	for _, fakeApp := range fakeApps {

		cluster := fakes[fakeApp.ID]

		if cluster == nil {
			cluster = &appCluster{ID: fakeApp.ID}
		}

		cluster.add(fakeApp)

		fakes[fakeApp.ID] = cluster
	}

	return &state{fakeApps: fakes, pollutionOn: pollute, chain: chain}
}

// A representation of the entire fake application configuration
// that will be injected in the original eureka applications list
type state struct {
	fakeApps    map[string]*appCluster
	pollutionOn bool
	chain       http.Handler
}

func (st *state) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if st.isRequestingApps(r) {

		st.respondWithFakes(w, r)
		return
	}

	for _, appCluster := range st.fakeApps {
		if appCluster.isRegistrationRequest(r) {

			appCluster.successfullyRegister(w)
			return
		}

		if appCluster.isHeartbeatRequest(r) {

			appCluster.successfulHeartbeat(w)
			return
		}

		if appCluster.isRequestingInstances(r) {

			appCluster.returnInstances(w)
			return
		}

		if ok, instanceId := appCluster.isDeregistrationRequest(r); ok {

			if ok, instance := appCluster.deregister(instanceId); ok {
				log.Printf("A deregistration request was detected. Deregistering: %s instance: %s\n", appCluster.ID, instance)
			}

			if appCluster.noInstances() {
				st.removeFakeApp(appCluster)
			}

			appCluster.successfullyDeregister(w)
			return
		}
	}

	if st.pollutionOn {
		st.chain.ServeHTTP(w, r)
		return
	}

	if registration, app := st.isRegistrationRequest(r); registration {
		st.injectFakeApp(app)

		st.successfullyRegister(w)
		return
	}

	if heartbeat, app := st.isHeartbeatRequest(r); heartbeat {
		st.injectFakeApp(app)

		st.successfulHeartbeat(w)
		return
	}

	st.chain.ServeHTTP(w, r)
}

func (st *state) isRequestingApps(r *http.Request) bool {

	isGetAppsUrl := strings.HasSuffix(r.URL.Path, "eureka/apps") || strings.HasSuffix(r.URL.Path, "eureka/apps/")

	return r.Method == http.MethodGet && isGetAppsUrl
}

func (st *state) respondWithFakes(w http.ResponseWriter, r *http.Request) {
	rec := httputil.Recorder(w)

	st.chain.ServeHTTP(rec, r)

	state := deserialize(rec)

	for _, appCluster := range st.fakeApps {

		if appExists, existingApp := state.Apps.ContainsApp(appCluster.ID); appExists {

			instances := appCluster.NewInstances()
			existingApp.ReplaceInstances(instances)
		} else {

			state.Apps.AddApp(appCluster.NewEurekaApp())
		}
	}

	appBytes := serialize(rec, state)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		appBytes = httputil.Gzip(appBytes)
	}

	if len(st.fakeApps) > 0 {
		fakeApps := make([]string, 0)

		for _, clust := range st.fakeApps {
			fakeApps = append(fakeApps, clust.ID)
		}

		log.Printf("Will respond with the following fake services:\n\n%s\n\n", strings.Join(fakeApps, "\n"))
	}

	rec.FlushWith(appBytes)

	return
}

func (st *state) isRegistrationRequest(r *http.Request) (bool, *Application) {

	isRegistrationUrl, err := regexp.MatchString(`/eureka/apps/[\w-]+`, r.URL.Path)

	if err != nil {
		panic(err)
	}

	isRegistrationReq := r.Method == http.MethodPost && isRegistrationUrl

	if !isRegistrationReq {
		return false, nil
	}

	instance := readInstance(r)

	app := SingleInstanceApp(instance.App, instance.InstanceID, instance.IPAddress, instance.HostName, instance.Port.Number)

	return true, app
}

func (st *state) isHeartbeatRequest(r *http.Request) (bool, *Application) {

	if r.Method != http.MethodPut {
		return false, nil
	}

	isHeartbeatUrl, err := regexp.MatchString(`/eureka/apps/[\w-]+/.*(\d{4})$`, r.URL.Path)

	if err != nil {
		panic(err)
	}

	if !isHeartbeatUrl {
		return false, nil
	}

	appId, instanceId, port := parseAppIDAndInstanceID(r.URL)

	return true, SingleLocalAppWithInstance(appId, instanceId, port)
}

func (st *state) successfullyRegister(w http.ResponseWriter) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (st *state) successfulHeartbeat(w http.ResponseWriter) {

	w.WriteHeader(http.StatusOK)
}

func (st *state) injectFakeApp(app *Application) {
	log.Printf("A new service was detected. Injecting: %s\n", app)

	clust := st.fakeApps[app.ID]

	if clust == nil {
		clust = &appCluster{ID: app.ID}
	}

	clust.add(app)

	st.fakeApps[app.ID] = clust
}

func (st *state) removeFakeApp(cluster *appCluster) {
	delete(st.fakeApps, cluster.ID)
}

func readInstance(r *http.Request) *eureka2.Instance {

	bytes, e := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if e != nil {
		panic(e)
	}

	req := &eureka2.RegistrationRequest{}

	if err := json.Unmarshal(bytes, req); err != nil {
		panic(err)
	}

	return req.Instance
}

func parseAppIDAndInstanceID(u *url.URL) (string, string, int) {
	parts := strings.Split(u.Path, `/`)

	appId := parts[len(parts)-2]
	instanceId := parts[len(parts)-1]
	port, e := strconv.Atoi(u.Path[len(u.Path)-4:])

	if e != nil {
		panic(e)
	}

	return appId, instanceId, port
}

func deserialize(rec *httputil.HttpResponseRecorder) *eureka2.State {
	contentType := rec.Header().Get("Content-Type")
	s := &eureka2.State{}

	if caseInsensitiveContains(contentType, "application/xml") {
		apps := &eureka2.Applications{}
		err := xml.Unmarshal(rec.Body(), apps)
		if err != nil {
			panic(err)
		}

		s.Apps = apps
		return s

	} else if caseInsensitiveContains(contentType, "application/json") {

		err := json.Unmarshal(rec.Body(), s)

		if err != nil {
			panic(err)
		}

		return s
	}

	panic(fmt.Errorf("could not deserialize body, unknown Content-Type: %s", contentType))
}

func serialize(rec *httputil.HttpResponseRecorder, state *eureka2.State) []byte {
	contentType := rec.Header().Get("Content-Type")

	if caseInsensitiveContains(contentType, "application/xml") {
		bytes, err := xml.Marshal(state.Apps)
		if err != nil {
			panic(fmt.Errorf("could not marshal xml content %s", err.Error()))
		}

		return bytes

	} else if caseInsensitiveContains(contentType, "application/json") {
		bytes, err := json.Marshal(state)
		if err != nil {
			panic(fmt.Errorf("could not marshal json content %s", err.Error()))
		}

		return bytes
	}

	panic(fmt.Errorf("could not serialize unknown Content-Type: %s", contentType))
}

func caseInsensitiveContains(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}
