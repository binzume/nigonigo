package nigonigo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// doc https://site.nicovideo.jp/search-api-docs/search.html

type SearchRequest struct {
	Query   string
	Offset  int
	Limit   int
	Sort    string
	Targets []SearchField
	Filter  SearchFilter
	Fields  []SearchField
}

type SearchResultItem struct {
	ContentID    string `json:"contentId"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     int    `json:"lengthSeconds"`
	ViewCount    int    `json:"viewCounter"`
	MylistCount  int    `json:"mylistCounter"`
	CommentCount int    `json:"commentCounter"`

	Description string `json:"description"`
	UserID      int    `json:"userId"`
	ChannelID   int    `json:"channelId"`
	ThreadID    int    `json:"threadId"`
	Tags        string `json:"tags"`
	StartTime   string `json:"startTime"`
}

type SearchField = string

const (
	SearchFieldContentID      SearchField = "contentId"
	SearchFieldTitle          SearchField = "title"
	SearchFieldLengthSeconds  SearchField = "lengthSeconds"
	SearchFieldThumbnailURL   SearchField = "thumbnailUrl"
	SearchFieldViewCounter    SearchField = "viewCounter"
	SearchFieldMylistCounter  SearchField = "mylistCounter"
	SearchFieldCommentCounter SearchField = "commentCounter"
	SearchFieldDescription    SearchField = "description"
	SearchFieldTags           SearchField = "tags"
	SearchFieldTagsExact      SearchField = "tagsExact"
	SearchFieldLockTagsExact  SearchField = "lockTagsExact"
	SearchFieldCategoryTags   SearchField = "categoryTags"
	SearchFieldGenre          SearchField = "genre"
	SearchFieldGenreKey       SearchField = "genre.keyword"
	SearchFieldStartTime      SearchField = "startTime"
	SearchFieldUserID         SearchField = "userId"
	SearchFieldChannelID      SearchField = "channelId"
	SearchFieldThreadID       SearchField = "threadId"
)

var DefaultFields = []SearchField{
	SearchFieldContentID,
	SearchFieldTitle,
	SearchFieldThumbnailURL,
	SearchFieldTags,
	SearchFieldStartTime,
	SearchFieldUserID,
	SearchFieldChannelID,
	SearchFieldLengthSeconds,
}

type SearchResult struct {
	TotalCount int
	Offset     int
	Items      []*SearchResultItem
}

// TODO
type SearchFilter interface{}

type simpleFilter struct {
	Type  string      `json:"type"`
	Field SearchField `json:"field"`
	Value string      `json:"value"`
}
type rangeFilter struct {
	Type         string      `json:"type"`
	Field        SearchField `json:"field"`
	From         string      `json:"from,omitempty"`
	To           string      `json:"to,omitempty"`
	IncludeUpper bool        `json:"include_upper"`
}

type groupFilter struct {
	Type    string         `json:"type"`
	Filters []SearchFilter `json:"filters"`
}

func AndFilter(filters []SearchFilter) SearchFilter {
	return &groupFilter{"and", filters}
}

func OrFilter(filters []SearchFilter) SearchFilter {
	return &groupFilter{"or", filters}
}

func NotFilter(filter SearchFilter) SearchFilter {
	return map[string]interface{}{"type": "not", "filter": filter}
}

func EqualFilter(field SearchField, value string) SearchFilter {
	return &simpleFilter{"equal", field, value}
}

func RangeFilter(field SearchField, from, to string, includeUpper bool) SearchFilter {
	return &rangeFilter{"range", field, from, to, includeUpper}
}

func (c *Client) SearchByTag(tag string, offset, limit int) (*SearchResult, error) {
	return c.SearchVideo(tag, []SearchField{"tagsExact,categoryTags"}, nil, "-startTime", offset, limit, nil)
}

func (c *Client) SearchByKeyword(s string, offset, limit int) (*SearchResult, error) {
	return c.SearchVideo(s, nil, nil, "-startTime", offset, limit, nil)
}

func (c *Client) SearchVideo(q string, targets, fields []SearchField, sort string, offset, limit int, filter SearchFilter) (*SearchResult, error) {
	return c.SearchVideo2(&SearchRequest{
		Query:   q,
		Offset:  offset,
		Limit:   limit,
		Sort:    sort,
		Filter:  filter,
		Targets: targets,
		Fields:  fields,
	})
}

func (c *Client) SearchVideo2(req *SearchRequest) (*SearchResult, error) {
	if req.Fields == nil {
		req.Fields = DefaultFields
	}
	if req.Targets == nil {
		req.Fields = []SearchField{"description,title"}
	}
	params := map[string]string{
		"q":        req.Query,
		"targets":  strings.Join(req.Targets, ","),
		"fields":   strings.Join(req.Fields, ","),
		"_sort":    req.Sort,
		"_offset":  fmt.Sprint(req.Offset),
		"_limit":   fmt.Sprint(req.Limit),
		"_context": "nigonigo",
	}
	if req.Filter != nil {
		encoded, err := json.Marshal(req.Filter)
		if err != nil {
			return nil, err
		}
		params["jsonFilter"] = string(encoded)
	}

	body, err := getContent(c.HttpClient, searchApiUrl, params)
	if err != nil {
		return nil, err
	}
	// Logger.Println(string(body))

	type searchResponse struct {
		Meta struct {
			Status       int    `json:"status"`
			TotalCount   int    `json:"totalCount"`
			ID           string `json:"id"`
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"meta"`
		Data []*SearchResultItem `json:"data"`
	}
	var res searchResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.Meta.Status != 200 {
		return nil, fmt.Errorf("%v code:%v", res.Meta.ErrorMessage, res.Meta.ErrorCode)
	}

	return &SearchResult{TotalCount: res.Meta.TotalCount, Offset: req.Offset, Items: res.Data}, nil
}
