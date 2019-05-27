package nigonigo

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
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

	CreatedDate int64 `json:"create_date"`
	UpdatedDate int64 `json:"update_date"`

	SortOrder     string `json:"sort_order"`
	PlaylistToken string `json:"watch_playlist"`
}

const ItemTypeVideo = 0
const ItemTypeSeiga = 5
const ItemTypeBook = 6

type MyListItemVideo struct {
	ContentID     string `json:"video_id"`
	Title         string `json:"title"`
	ThumbnailURL  string `json:"thumbnail_url"`
	Duration      int    `json:"length_seconds,string"`
	ViewCount     int    `json:"view_counter,string"`
	MylistCount   int    `json:"mylist_counter,string"`
	CommentCount  int    `json:"num_res,string"`
	FirstRetrieve int64  `json:"first_retrieve"`

	Deleted int `json:"deleted,string"`
}

type MyListItem struct {
	ItemID      string          `json:"item_id"`
	Type        int             `json:"item_type,string"`
	Description string          `json:"description"`
	Data        MyListItemVideo `json:"item_data"`

	CreatedTime int64 `json:"create_time"`
	UpdatedTime int64 `json:"update_time"`
	Watch       int   `json:"watch"`
}

func (c *Client) GetMyLists() ([]*MyList, error) {
	body, err := GetContent(c.HttpClient, topUrl+"api/mylistgroup/list", nil)
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
	token, err := c.GetCsrfToken()
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

	req, err := NewPostReq(topUrl+"api/mylistgroup/add", params)
	if err != nil {
		return err
	}

	res, err := DoRequest(c.HttpClient, req)
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
	token, err := c.GetCsrfToken()
	if err != nil {
		return err
	}
	params := map[string]string{
		"group_id": mylistId,
		"token":    token,
	}

	req, err := NewPostReq(topUrl+"api/mylistgroup/delete", params)
	if err != nil {
		return err
	}
	res, err := DoRequest(c.HttpClient, req)
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
	if mylistId != "" {
		url = topUrl + "api/mylist/list?group_id=" + mylistId
	} else {
		url = topUrl + "api/deflist/list"
	}

	body, err := GetContent(c.HttpClient, url, nil)
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
	token, err := c.GetCsrfToken()
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
	req, err := NewPostReq(url, params)
	if err != nil {
		return err
	}
	res, err := DoRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}

func (c *Client) DeleteMyListItem(mylistId string, itemID string) error {
	token, err := c.GetCsrfToken()
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
	req, err := NewPostReq(url, params)
	if err != nil {
		return err
	}
	res, err := DoRequest(c.HttpClient, req)
	if err != nil {
		return err
	}
	return checkMylistResponse(res)
}
func (c *Client) GetCsrfToken() (string, error) {
	body, err := GetContent(c.HttpClient, topUrl+"my/mylist", nil)
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

func (c *Client) GetPublicMyList(mylistId string) (*MyList, []*MyListItem, error) {
	body, err := GetContent(c.HttpClient, topUrl+"/mylist/"+mylistId+"?rss=2.0", nil)
	if err != nil {
		return nil, nil, err
	}

	var myListRss struct {
		Title string `xml:"channel>title"`
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Description string `xml:"description"`
		} `xml:"channel>item"`
	}
	err = xml.Unmarshal(body, &myListRss)
	if err != nil {
		return nil, nil, err
	}

	var desc struct {
		Imgs []struct {
			Src string `xml:"src,attr"`
		} `xml:"p>img"`
		Info []struct {
			Class string `xml:"class,attr"`
			Value string `xml:",chardata"`
		} `xml:"p>small>strong"`
	}
	var items []*MyListItem
	for i, ritem := range myListRss.Items {
		id := path.Base(ritem.Link)
		t, _ := time.Parse(time.RFC1123Z, ritem.PubDate)

		item := &MyListItem{
			ItemID:      id,
			Data:        MyListItemVideo{Title: ritem.Title, ContentID: id, FirstRetrieve: t.Unix()},
			CreatedTime: t.Unix(),
			UpdatedTime: t.Unix(),
		}
		err = xml.Unmarshal([]byte("<x>"+myListRss.Items[i].Description+"</x>"), &desc)
		if err == nil && len(desc.Imgs) > 0 {
			item.Data.ThumbnailURL = desc.Imgs[0].Src
		}
		for _, d := range desc.Info {
			if d.Class == "nico-info-length" {
				t := strings.SplitN(d.Value, ":", 2)
				if len(t) == 2 {
					m, _ := strconv.Atoi(t[0])
					s, _ := strconv.Atoi(t[1])
					item.Data.Duration = m*60 + s
				}
			}
			if d.Class == "nico-info-date" {
				t, err := time.Parse("2006年01月02日 15：04：05", d.Value)
				if err == nil {
					item.Data.FirstRetrieve = t.Unix()
				}
			}
		}
		items = append(items, item)
	}

	return &MyList{ID: mylistId, Name: myListRss.Title}, items, nil
}
