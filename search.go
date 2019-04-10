package nigonigo

import (
	"encoding/json"
	"fmt"
)

type SearchResultItem struct {
	ContentID    string `json:"contentID"`
	Title        string `json:"title"`
	UserID       int    `json:"userId"`
	ChannelID    int    `json:"channelId"`
	Tags         string `json:"tags"`
	StartTime    string `json:"startTime"`
	ThumbnailURL string `json:"thumbnailUrl"`
	ViewCounter  int    `json:"viewCounter"`
}

type SearchResult struct {
	TotalCount int
	Offset     int
	Items      []SearchResultItem
}

type searchResponse struct {
	Meta struct {
		Status     int    `json:"status"`
		TotalCount int    `json:"totalCount"`
		ID         string `json:"id"`
	} `json:"meta"`
	Data []SearchResultItem `json:"data"`
}

func (c *Client) SearchByTag(tag string, offset, limit int, filter interface{}) (*SearchResult, error) {
	params := map[string]string{
		"q":        tag,
		"targets":  "tags,categoryTags",
		"fields":   "contentId,title,channelId,userId,tags,startTime",
		"_sort":    "-startTime",
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

	var res searchResponse
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return nil, err
	}

	return &SearchResult{TotalCount: res.Meta.TotalCount, Offset: offset, Items: res.Data}, nil
}
