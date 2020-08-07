package testutil

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func AssertUrl(t *testing.T, r *http.Request, want string) {
	assert.Equal(t, want, r.URL.Path)
}

func AssertReq(t *testing.T, r *http.Request, want interface{}, holder interface{}) {
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		panic("cant read body err: " + readErr.Error())
	}

	err := json.Unmarshal(body, holder)
	if err != nil {
		panic("cant unmarshal response err: " + err.Error())
	}

	assert.Equal(t, holder, want)
}

func WriteResp(w http.ResponseWriter, resp interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	b, _ := json.Marshal(resp)
	w.Write(b)
}
