package nigonigo

import (
	"encoding/xml"
	"path"
	"strconv"
	"strings"
	"time"
)

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
		id := strings.SplitN(path.Base(ritem.Link), "?", 2)[0]
		t, _ := time.Parse(time.RFC1123Z, ritem.PubDate)

		item := &VideoInfo{BaseVideoInfo: BaseVideoInfo{Title: ritem.Title, ContentID: id, RegisteredAt: t.Format(time.RFC3339)}}

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
			item.Thumbnail.Url = desc.Imgs[0].Src
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
					item.RegisteredAt = t.Format(time.RFC3339)
				}
			} else if d.Class == "nico-numbers-view" || d.Class == "nico-info-total-view" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.Count.View = n
			} else if d.Class == "nico-numbers-mylist" || d.Class == "nico-info-total-mylist" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.Count.Mylist = n
			} else if d.Class == "nico-numbers-res" || d.Class == "nico-info-total-res" {
				n, _ := strconv.Atoi(strings.Replace(d.Value, ",", "", -1))
				item.Count.Comment = n
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
