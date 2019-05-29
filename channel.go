package nigonigo

import (
	"fmt"
	"regexp"
)

var channelUrl = "https://ch.nicovideo.jp/"

type SearchChannelMode string

var (
	SearchChannelModeKeyword SearchChannelMode = "s"
	SearchChannelModeTag     SearchChannelMode = "t"
)

type ChannelInfo struct {
	ID          string
	Name        string
	Description string
}

func (c *Client) SearchChannel(q string, mode SearchChannelMode, page int) ([]*ChannelInfo, error) {
	body, err := GetContent(c.HttpClient, channelUrl+"search/"+q+"?mode="+string(mode)+"&page="+fmt.Sprint(page), nil)
	if err != nil {
		return nil, err
	}

	var channels []*ChannelInfo
	re := regexp.MustCompile(`(?s)<span class="thumb_wrapper thumb_wrapper_ch[^"]*">(.+?</div>.+?)</span>`)
	linkRe := regexp.MustCompile(`(?s)<a href="/(\w+)"[^>]*title="([^"]*)"`)
	descRe := regexp.MustCompile(`(?s)<span class="channel_detail">(.*?)</span>`)
	for _, match := range re.FindAllSubmatch(body, -1) {
		m := linkRe.FindStringSubmatch(string(match[1]))
		if m == nil {
			continue
		}
		ch := &ChannelInfo{ID: m[1], Name: m[2]}
		m = descRe.FindStringSubmatch(string(match[1]))
		if m != nil {
			ch.Description = m[1]
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (c *Client) GetChannelVideos(channelID string, page int) (*VideoListPage, error) {
	body, err := GetContent(c.HttpClient, channelUrl+channelID+"/video?rss=2.0&numbers=1&page="+fmt.Sprint(page), nil)
	if err != nil {
		return nil, err
	}
	return parseVideoRss(body)
}
