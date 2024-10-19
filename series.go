package nigonigo

import (
	"regexp"
	"strconv"
	"strings"
)

func (c *Client) FindSeriesVideos(sid string) (*SearchResult, error) {

	body, err := getContent(c.HttpClient, seriesUrl+sid, nil)
	if err != nil {
		return nil, err
	}
	items := []*SearchResultItem{}

	re := regexp.MustCompile(`(?s)<div class="NC-VideoMediaObjectWrapper"(.+?NC-VideoMetaCount_mylist"\s*>[0-9,]*</div>)`)
	linkRe := regexp.MustCompile(`NC-MediaObject-contents"\s+href="https://www.nicovideo.jp/watch/([a-z0-9]+)`)
	titleRe := regexp.MustCompile(`data-background-image="([^"]+)"[^>]+aria-label="([^"]+)"`)
	countRe := regexp.MustCompile(`(?s)NC-VideoMetaCount_view"\s*>([0-9,]*).*?NC-VideoMetaCount_comment"\s*>([0-9,]*).*?NC-VideoMetaCount_mylist"\s*>([0-9,]*)`)
	channelRe := regexp.MustCompile(`data-owner-id="ch([0-9]+)"`)
	dateRe := regexp.MustCompile(`(?s)"NC-VideoRegisteredAtText-text">\s*([0-9/]+ [0-9:]+)\s*<`)

	for _, match := range re.FindAllSubmatch(body, -1) {
		m := linkRe.FindStringSubmatch(string(match[1]))
		if m == nil {
			continue
		}
		item := &SearchResultItem{ContentID: m[1]}

		m = titleRe.FindStringSubmatch(string(match[1]))
		if m != nil {
			item.ThumbnailURL = m[1]
			item.Title = m[2]
		}

		m = countRe.FindStringSubmatch(string(match[1]))
		if m != nil {
			item.ViewCount, _ = strconv.Atoi(strings.ReplaceAll(m[1], ",", ""))
			item.CommentCount, _ = strconv.Atoi(strings.ReplaceAll(m[2], ",", ""))
			item.MylistCount, _ = strconv.Atoi(strings.ReplaceAll(m[3], ",", ""))
		}

		m = dateRe.FindStringSubmatch(string(match[1]))
		if m != nil {
			// time.Parse("2006/1/2 15:4", m[1])
			item.StartTime = m[1]
		}

		m = channelRe.FindStringSubmatch(string(match[1]))
		if m != nil {
			item.ChannelID, _ = strconv.Atoi(m[1])
		}

		items = append(items, item)
	}

	return &SearchResult{Items: items, TotalCount: len(items)}, nil
}
