package nigonigo

import "strconv"

type RankingGenre = string
type RankingTerm = string

const (
	RankingAll             RankingGenre = "all"
	RankingHotTopic        RankingGenre = "hot-topic"
	RankingAnimal          RankingGenre = "animal"
	RankingCooking         RankingGenre = "cooking"
	RankingSports          RankingGenre = "sports"
	RankingTechnologyCraft RankingGenre = "technology_craft"
	RankingAnime           RankingGenre = "anime"
	RankingGame            RankingGenre = "game"
	RankingOther           RankingGenre = "other"

	Ranking24H  RankingTerm = "24h"
	RankingHour RankingTerm = "hour"
)

func (c *Client) GetVideoRanking(genre RankingGenre, term RankingTerm, page int) (*VideoListPage, error) {
	body, err := getContent(c.HttpClient, topUrl+"ranking/genre/"+genre+"?term="+term+"&page="+strconv.Itoa(page)+"&rss=2.0&lang=ja-jp", nil)
	if err != nil {
		return nil, err
	}
	return parseVideoRss(body)
}
