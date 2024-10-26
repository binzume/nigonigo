package nigonigo

import (
	"strconv"
	"testing"
)

var myListNameForTest = "_test_mylist"
var publicMyListId = "2569551"
var testUserID = 1842795

func TestGetUserMyList(t *testing.T) {
	client := newClientForTest(t, false)

	mylists, err := client.GetUserMyLists(testUserID)
	if err != nil {
		t.Fatalf("Failed to get mylists: %v", err)
	}
	if len(mylists) == 0 {
		t.Fatalf("no mylist")
	}
	for _, item := range mylists {
		t.Logf("%v", item)
	}
}

func TestGetMyList(t *testing.T) {
	client := newClientForTest(t, false)

	mylist, err := client.GetMyList(publicMyListId)
	if err != nil {
		t.Fatalf("Failed to request: %v", err)
	}
	if len(mylist.Items) == 0 {
		t.Fatalf("no items")
	}

	for _, item := range mylist.Items {
		t.Logf("%v", item)
	}
}

func TestGetMyLists(t *testing.T) {
	client := newClientForTest(t, true)

	result, err := client.GetMyLists()
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	if len(result) == 0 {
		t.Fatalf("this account has no mylist: %v", result)
	}
	for _, item := range result {
		t.Logf("%#v", item)
	}

	items, err := client.GetMyListItems(strconv.Itoa(result[2].ID))
	if err != nil {
		t.Fatalf("Failed to request %v", err)
	}
	for _, item := range items {
		t.Logf("%v", item)
	}
}

func TestMyList_CreateDelete(t *testing.T) {
	client := newClientForTest(t, true)

	var mylist = &MyList{Name: myListNameForTest, Description: "test"}

	err := client.CreateMyList(mylist)
	if err != nil {
		t.Fatalf("failed to CreateMyList : %v", err)
	}
	if mylist.ID == 0 {
		t.Fatalf("mylist.ID should not empty")
	}

	err = client.AddMyListItem(strconv.Itoa(mylist.ID), "sm9", "test test test")
	if err != nil {
		t.Fatalf("failed to AddMyListItem : %v", err)
	}

	err = client.DeleteMyListItem(strconv.Itoa(mylist.ID), "sm9")
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

	err = client.DeleteMyList(strconv.Itoa(mylist.ID))
	if err != nil {
		t.Fatalf("failed to DeleteMyList : %v", err)
	}

	// clean up
	for _, m := range lists {
		if m.Name == myListNameForTest && m.ID != mylist.ID {
			err = client.DeleteMyList(strconv.Itoa(m.ID))
			if err != nil {
				t.Errorf("failed to DeleteMyList : %v", err)
			}
		}
	}
}

func TestGetPublicMyList(t *testing.T) {
	client := newClientForTest(t, false)

	_, items, err := client.GetPublicMyListRSS(publicMyListId)
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
	client := newClientForTest(t, false)

	_, err := client.GetMyLists()
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.CreateMyList(&MyList{})
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.DeleteMyList(publicMyListId)
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.AddMyListItem(publicMyListId, "sm9", "test test")
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.DeleteMyListItem(publicMyListId, "sm9")
	if err != AuthenticationRequired {
		t.Fatalf("expected: %v  got: %v", AuthenticationRequired, err)
	}

	err = client.DeleteMyListItem("0", "sm9")
	if err != NotFound {
		t.Fatalf("expected: %v  got: %v", NotFound, err)
	}
}
