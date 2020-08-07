package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newestuser/eureka-proxy/lib/httputil/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeLogger struct {
	mock.Mock
}

func (f *fakeLogger) Separator() {
	f.Called()
}

func (f *fakeLogger) TraceF(format string, vals ...interface{}) string {
	f.Called()
	return fmt.Sprintf(format, vals...)
}

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestExecutePostRequest(t *testing.T) {
	wantReq := &person{Name: "foo", Age: 20}
	wantResp := &person{Name: "bar", Age: 40}

	logger := new(fakeLogger)
	logger.On("Separator").Once()
	logger.On("TraceF").Twice()

	ts := httptest.NewServer(withReqResp(t, wantReq, wantResp))
	defer ts.Close()

	executor := jsonExecutor{logger: logger}
	err := executor.Post(ts.URL, wantReq, &person{})

	assert.Nil(t, err)
	logger.AssertExpectations(t)
}

func TestExecutePutRequest(t *testing.T) {
	wantReq := &person{Name: "aaa", Age: 111}
	wantResp := &person{Name: "bbb", Age: 222}

	logger := new(fakeLogger)
	logger.On("Separator").Once()
	logger.On("TraceF").Twice()

	ts := httptest.NewServer(withReqResp(t, wantReq, wantResp))
	defer ts.Close()

	executor := jsonExecutor{logger: logger}
	err := executor.Put(ts.URL, wantReq, &person{})

	assert.Nil(t, err)
	logger.AssertExpectations(t)
}

func withReqResp(t *testing.T, req, resp *person) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertReq(t, r, req, &person{})
		testutil.AssertUrl(t, r, "/")
		testutil.WriteResp(w, resp, http.StatusCreated)
	}
}
