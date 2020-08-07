package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Logger interface {
	Separator()

	TraceF(format string, vals ...interface{}) string
}

type Executor interface {
	Post(url string, req, resp interface{}) error

	Put(url string, req, resp interface{}) error

	NewReq(method string, url string, req interface{}) (*http.Request, error)

	Execute(req *http.Request, resp interface{}) (*http.Response, error)
}

func NewExecutor(Logger Logger) Executor {
	return &jsonExecutor{logger: Logger}
}

type jsonExecutor struct {
	logger Logger
}

func (e *jsonExecutor) NewReq(method string, url string, reqBody interface{}) (*http.Request, error) {

	reqBytes, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return &http.Request{}, fmt.Errorf("cant marshal request err: %v\n", marshalErr)
	}

	httpReq, reqErr := http.NewRequest(method, url, bytes.NewReader(reqBytes))

	if reqErr != nil {
		return &http.Request{}, fmt.Errorf("cant construct req err: %v\n", reqErr)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	return httpReq, nil
}

func (e *jsonExecutor) Post(url string, req interface{}, resp interface{}) error {
	e.logger.Separator()

	httpResp, postErr := e.execReq(url, http.MethodPost, req)
	if postErr != nil {
		return fmt.Errorf("%s %s sending req err: %v\n", "POST", url, postErr)
	}

	if httpResp.StatusCode != http.StatusCreated {
		return formatErrFromResponse(url, "POST", httpResp)
	}

	return e.extractResp(httpResp, resp)
}

func (e *jsonExecutor) Put(url string, req interface{}, resp interface{}) error {
	e.logger.Separator()

	httpResp, putErr := e.execReq(url, http.MethodPut, req)
	if putErr != nil {
		return fmt.Errorf("error sending req Method: %s URL: %s err: %v\n", url, "PUT", putErr)
	}

	if httpResp.StatusCode == http.StatusOK || httpResp.StatusCode == http.StatusCreated {
		return e.extractResp(httpResp, resp)
	}

	return formatErrFromResponse(url, "PUT", httpResp)
}

func (e *jsonExecutor) Execute(req *http.Request, resp interface{}) (*http.Response, error) {
	e.logger.Separator()

	buf, _ := ioutil.ReadAll(req.Body)
	firstBodyCopy := ioutil.NopCloser(bytes.NewBuffer(buf))
	secondBodyCopy := ioutil.NopCloser(bytes.NewBuffer(buf))

	req.Body = secondBodyCopy

	reqBytes, marshalErr := ioutil.ReadAll(firstBodyCopy)
	if marshalErr != nil {
		return nil, e.execErr(req, fmt.Sprintf("cant marshal request err: %v\n", marshalErr))
	}

	e.logger.TraceF("Request %s: %s Body: %s\n", req.Method, req.URL.String(), prettyJson(reqBytes))
	httpResp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return nil, e.execErr(req, fmt.Sprintf("could not execute request err: %s\n", respErr))
	}

	if err := e.extractResp(httpResp, resp); err != nil {
		return nil, err
	}

	return httpResp, nil
}

func (e *jsonExecutor) execReq(url, method string, reqBody interface{}) (*http.Response, error) {
	reqBytes, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return nil, fmt.Errorf("%s %s cant marshal request err: %s\n", method, url, marshalErr)
	}

	e.logger.TraceF("Request %s Body: %s\n", url, prettyJson(reqBytes))
	httpReq, reqEr := http.NewRequest(method, url, bytes.NewReader(reqBytes))
	if reqEr != nil {
		return nil, fmt.Errorf("%s %s cant construct req err: %s\n", url, method, reqEr)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, respErr := http.DefaultClient.Do(httpReq)
	if respErr != nil {
		return nil, e.execErr(httpReq, fmt.Sprintf("could not execute request: %s\n", respErr))
	}

	return httpResp, nil
}

func (e *jsonExecutor) extractResp(httpResp *http.Response, holder interface{}) error {
	defer httpResp.Body.Close()

	body, readErr := ioutil.ReadAll(httpResp.Body)
	if readErr != nil {
		return fmt.Errorf("cant read body err: %s\n", readErr)
	}

	e.logger.TraceF("Response Status: %d Body: %s\n", httpResp.StatusCode, prettyJson(body))

	if holder == nil {
		return nil
	}

	unmarshalErr := json.Unmarshal(body, holder)
	if unmarshalErr != nil {
		return fmt.Errorf("cant unmarshal response err: %s\n", unmarshalErr)
	}

	return nil
}

func (e *jsonExecutor) execErr(req *http.Request, msg string) error {
	return fmt.Errorf("Request %s: %s %s\n", req.Method, req.URL, msg)
}

func formatErrFromResponse(url, method string, resp *http.Response) error {
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		panic(fmt.Errorf("cant read body err: %s\n", readErr))

	}

	return fmt.Errorf("%s %s %s\n", method, url, prettyJson(body))
}

func prettyJson(body []byte) string {
	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, body, "", "\t")

	return string(prettyJSON.Bytes())
}