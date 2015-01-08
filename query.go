package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	AdminUserName = "admin"
	AdminPassword = "admin"
	BaseUrl       = "http://localhost"
)

type QueryApi struct {
	Resource Query
	User     string
	Password string
	BaseUrl  string
}

type Query struct {
	Query           string   `json:"query"`
	PaginationSkip  int      `json:"skip,omitempty"`
	PaginationLimit int      `json:"limit,omitempty"`
	ContextInclude  []string `json:"hostContextInclude,omitempty"`
	ContextExclude  []string `json:"hostContextExclude,omitempty"`
	SortDescending  bool     `json:"sortDescending,omitempty"`
	SortColumn      string   `json:"sortColumn,omitempty"`
}

func (q *QueryApi) RequestTest() (int64, error) {

	t0 := time.Now().UnixNano() / int64(time.Millisecond)
	resp, err := q.sendRequest(http.Client{})
	if err != nil {
		return 0, err
	}

	resp.Body.Close()

	t1 := time.Now().UnixNano() / int64(time.Millisecond)

	diff := t1 - t0

	return diff, nil
}

func (q *Query) ExpectedHttpCode() int {
	return 200
}

func (q *QueryApi) sendRequest(client http.Client) (*http.Response, error) {

	req, err := q.compileReqest()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != q.Resource.ExpectedHttpCode() {
		body, _ := ioutil.ReadAll(resp.Body)
		error_messgae := errors.New("Unexpected http response code:" + strconv.Itoa(resp.StatusCode) + "with message:" + string(body))
		resp.Body.Close()
		return nil, error_messgae
	}

	return resp, nil
}

func (q *QueryApi) GetResponse() ([]byte, error) {

	resp, err := q.sendRequest(http.Client{})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (q *QueryApi) getUri() string {
	return q.BaseUrl + q.Resource.GetResourceUri()
}

func (q *QueryApi) compileReqest() (*http.Request, error) {

	payload, err := q.Resource.GetPayload()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("post",
		q.getUri(),
		strings.NewReader(payload))

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(q.User, q.Password)
	req.Header.Add("Accept-Encoding", "identity")

	return req, nil
}

func (q *Query) GetResourceUri() string {
	return "/api/query"
}

func (q *Query) GetPayload() (string, error) {

	payload, err := json.Marshal(q)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
