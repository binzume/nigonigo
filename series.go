package nigonigo

import (
	"encoding/json"
	"strconv"
	"strings"
)

type SeriesItem struct {
	Meta  map[string]any `json:"meta"`
	Video *struct {
		Type         string `json:"type"`
		ID           string `json:"id"`
		Title        string `json:"title"`
		RegisteredAt string `json:"registeredAt"`
		Count        struct {
			View    int `json:"view"`
			Mylist  int `json:"mylist"`
			Comment int `json:"comment"`
		} `json:"count"`
		Thumbnail struct {
			Url string `json:"url"`
		} `json:"thumbnail"`
		Duration         int    `json:"duration"`
		ShortDescription string `json:"shortDescription"`
		Owner            struct {
			Type string `json:"type"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"owner"`
	}
}
type SeriesResponse struct {
	Meta map[string]any `json:"meta"`
	Data struct {
		TotalCount int          `json:"totalCount"`
		Items      []SeriesItem `json:"items"`
	} `json:"data"`
}

func (c *Client) FindSeriesVideos(sid string) (*SearchResult, error) {
	body, err := getContent(c.HttpClient, nvApiV2Url+"series/"+sid+"?pageSize=500&page=1", nil)
	if err != nil {
		return nil, err
	}
	res := SeriesResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	items := []*SearchResultItem{}
	for _, item := range res.Data.Items {
		v := item.Video
		if v == nil {
			continue
		}
		r := &SearchResultItem{
			ContentID:    v.ID,
			Title:        v.Title,
			ThumbnailURL: v.Thumbnail.Url,
			ViewCount:    v.Count.View,
			MylistCount:  v.Count.Mylist,
			CommentCount: v.Count.Comment,
			Description:  v.ShortDescription,
			StartTime:    v.RegisteredAt,
		}
		if v.Owner.Type == "user" {
			r.UserID, _ = strconv.Atoi(v.Owner.ID)
		} else if v.Owner.Type == "channel" {
			r.ChannelID, _ = strconv.Atoi(strings.TrimPrefix(v.Owner.ID, "ch"))
		}
		items = append(items, r)

	}
	return &SearchResult{Items: items, TotalCount: res.Data.TotalCount}, nil
}
