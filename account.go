package nigonigo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type AccountConfig struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type NicoSession struct {
	NiconicoID    string `json:"niconicoId"`
	IsPremium     bool   `json:"premium"`
	SessionString string `json:"user_session"`
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
	return c.LoginWithPassword(ac.ID, ac.Password)
}

func (c *Client) LoginWithPassword(id, password string) error {
	params := map[string]string{
		"mail_tel": id,
		"password": password,
	}
	req, err := NewPostReq(loginApiUrl, params)
	if err != nil {
		return err
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if strings.Contains(res.Request.URL.String(), "/account.nicovideo.jp/") {
		return errors.New("login error")
	}

	cookie, err := res.Request.Cookie("user_session")
	if err != nil {
		return err
	}

	c.Session = &NicoSession{
		NiconicoID:    res.Header.Get("X-Niconico-Id"),
		IsPremium:     res.Header.Get("X-Niconico-Authflag") == "3",
		SessionString: cookie.Value,
	}
	return err
}

func (c *Client) SaveLoginSession(path string) error {
	json, err := json.Marshal(c.Session)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) LoadLoginSession(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var s NicoSession
	err = json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	return c.SetSessionString(s.SessionString)
}

func (c *Client) SetSessionString(sessionStr string) error {
	cookie := &http.Cookie{Name: "user_session", Value: sessionStr}
	parsed, _ := url.Parse(topUrl)
	c.HttpClient.Jar.SetCookies(parsed, []*http.Cookie{cookie})

	req, err := NewGetReq(topUrl, nil)
	if err != nil {
		return err
	}
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.Header.Get("X-Niconico-Id") == "" {
		return errors.New("expired")
	}

	c.Session = &NicoSession{
		NiconicoID:    res.Header.Get("X-Niconico-Id"),
		IsPremium:     res.Header.Get("X-Niconico-Authflag") == "3",
		SessionString: sessionStr,
	}
	return nil
}
