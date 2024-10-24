package nigonigo

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

var Logger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)
var RequestLogger *log.Logger = nil
var UserAgent = "Mozilla/5.0 Nigonigo/1.0"

func NewHttpClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	return &http.Client{Jar: jar, Transport: &agentSetter{}}, err
}

type agentSetter struct{}

func (t *agentSetter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Origin", httpOrigin)

	// headers for nvapi.nicovideo.jp
	req.Header.Set("x-frontend-id", "6")
	req.Header.Set("x-request-with", "nicovideo")

	if RequestLogger != nil {
		RequestLogger.Println("REQUEST", req.Method, req.URL)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func newPostReq(urlstr string, params map[string]string) (*http.Request, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	req, err := http.NewRequest("POST", urlstr, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func newGetReq(urlstr string, params map[string]string) (*http.Request, error) {
	if len(params) == 0 {
		return http.NewRequest("GET", urlstr, nil)
	}
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return http.NewRequest("GET", urlstr+"?"+values.Encode(), nil)
}

func getContent(client *http.Client, url string, params map[string]string) ([]byte, error) {
	req, err := newGetReq(url, params)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		return nil, AuthenticationRequired
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code :%v", res.StatusCode)
	}
	if res.Request.Response != nil && strings.Contains(res.Request.URL.String(), "/account.nicovideo.jp/") {
		return nil, AuthenticationRequired
	}

	return ioutil.ReadAll(res.Body)
}

func doRequest(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, err
}
