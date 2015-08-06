package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ApiTiming struct {
	request  Request
	prefix   string
	username string
	pass     string
}

func NewApiTiming(request Request, prefix string, user string, pass string) *ApiTiming {
	return &ApiTiming{
		request:  request,
		prefix:   prefix,
		username: user,
		pass:     pass,
	}
}

func (a *ApiTiming) GetName() string {
	return a.prefix + a.request.UniqeName()
}

func (a *ApiTiming) GetUnits() string {
	return "ms"
}

func (a *ApiTiming) GetValue() (diff float64, err error) {
	t0 := time.Now().UnixNano() / int64(time.Millisecond)

	request, err := a.request.Make()
	if err != nil {
		return 0, err
	}

	request.SetBasicAuth(a.username, a.pass)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(b))

	if resp.StatusCode != a.request.SuccessHttpCode() {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, errors.New("Unexpected http response code:" + strconv.Itoa(resp.StatusCode) + "with message:" + string(body))
	}

	t1 := time.Now().UnixNano() / int64(time.Millisecond)

	return float64(t1 - t0), nil
}

type Request struct {
	Name     string
	Uri      string
	HttpCode int
	Method   string
	Payload  json.RawMessage
}

func (r *Request) UniqeName() string {
	return r.Name
}

func (r *Request) Make() (*http.Request, error) {
	payload := strings.NewReader(string(r.Payload))
	return http.NewRequest(r.Method, r.Uri, payload)
}

func (r *Request) SuccessHttpCode() int {
	if r.HttpCode == 0 {
		return 200
	}
	return r.HttpCode
}
