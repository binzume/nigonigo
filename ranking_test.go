package nigonigo

import (
	"testing"
)

func TestGetRanking(t *testing.T) {
	client := newClientForTest(t, false)

	ranking, err := client.GetVideoRanking(RankingAll, Ranking24H, 1)
	if err != nil {
		t.Fatalf("Failed to get ranking %v", err)
	}
	t.Log(ranking.Title)
	for _, item := range ranking.Items {
		t.Log(item.ContentID, item.Title)
	}
}
