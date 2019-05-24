package nigonigo

import (
	"os"
	"testing"
)

var myListNameForTest = "_test_mylist"

func TestGetMyLists(t *testing.T) {
	if _, err := os.Stat(sessionFile); err != nil {
		t.Log("session file not exists")
		t.SkipNow()
	}

	client := NewClient()
	err := client.LoadLoginSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to login %v", err)
	}

	result, err := client.GetMyLists()
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("this account has no mylist: %v", result)
	}

	t.Logf("list :%v", result[0])
	items, err := client.GetMyListItems("")
	for _, item := range items {
		t.Log(item)
	}
}

func TestMyList_CreateDelete(t *testing.T) {
	if _, err := os.Stat(sessionFile); err != nil {
		t.Log("session file not exists")
		t.SkipNow()
	}

	client := NewClient()
	err := client.LoadLoginSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to login %v", err)
	}

	var mylist = &MyList{Name: myListNameForTest, Description: "test"}

	err = client.CreateMyList(mylist)
	if err != nil {
		t.Fatalf("failed to CreateMyList : %v", err)
	}
	if mylist.ID == "" {
		t.Fatalf("mylist.ID should not empty")
	}

	err = client.AddMyListItem(mylist.ID, "sm9", "test test test")
	if err != nil {
		t.Fatalf("failed to AddMyListItem : %v", err)
	}

	err = client.DeleteMyListItem(mylist.ID, "sm9")
	if err != nil {
		t.Fatalf("failed to DeleteMyListItem : %v", err)
	}

	lists, err := client.GetMyLists()
	if err != nil {
		t.Fatalf("Failed to GetMyLists %v", err)
	}
	for _, m := range lists {
		if m.ID == mylist.ID {
			if m.Name != mylist.Name {
				t.Errorf("mylist name unmached: %v != %v", m.Name, mylist.Name)
			}
			if m.Description != mylist.Description {
				t.Errorf("mylist desc unmached: %v != %v", m.Name, mylist.Name)
			}
			break
		}
	}

	err = client.DeleteMyList(mylist.ID)
	if err != nil {
		t.Fatalf("failed to DeleteMyList : %v", err)
	}

	// clean up
	for _, m := range lists {
		if m.Name == myListNameForTest && m.ID != mylist.ID {
			err = client.DeleteMyList(m.ID)
			if err != nil {
				t.Errorf("failed to DeleteMyList : %v", err)
			}
		}
	}
}

func TestGetPublicMyList(t *testing.T) {
	client := NewClient()

	_, items, err := client.GetPublicMyList("2569551")
	if err != nil {
		t.Fatalf("failed to DeleteMyList : %v", err)
	}

	if len(items) == 0 {
		t.Fatalf("failed to get items")
	}

	for _, item := range items {
		t.Log(item)
	}
}

func TestMyList_AuthError(t *testing.T) {
	client := NewClient()

	_, err := client.GetMyLists()
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.CreateMyList(&MyList{})
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.DeleteMyList("0")
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.AddMyListItem("0", "sm9", "test test")
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.DeleteMyListItem("0", "sm9")
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}
}
