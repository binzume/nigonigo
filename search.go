package nigonigo

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SearchResultItem struct {
	ContentID    string `json:"contentID"`
	Title        string `json:"title"`
	Description  int    `json:"description"`
	Duration     int    `json:"lengthSeconds"`
	UserID       int    `json:"userId"`
	ChannelID    int    `json:"channelId"`
	Tags         string `json:"tags"`
	StartTime    string `json:"startTime"`
	ThumbnailURL string `json:"thumbnailUrl"`
	ViewCount    int    `json:"viewCounter"`
	MylistCount  int    `json:"mylistCounter"`
	CommentCount int    `json:"commentCounter"`
}

// TODO
var DefaultFields = []string{"contentId", "title", "channelId", "userId", "tags", "startTime"}

type SearchResult struct {
	TotalCount int
	Offset     int
	Items      []SearchResultItem
}

// TODO
type SearchFilter interface{}

type SimpleFilter struct {
	Type  string `json:"type"`
	Field string `json:"field"`
	Value string `json:"value"`
}
type RangeFilter struct {
	Type string `json:"type"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type FilterGroup struct {
	Type    string         `json:"type"`
	Filters []SearchFilter `json:"filters"`
}

func AndFilter(filters []SearchFilter) SearchFilter {
	return &FilterGroup{"and", filters}
}

func OrFilter(filters []SearchFilter) SearchFilter {
	return &FilterGroup{"or", filters}
}

func NotFilter(filter SearchFilter) SearchFilter {
	return map[string]interface{}{"type": "not", "filter": filter}
}

func EqualFilter(field, value string) SearchFilter {
	return &SimpleFilter{"equal", field, value}
}

func (c *Client) SearchByTag(tag string, offset, limit int) (*SearchResult, error) {
	return c.SearchVideo(tag, "tagsExact,categoryTags", "-startTime", offset, limit, nil)
}

func (c *Client) SearchByKeyword(s string, offset, limit int) (*SearchResult, error) {
	return c.SearchVideo(s, "description,title", "-startTime", offset, limit, nil)
}

func (c *Client) SearchByChannel(channelID string, offset, limit int) (*SearchResult, error) {
	filter := EqualFilter("channelId", channelID)
	q := "アニメ OR ゲーム OR 料理" // TODO
	return c.SearchVideo(q, "categoryTags", "-startTime", offset, limit, filter)
}

func (c *Client) SearchVideo(q, targets, sort string, offset, limit int, filter SearchFilter) (*SearchResult, error) {
	params := map[string]string{
		"q":        q,
		"targets":  targets,
		"fields":   strings.Join(DefaultFields, ","),
		"_sort":    sort,
		"_offset":  fmt.Sprint(offset),
		"_limit":   fmt.Sprint(limit),
		"_context": "nigonigo",
	}
	if filter != nil {
		encoded, err := json.Marshal(filter)
		if err != nil {
			return nil, err
		}
		params["jsonFilter"] = string(encoded)
	}

	body, err := c.getWithParams(searchApiUrl, params)
	if err != nil {
		return nil, err
	}

	type searchResponse struct {
		Meta struct {
			Status       int    `json:"status"`
			TotalCount   int    `json:"totalCount"`
			ID           string `json:"id"`
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"meta"`
		Data []SearchResultItem `json:"data"`
	}
	var res searchResponse
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return nil, err
	}
	if res.Meta.Status != 200 {
		return nil, fmt.Errorf("%v code:%v", res.Meta.ErrorMessage, res.Meta.ErrorCode)
	}

	return &SearchResult{TotalCount: res.Meta.TotalCount, Offset: offset, Items: res.Data}, nil
}
