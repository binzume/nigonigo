package nigonigo

import (
	"os"
	"testing"
)

var accountFile = "configs/account.json"
var sessionFile = "configs/session.json"

func newClientForTest(t *testing.T, login bool) *Client {
	client := NewClient()
	if login {
		if _, err := os.Stat(sessionFile); err != nil {
			t.Log("account file not exists")
			t.SkipNow()
		}
		err := client.LoadLoginSession(sessionFile)
		if err != nil {
			t.Log("do login")
			err = client.LoginWithJsonFile(accountFile)
			if err == nil {
				_ = client.SaveLoginSession(sessionFile)
			}
		}
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}
	}
	return client
}

func TestNiconico(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatalf("Failed to create instance.")
	}
	// TODO
}
