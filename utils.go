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
	return &http.Client{Jar: jar, Transport: &AgentSetter{}}, err
}

type AgentSetter struct{}

func (t *AgentSetter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Origin", httpOrigin)
	if RequestLogger != nil {
		RequestLogger.Println("REQUEST", req.Method, req.URL)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func NewPostReq(urlstr string, params map[string]string) (*http.Request, error) {
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

func NewGetReq(urlstr string, params map[string]string) (*http.Request, error) {
	if len(params) == 0 {
		return http.NewRequest("GET", urlstr, nil)
	}
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return http.NewRequest("GET", urlstr+"?"+values.Encode(), nil)
}

func GetContent(client *http.Client, url string, params map[string]string) ([]byte, error) {
	req, err := NewGetReq(url, params)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("incalid status code :%v", res.StatusCode)
	}
	if res.Request.Response != nil && strings.Contains(res.Request.URL.String(), "/account.nicovideo.jp/") {
		return nil, AuthenticationRequired
	}

	return ioutil.ReadAll(res.Body)
}

func DoRequest(client *http.Client, req *http.Request) ([]byte, error) {
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
