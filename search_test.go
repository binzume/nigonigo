package nigonigo

import (
	"testing"
)

var testTag = "MMD"
var testChannelID = "2632720"

func TestSearchByTag(t *testing.T) {
	client := newClientForTest(t, false)

	result, err := client.SearchByTag(testTag, 0, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v", result.Items)
	}
}

func TestSearchByChannel(t *testing.T) {
	client := newClientForTest(t, false)

	result, err := client.SearchByChannel(testChannelID, 0, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v  (%v)", result.Items, result)
	}
}
