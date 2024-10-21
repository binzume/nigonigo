package nigonigo

import (
	"testing"
)

var testTag = "MMD"

func TestSearchByTag(t *testing.T) {
	t.SkipNow() // 2024/10 search api is not available
	client := newClientForTest(t, false)

	result, err := client.SearchByTag(testTag, 0, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v", result.Items)
	}
	for _, item := range result.Items {
		t.Log(item)
	}
}

func TestSearchVideo(t *testing.T) {
	t.SkipNow() // 2024/10 search api is not available
	client := newClientForTest(t, false)

	filter := RangeFilter("viewCounter", "1000", "", false) // > 1000

	q := "アニメ OR ゲーム OR 料理" // TODO
	result, err := client.SearchVideo(q, []SearchField{SearchFieldCategoryTags}, DefaultFields, "-startTime", 0, 1, filter)

	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v  (%v)", result.Items, result)
	}
	for _, item := range result.Items {
		t.Log(item)
	}
}
