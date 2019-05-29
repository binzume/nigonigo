package nigonigo

import (
	"encoding/xml"
	"path"
	"strconv"
	"strings"
	"time"
)

type VideoInfo struct {
	ContentID    string `json:"video_id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Duration     int    `json:"length_seconds,string"`
	ViewCount    int    `json:"view_counter,string"`
	MylistCount  int    `json:"mylist_counter,string"`
	CommentCount int    `json:"num_res,string"`
	StartTime    int64  `json:"first_retrieve"`
}

type VideoListPage struct {
	Title string
	Owner string
	Items []*VideoInfo
}

// parse RSS feed
// see http://nicowiki.com/RSS%E3%83%95%E3%82%A3%E3%83%BC%E3%83%89%E4%B8%80%E8%A6%A7.html
func parseVideoRss(body []byte) (*VideoListPage, error) {

	var videoListRss struct {
		Title   string `xml:"channel>title"`
		Creater string `xml:"channel>creator"`
		Items   []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Description string `xml:"description"`
		} `xml:"channel>item"`
	}
	err := xml.Unmarshal(body, &videoListRss)
	if err != nil {
		return nil, err
	}

	var items []*VideoInfo
	for _, ritem := range videoListRss.Items {
		id := path.Base(ritem.Link)
		t, _ := time.Parse(time.RFC1123Z, ritem.PubDate)

		item := &VideoInfo{Title: ritem.Title, ContentID: id, StartTime: t.Unix()}

		var desc struct {
			Imgs []struct {
				Src string `xml:"src,attr"`
			} `xml:"p>img"`
			Info []struct {
				Class string `xml:"class,attr"`
				Value string `xml:",chardata"`
			} `xml:"p>small>strong"`
		}
		descText := strings.Replace(ritem.Description, "&nbsp;", " ", -1)
		err = xml.Unmarshal([]byte("<xx>"+descText+"</xx>"), &desc)
		if len(desc.Imgs) > 0 {
			item.ThumbnailURL = desc.Imgs[0].Src
		}
		for _, d := range desc.Info {
			if d.Class == "nico-info-length" {
				t := strings.SplitN(d.Value, ":", 2)
				if len(t) == 2 {
					m, _ := strconv.Atoi(t[0])
					s, _ := strconv.Atoi(t[1])
					item.Duration = m*60 + s
				}
			} else if d.Class == "nico-info-date" {
				t, err := time.Parse("2006年01月02日 15：04：05", d.Value)
				if err == nil {
					item.StartTime = t.Unix()
				}
			} else if d.Class == "nico-numbers-view" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.ViewCount = n
			} else if d.Class == "nico-numbers-mylist" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.MylistCount = n
			} else if d.Class == "nico-numbers-res" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.CommentCount = n
			}
		}
		items = append(items, item)
	}

	return &VideoListPage{
		Title: videoListRss.Title,
		Owner: videoListRss.Creater,
		Items: items,
	}, nil
}
