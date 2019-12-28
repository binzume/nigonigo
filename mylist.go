package nigonigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// errors
var (
	AuthenticationRequired = errors.New("authentication required")
	InvalidResponse        = errors.New("invalid response received")
)

type MyList struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int64  `json:"user_id,string"`
	Public      int    `json:"public,string"`

	CreatedTime int64 `json:"create_time"`
	UpdatedTime int64 `json:"update_time"`

	SortOrder     string `json:"sort_order"`
	PlaylistToken string `json:"watch_playlist"`
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

type MyListItem struct {
	ItemID      string         `json:"item_id"`
	Type        MyListItemType `json:"item_type"`
	Description string         `json:"description"`
	Data        VideoInfo      `json:"item_data"`

	CreatedTime int64 `json:"create_time"`
	UpdatedTime int64 `json:"update_time"`
	Watch       int   `json:"watch"`
}

func (c *Client) GetMyLists() ([]*MyList, error) {
	body, err := getContent(c.HttpClient, topUrl+"api/mylistgroup/list", nil)
	if err != nil {
		return nil, err
	}

	type myListResponse struct {
		MyListGroup []*MyList `json:"mylistgroup"`
		Status      string    `json:"status"`
	}

	var res myListResponse
	err = json.Unmarshal(body, &res)
	if err == nil {
		err = checkMylistResponse(body)
	}
	return res.MyListGroup, err
}

func (c *Client) CreateMyList(mylist *MyList) error {
	token, err := c.getCsrfToken()
	if err != nil {
		return err
	}

	params := map[string]string{
		"name":         mylist.Name,
		"description":  mylist.Description,
		"public":       fmt.Sprint(mylist.Public),
		"default_sort": "1",
		"icon_id":      "1",
		"token":        token,
	}

	req, err := newPostReq(topUrl+"api/mylistgroup/add", params)
	if err != nil {
		return err
	}

	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}

	type myListResponse struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	var result myListResponse
	err = json.Unmarshal(res, &result)
	if err != nil {
		return err
	}
	mylist.ID = fmt.Sprint(result.ID)

	return checkMylistResponse(res)
}

func (c *Client) DeleteMyList(mylistId string) error {
	token, err := c.getCsrfToken()
	if err != nil {
		return err
	}
	params := map[string]string{
		"group_id": mylistId,
		"token":    token,
	}

	req, err := newPostReq(topUrl+"api/mylistgroup/delete", params)
	if err != nil {
		return err
	}
	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}

func (c *Client) GetDefListItems() ([]*MyListItem, error) {
	return c.GetMyListItems("")
}

func (c *Client) GetMyListItems(mylistId string) ([]*MyListItem, error) {
	var url string
	if mylistId == "" || mylistId == "default" {
		url = topUrl + "api/deflist/list"
	} else {
		url = topUrl + "api/mylist/list?group_id=" + mylistId
	}

	body, err := getContent(c.HttpClient, url, nil)
	if err != nil {
		return nil, err
	}

	type myListResponse struct {
		MylistItems []*MyListItem `json:"mylistitem"`
		Status      string        `json:"status"`
	}

	var res myListResponse
	err = json.Unmarshal(body, &res)
	if err == nil {
		err = checkMylistResponse(body)
	}
	return res.MylistItems, err
}

func (c *Client) AddMyListItem(mylistId, contentID, description string) error {
	token, err := c.getCsrfToken()
	if err != nil {
		return err
	}
	params := map[string]string{
		"group_id":    mylistId,
		"item_id":     contentID,
		"description": description,
		"token":       token,
	}

	var url string
	if mylistId != "" {
		url = topUrl + "api/mylist/add"
	} else {
		url = topUrl + "api/deflist/add"
	}
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
	token, err := c.getCsrfToken()
	if err != nil {
		return err
	}
	params := map[string]string{
		"group_id":     mylistId,
		"id_list[0][]": itemID,
		"token":        token,
	}

	var url string
	if mylistId != "" {
		url = topUrl + "api/mylist/delete"
	} else {
		url = topUrl + "api/deflist/delete"
	}
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
func (c *Client) getCsrfToken() (string, error) {
	body, err := getContent(c.HttpClient, topUrl+"my/mylist", nil)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`NicoAPI.token\s*=\s*"([0-9a-f-]+)"`)
	match := re.FindStringSubmatch(string(body))
	if match == nil {
		return "", InvalidResponse
	}

	return match[1], nil
}

func checkMylistResponse(body []byte) error {
	type myListResponse struct {
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
		Status string `json:"status"`
	}

	var res myListResponse
	err := json.Unmarshal(body, &res)
	if res.Error != nil {
		switch res.Error.Code {
		case "NOAUTH":
			return AuthenticationRequired
		case "NONEXIST":
			return errors.New(res.Error.Code)
		default:
			return errors.New(res.Error.Code)
		}
	}
	return err
}

func (c *Client) GetPublicMyList(mylistId string) (*MyList, []*VideoInfo, error) {
	body, err := getContent(c.HttpClient, topUrl+"/mylist/"+mylistId+"?rss=2.0&numbers=1", nil)
	if err != nil {
		return nil, nil, err
	}

	videos, err := parseVideoRss(body)
	return &MyList{ID: mylistId, Name: videos.Title}, videos.Items, nil
}
