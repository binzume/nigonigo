package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/binzume/nigonigo"
)

func cmdSearch() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s search [options] [query]\n", os.Args[0])
		flag.PrintDefaults()
	}
	tag := flag.String("t", "", "tag")
	offset := flag.Int("offset", 0, "offset")
	limit := flag.Int("limit", 100, "limit")
	jsonFormat := flag.Bool("json", false, "json output")
	sortOrder := flag.String("sort", "-startTime", "sort order")
	filterGenre := flag.String("genre", "", "filter by genre")
	filterUserID := flag.String("user", "", "filter by userId (obsoleted)")
	filterChannelID := flag.String("ch", "", "filter by channelId (obsoleted)")
	filterViewCount := flag.String("viewCount", "0", "filter by viewCounter")
	seriesID := flag.String("series", "", "series id")
	videoID := flag.String("id", "", "video id")
	chOnly := flag.Bool("chOnly", false, "channel only")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	client := nigonigo.NewClient()

	req := &nigonigo.SearchRequest{
		Query:   flag.Arg(0),
		Sort:    *sortOrder,
		Offset:  *offset,
		Limit:   *limit,
		Targets: []nigonigo.SearchField{"description,title"},
	}

	var filters []nigonigo.SearchFilter
	if *filterGenre != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldGenre, *filterGenre))
	}
	if *filterUserID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldUserID, *filterUserID))
	}
	if *filterChannelID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldChannelID, *filterChannelID))
	}
	if *filterViewCount != "0" {
		filters = append(filters, nigonigo.RangeFilter(nigonigo.SearchFieldViewCounter, *filterViewCount, "", true))
	}

	if *tag != "" {
		if req.Query == "" {
			req.Targets = []nigonigo.SearchField{"tagsExact,categoryTags"}
			req.Query = *tag
		} else {
			filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldTagsExact, *tag))
		}
	}

	if len(filters) == 1 {
		req.Filter = filters[0]
	} else if len(filters) > 1 {
		req.Filter = nigonigo.AndFilter(filters)
	}

	var err error
	var result *nigonigo.SearchResult
	if *seriesID != "" {
		result, err = client.FindSeriesVideos(*seriesID)
	} else if *videoID != "" {
		result = fromVideoID(client, *videoID)
	} else {
		result, err = client.SearchVideo2(req)
	}
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}

	if *chOnly {
		r := []*nigonigo.SearchResultItem{}
		for _, v := range result.Items {
			if v.ChannelID != 0 {
				r = append(r, v)
			}
		}
		result.Items = r
	}

	if *jsonFormat {
		jsonStr, _ := json.MarshalIndent(result.Items, "", "   ")
		os.Stdout.Write(jsonStr)
	} else {
		for _, item := range result.Items {
			fmt.Printf("%v\t%v\t%v\n", item.ContentID, item.StartTime, item.Title)
		}
	}
}

func fromVideoID(client *nigonigo.Client, id string) *nigonigo.SearchResult {
	data, err := client.GetVideoData(id)
	if err != nil {
		return &nigonigo.SearchResult{TotalCount: 0}
	}
	var tags []string
	for _, t := range data.Tag.Items {
		tags = append(tags, t.Name)
	}
	userId, _ := data.Owner["id"].(float64)
	channelId, _ := data.Channel["id"].(string)
	channelId2, _ := strconv.Atoi(strings.TrimPrefix(channelId, "ch"))
	return &nigonigo.SearchResult{TotalCount: 1, Items: []*nigonigo.SearchResultItem{{
		ContentID:    data.Video.ContentID,
		Title:        data.Video.Title,
		Description:  data.Video.Description,
		ViewCount:    data.Video.Count.View,
		MylistCount:  data.Video.Count.Mylist,
		CommentCount: data.Video.Count.Comment,
		ThumbnailURL: data.Video.Thumbnail.Url,
		Duration:     data.Video.Duration,
		StartTime:    data.Video.RegisteredAt,
		UserID:       int(userId),
		ChannelID:    int(channelId2),
		Tags:         strings.Join(tags, ","),
	}}}
}
