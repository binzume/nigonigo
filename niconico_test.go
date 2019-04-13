package nigonigo

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

var accountFile = "configs/binzume.json"
var testVid = "sm9"
var testChannelID = "2632720"

func TestNiconico(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatalf("Failed to create instance.")
	}
	// TODO
}

func TestLogin(t *testing.T) {
	client := NewClient()
	err := client.LoginWithJsonFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to login %v", err)
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

	session, err := client.GetDMCSessionById(testVid)
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
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	client := NewClient()
	session, err := client.GetDMCSessionById(testVid)
	if err != nil {
		t.Errorf("Failed to create session: %v", err)
	}

	out, _ := os.Create(testVid + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(ctx, session, out)
	if err != nil {
		t.Errorf("Failed to download: %v", err)
	}
}

func TestDownloadLoggedIn(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client := NewClient()
	err := client.LoginWithJsonFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to login %v", err)
	}

	result, err := client.SearchByChannel(testChannelID, 0, 2)
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result.Items) < 2 {
		t.Fatalf("Failed to get result. items: %v  (%v)", result.Items, result)
	}

	contantID := result.Items[1].ContentID
	session, err := client.GetDMCSessionById(contantID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contantID + "." + session.FileExtension())
	defer out.Close()
	time.Sleep(1 * time.Second)
	err = client.Download(ctx, session, out)
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}
}
