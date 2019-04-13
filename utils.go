package nigonigo

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const UserAgent = "Mozilla/5.0 Nigonigo/1.0"

func NewHttpClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	return &http.Client{Jar: jar, Transport: &AgentSetter{}}, err
}

type AgentSetter struct{}

func (t *AgentSetter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Origin", httpOrigin)
	log.Println("REQUEST", req.Method, req.URL)
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

func GetContent(client *http.Client, url string) ([]byte, error) {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
