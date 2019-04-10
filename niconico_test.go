package nigonigo

import (
	"testing"
)

func TestNiconico(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Errorf("Failed to create instance.")
	}
	// TODO
}

func TestLogin(t *testing.T) {
	client := NewClient()
	err := client.LoginWithJsonFile("configs/binzume.json")
	if err != nil {
		t.Errorf("Failed to login %v", err)
	}
}

func TestSearch(t *testing.T) {
	client := NewClient()
	result, err := client.SearchByTag("a", 0, 1, nil)
	if err != nil {
		t.Errorf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Failed to get result. items: %v", result.Items)
	}
}

func TestPlay(t *testing.T) {
	client := NewClient()

	url, err := client.GetContentUrl("sm9")
	if err != nil {
		t.Errorf("Failed to request %v", err)
	}
	t.Log(url)
}
