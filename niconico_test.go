package nigonigo

import (
	"log"
	"os"
	"testing"
)

var accountFile = "configs/account.json"
var sessionFile = "configs/session.json"
var testChannelID = "2632720"

func newClientForTest(t *testing.T, login bool) *Client {
	client := NewClient()
	if login {
		if _, err := os.Stat(sessionFile); err != nil {
			t.Log("account file not exists")
			t.SkipNow()
		}
		err := client.LoadLoginSession(sessionFile)
		if err != nil {
			t.Fatalf("Failed to login %v", err)
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

func TestLogin(t *testing.T) {
	if _, err := os.Stat(accountFile); err != nil {
		t.Log("account file not exists")
		t.SkipNow()
	}

	client := NewClient()
	err := client.LoginWithJsonFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	client.SaveLoginSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	client2 := NewClient()
	// SessionString format : user_session_XXXX_XXXXXXXXXXXXXXX...
	err = client2.LoadLoginSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if client2.Session.NiconicoID != client.Session.NiconicoID {
		t.Fatalf("Failed to set session string")
	}
}

func TestSession(t *testing.T) {
	client := NewClient()

	session, err := client.CreateDMCSessionById(testVid)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}

	err = client.Heartbeat(session)
	if err != nil {
		t.Fatalf("Failed to Heartbeat %v", err)
	}
	log.Println(session.ContentURI)
}
