package httputil

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Create a new http response recorder that can be used to copy the response, status code and headers.
func Recorder(w http.ResponseWriter) *HttpResponseRecorder {
	return &HttpResponseRecorder{w: w, header: make(http.Header)}
}

// HttpResponseRecorder is wrapper of http.ResponseWriter
type HttpResponseRecorder struct {
	http.ResponseWriter

	w      http.ResponseWriter
	buff   bytes.Buffer
	header http.Header
	status int
}

func (rec *HttpResponseRecorder) Header() http.Header {
	return rec.w.Header()
}

func (rec *HttpResponseRecorder) Write(b []byte) (int, error) {
	return rec.buff.Write(b)
}

func (rec *HttpResponseRecorder) WriteHeader(s int) {
	rec.status = s
}

// Write the copied body to the original http.ResponseWriter.
func (rec *HttpResponseRecorder) Flush() {
	rec.FlushWith(rec.buff.Bytes())
}

// Write the provided bytes to the original http.ResponseWriter.
func (rec *HttpResponseRecorder) FlushWith(bytes []byte) {
	rec.w.Header().Del("Content-Length") // remove this header in order to write a proper content-length value
	rec.w.WriteHeader(rec.status)
	rec.w.Write(bytes)
}

// A convenient method for extracting the recorded bytes.
// Note that if the content is encoded it will be decoded.
func (rec *HttpResponseRecorder) Body() []byte {
	return rec.gunzipBytes()
}

// A convenient method for extracting the recorded bytes in a string format.
// Note that if the content is encoded it will be decoded.
func (rec *HttpResponseRecorder) BodyString() string {
	if rec.buff.Len() == 0 {
		return ""
	}

	return string(rec.gunzipBytes())
}

// Return the recorded status.
func (rec *HttpResponseRecorder) Status() int {
	return rec.status
}

func (rec *HttpResponseRecorder) gunzipBytes() []byte {

	if rec.Header().Get("Content-Encoding") == "gzip" {

		return Gunzip(rec.buff.Bytes())
	}

	return rec.buff.Bytes()
}

// Extract the gzip bytes.
func Gunzip(v []byte) []byte {
	reader, err := gzip.NewReader(bytes.NewReader(v))
	if err != nil {
		panic(fmt.Errorf("could not create gzip reader err: %s", err.Error()))
	}

	unziped, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("could not read gzip content: %s", err.Error()))
	}

	return unziped
}

// Gzip the provided bytes.
func Gzip(v []byte) []byte {
	buff := &bytes.Buffer{}

	w := gzip.NewWriter(buff)

	_, err := w.Write(v)

	if err != nil {
		panic(fmt.Errorf("could not write gzip content err: %s", err.Error()))
	}

	if err := w.Close(); err != nil {
		panic(fmt.Errorf("could not close gzip writer err: %s", err.Error()))
	}

	return buff.Bytes()
}
