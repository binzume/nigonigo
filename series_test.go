package nigonigo

import (
	"testing"
)

var testSeries = "96269"

func TestSeries(t *testing.T) {
	client := newClientForTest(t, false)

	result, err := client.GetSeriesVideos(testSeries)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatalf("Failed to get result. items: %v", result.Items)
	}
	for _, item := range result.Items {
		t.Log(item)
	}
}
