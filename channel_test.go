package nigonigo

import (
	"testing"
)

var testStrChannelID = "ch1"

func TestSearchChannel(t *testing.T) {
	client := newClientForTest(t, false)

	result, err := client.SearchChannel("MMD", SearchChannelModeKeyword, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	for _, item := range result {
		t.Log(item.ID, item.Name)
	}
}

func TestGetChannelVideos(t *testing.T) {
	client := newClientForTest(t, false)

	result, err := client.GetChannelVideos(testStrChannelID, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatalf("this account has no mylist: %v", result)
	}

	t.Logf("title: %v videos:%v", result.Title, len(result.Items))
	for _, item := range result.Items {
		t.Log(item)
	}
}
