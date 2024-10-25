package nigonigo

import (
	"os"
	"testing"
)

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
	err = client2.LoadLoginSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	if client2.Session.NiconicoID != client.Session.NiconicoID {
		t.Fatalf("Failed to set session string")
	}
}

func TestLogout(t *testing.T) {
	client := newClientForTest(t, true)

	err := client.Logout()
	if err != nil {
		t.Fatalf("Failed to logout: %v", err)
	}

	err = client.updateLoginSession()
	if err == nil {
		t.Fatalf("err should be not nil")
	}
}

func TestGetSessions(t *testing.T) {
	client := newClientForTest(t, true)

	sessions, err := client.GetAvailableSessions()
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	for _, s := range sessions {
		t.Logf("SESSION :%v", s)
	}
}

func TestGetAccountStatus(t *testing.T) {
	client := newClientForTest(t, true)

	status, err := client.GetAccountStatus()
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	t.Logf("SESSION :%#v", status)
}
