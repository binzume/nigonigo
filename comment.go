package nigonigo

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Comment struct {
	ID          string   `json:"id"`
	No          int      `json:"no"`
	VposMs      int      `json:"vposMs"`
	Body        string   `json:"body"`
	Commands    []string `json:"commands"`
	UserID      string   `json:"userId"`
	IsPremium   bool     `json:"isPremium"`
	Score       int      `json:"score"`
	PostedAt    string   `json:"postedAt"`
	NicoruCount int      `json:"nicoruCount"`
	NicoruId    any      `json:"nicoruId"`
	Source      string   `json:"source"`
	IsMyPost    bool     `json:"isMyPost"`
}

type CommentThread struct {
	ID           string     `json:"id"`
	Fork         string     `json:"fork"`
	CommentCount int        `json:"commentCount"`
	Comments     []*Comment `json:"comments"`
}

func (c *Client) GetComments(server, threadKey string, params map[string]any) ([]*CommentThread, error) {
	url := server + "/v1/threads"

	reqData := struct {
		Params      map[string]any `json:"params"`
		ThreadKey   string         `json:"threadKey"`
		Additionals map[string]any `json:"additionals"`
	}{
		Params:      params,
		ThreadKey:   threadKey,
		Additionals: map[string]any{},
	}

	reqJson, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := doRequest(c.HttpClient, req)
	if err != nil {
		return nil, err
	}

	res := struct {
		Meta map[string]any `json:"meta"`
		Data struct {
			GlobalComments []map[string]any `json:"globalComments"`
			Threads        []*CommentThread `json:"threads"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return res.Data.Threads, nil
}
