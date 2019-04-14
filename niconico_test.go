package nigonigo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

var accountFile = "configs/account.json"
var sessionFile = "configs/session.json"
var testVid = "sm9"
var testChannelID = "2632720"
var downloadTimeout = 5 * time.Second

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

func TestSearchByTag(t *testing.T) {
	client := NewClient()
	result, err := client.SearchByTag("MMD", 0, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v", result.Items)
	}
}

func TestSearchByChannel(t *testing.T) {
	client := NewClient()
	result, err := client.SearchByChannel(testChannelID, 0, 1)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("Failed to get result. items: %v  (%v)", result.Items, result)
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

func TestDownload(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := NewClient()
	var contentID = testVid
	session, err := client.CreateDMCSessionById(contentID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contentID + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(ctx, session, out)
	if err == context.DeadlineExceeded {
		t.Logf("Download stoppped")
	} else if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}
}

func TestDownloadLoggedIn(t *testing.T) {
	if _, err := os.Stat(sessionFile); err != nil {
		t.Log("account file not exists")
		t.SkipNow()
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	client := NewClient()
	err := client.LoadLoginSession(sessionFile)
	if err != nil {
		// err = client.LoginWithJsonFile(accountFile)
		t.Fatalf("Failed to login %v", err)
	}

	result, err := client.SearchByChannel(testChannelID, 0, 10)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatalf("Failed to get result: %v", result)
	}

	contantID := result.Items[0].ContentID
	session, err := client.CreateDMCSessionById(contantID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contantID + "." + session.FileExtension())
	defer out.Close()
	time.Sleep(1 * time.Second)
	err = client.Download(ctx, session, out)
	if err == context.DeadlineExceeded {
		t.Logf("Download stoppped")
	} else if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}
}

func TestDownloadFromSmile(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	client := NewClient()
	var contentID = testVid
	video, err := client.GetVideoData(contentID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contentID + "." + video.SmileFileExtension())
	defer out.Close()
	err = client.DownloadFromSmile(ctx, video, out)
	if err == context.DeadlineExceeded {
		t.Logf("Download stoppped")
	} else if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}
}
