package nigonigo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	loginApiUrl  = "https://account.nicovideo.jp/api/v1/login?site=niconico"
	searchApiUrl = "http://api.search.nicovideo.jp/api/v2/video/contents/search"
	topUrl       = "https://www.nicovideo.jp/"
	watchUrl     = "https://www.nicovideo.jp/watch/"
	httpOrigin   = "https://www.nicovideo.jp"
	nvApiUrl     = "https://nvapi.nicovideo.jp/v1/"
)

type Client struct {
	HttpClient *http.Client
	Session    *NicoSession
}

type AccountConfig struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type NicoSession struct {
	Id            int64
	SessionString string
	IsPremium     bool
}

func NewClient() *Client {
	client, err := NewHttpClient()
	if err != nil {
		panic(err)
	}
	return &Client{client, nil}
}

func (n *Client) LoginWithJsonFile(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var c AccountConfig
	err = json.Unmarshal(buf, &c)
	if err != nil {
		return err
	}
	return n.Login(&c)
}

func (c *Client) Login(ac *AccountConfig) error {
	params := map[string]string{
		"mail_tel": ac.Id,
		"password": ac.Password,
	}
	_, err := c.post(loginApiUrl, params)
	if err != nil {
		return err
	}
	// TODO
	return nil
}

func (c *Client) GetContent(url string) ([]byte, error) {
	return GetContent(c.HttpClient, url)
}

func (c *Client) getWithParams(urlstr string, params map[string]string) (string, error) {
	req, err := NewGetReq(urlstr, params)
	if err != nil {
		return "", err
	}
	return c.request(req)
}

func (a *Client) get(apiurl string) (string, error) {
	req, err := http.NewRequest("GET", apiurl, nil)
	if err != nil {
		return "", err
	}
	return a.request(req)
}

func (c *Client) post(urlstr string, params map[string]string) (string, error) {
	req, err := NewPostReq(urlstr, params)
	if err != nil {
		return "", err
	}
	return c.request(req)
}

func (c *Client) request(req *http.Request) (string, error) {
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	body := string(b)
	return body, err
}
