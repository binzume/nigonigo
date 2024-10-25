package nigonigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// errors
var (
	AuthenticationRequired = errors.New("authentication required")
	NotFound               = errors.New("not exist")
	InvalidResponse        = errors.New("invalid response received")
)

type MyList struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int64  `json:"user_id,string"`

	IsPublic    bool          `json:"isPublic"`
	CreatedAt   string        `json:"createdAt"`
	SampleItems []*MyListItem `json:"sampleItems"`
	Items       []*MyListItem `json:"items"`

	DefaultSortKey   string `json:"defaultSortKey"`
	DefaultSortOrder string `json:"defaultSortOrder"`
}

type MyListItemType int

const (
	MyListItemTypeVideo MyListItemType = 0
	MyListItemTypeSeiga MyListItemType = 5
	MyListItemTypeBook  MyListItemType = 6
)

func (m *MyListItemType) UnmarshalJSON(b []byte) error {
	var value json.Number
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	i, err := value.Int64()
	*m = MyListItemType(i)
	return err
}

type VideoInfo struct {
	BaseVideoInfo

	Owner OwnerInfo `json:"owner"`

	// mylist
	Type                 string   `json:"type"`
	ShortDescription     string   `json:"shortDescription"`
	IsChannelVideo       bool     `json:"isChannelVideo"`
	IsPaymentRequired    bool     `json:"isPaymentRequired"`
	LatestCommentSummary string   `json:"latestCommentSummary"`
	PlaybackPosition     *float32 `json:"playbackPosition"`
}

type MyListItem struct {
	ItemID      int       `json:"itemId"`
	WatchID     string    `json:"watchId"`
	Description string    `json:"description"`
	Data        VideoInfo `json:"video"`
	Status      string    `json:"status"`

	AddedAt string `json:"addedAt"`
}

func (c *Client) GetMyLists() ([]*MyList, error) {
	body, err := getContent(c.HttpClient, nvApiUrl+"users/me/mylists?sampleItemCount=0", nil)
	if err != nil {
		return nil, err
	}

	type myListResponse struct {
		Data struct {
			MyLists []*MyList `json:"mylists"`
		} `json:"data"`
		Meta any `json:"meta"`
	}

	var res myListResponse
	err = json.Unmarshal(body, &res)
	if err == nil {
		err = checkMylistResponse(body)
	}
	return res.Data.MyLists, err
}

func (c *Client) CreateMyList(mylist *MyList) error {
	params := map[string]string{
		"name":             mylist.Name,
		"description":      mylist.Description,
		"isPublic":         fmt.Sprint(mylist.IsPublic),
		"defaultSortKey":   "addedAt",
		"defaultSortOrder": "desc",
	}

	req, err := newPostReq(nvApiUrl+"users/me/mylists", params)
	if err != nil {
		return err
	}

	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}

	type myListResponse struct {
		Meta any `json:"meta"`
		Data struct {
			MyListID int    `json:"mylistId"`
			MyList   MyList `json:"mylist"`
		} `json:"Data"`
	}
	var result myListResponse
	err = json.Unmarshal(res, &result)
	if err != nil {
		return err
	}
	mylist.ID = result.Data.MyListID

	return checkMylistResponse(res)
}

func (c *Client) DeleteMyList(mylistId string) error {
	req, err := http.NewRequest("DELETE", nvApiUrl+"users/me/mylists/"+mylistId, nil)
	if err != nil {
		return err
	}
	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}

func (c *Client) GetMyListItems(mylistId string) ([]*MyListItem, error) {
	url := nvApiUrl + "users/me/mylists/" + mylistId + "/items"

	body, err := getContent(c.HttpClient, url, nil)
	if err != nil {
		return nil, err
	}

	type myListResponse struct {
		Data MyList `json:"Data"`
	}

	var res myListResponse
	err = json.Unmarshal(body, &res)
	if err == nil {
		err = checkMylistResponse(body)
	}
	return res.Data.Items, err
}

func (c *Client) AddMyListItem(mylistId, contentID, description string) error {
	params := map[string]string{
		"itemId":      contentID,
		"description": description,
	}

	url := nvApiUrl + "users/me/mylists/" + mylistId + "/items"
	req, err := newPostReq(url, params)
	if err != nil {
		return err
	}
	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}

func (c *Client) DeleteMyListItem(mylistId string, itemID string) error {
	url := nvApiUrl + "users/me/mylists/" + mylistId + "/items?itemIds=" + itemID
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}

func checkMylistResponse(body []byte) error {
	type myListResponse struct {
		Meta struct {
			Status    int    `json:"status"`
			ErrorCode string `json:"errorCode"`
		} `json:"meta"`
	}

	var res myListResponse
	err := json.Unmarshal(body, &res)
	if res.Meta.ErrorCode != "" {
		switch res.Meta.Status {
		case 401:
			return AuthenticationRequired
		case 403:
			return AuthenticationRequired
		case 404:
			return NotFound
		default:
			return errors.New(res.Meta.ErrorCode)
		}
	}
	return err
}

func (c *Client) GetPublicMyList(mylistId string) (*MyList, []*VideoInfo, error) {
	body, err := getContent(c.HttpClient, topUrl+"/mylist/"+mylistId+"?rss=2.0&numbers=1", nil)
	if err != nil {
		return nil, nil, err
	}

	intId, _ := strconv.Atoi(mylistId)
	videos, err := parseVideoRss(body)
	return &MyList{ID: intId, Name: videos.Title}, videos.Items, err
}
