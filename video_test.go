package nigonigo

import (
	"context"
	"os"
	"testing"
	"time"
)

var testVid = "sm5188096"
var downloadTimeout = 25 * time.Second

func TestGetVideoData(t *testing.T) {
	client := newClientForTest(t, false)
	data, err := client.GetVideoData(testVid)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	nvComment := data.Comment.NvComment
	if nvComment != nil {
		threads, err := client.GetComments(nvComment.Server, nvComment.Threadkey, nvComment.Params)
		if err != nil {
			t.Fatalf("Failed to get comment: %v", err)
		}
		for _, thread := range threads {
			for _, comment := range thread.Comments {
				t.Logf("comment: %v", comment)
			}
		}
	}
}

func TestDownload(t *testing.T) {
	client := newClientForTest(t, false)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var contentID = testVid
	session, err := client.CreateVideoSession(contentID)
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
	client := newClientForTest(t, true)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	contantID := testVid
	session, err := client.CreateVideoSession(contantID)
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
