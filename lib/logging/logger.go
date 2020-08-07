package logging

import (
	"bytes"
	"fmt"
	"github.com/newestuser/eureka-proxy/lib/httputil"
	"io/ioutil"
	"net/http"
	"strings"
)

type Logger interface {
	Separator()

	Trace(vals ...interface{})

	TraceF(format string, vals ...interface{}) string

	Info(vals ...interface{})

	InfoF(format string, vals ...interface{}) string

	Err(vals ...interface{})

	ErrF(format string, vals ...interface{}) error
}

func NewHandler(l Logger, chain http.Handler) http.Handler {

	return &logHandler{l: l, chain: chain}
}

type logHandler struct {
	l     Logger
	chain http.Handler
}

func (h *logHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.l.Separator()

	h.l.InfoF("REQUEST: %s %s\n", r.Method, r.URL.Path)
	h.l.TraceF("HEADERS: %s\n", prettyHeaders(r.Header))
	h.l.TraceF("BODY: \n%s\n", copyBody(r))

	recorder := httputil.Recorder(w)
	h.chain.ServeHTTP(recorder, r)

	respBody := recorder.BodyString();
	if !strings.HasPrefix(recorder.Header().Get("Content-Type"), "application") {
		respBody = "SOME-BYTES"
	}

	h.l.InfoF("RESPONSE StatusCode: %d\n", recorder.Status())
	h.l.TraceF("HEADERS: %s\n", prettyHeaders(recorder.Header()))
	h.l.TraceF("BODY: \n%s\n", respBody)

	recorder.Flush()
}

func copyBody(r *http.Request) string {
	bodyBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(fmt.Errorf("error reading body err:%v", err.Error()))
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return string(bodyBytes)
}

func prettyHeaders(headers http.Header) string {
	buff := bytes.NewBufferString("\n")

	for k, v := range headers {
		buff.WriteString(fmt.Sprintf("%s : %s\n", k, v))
	}

	return buff.String()
}
