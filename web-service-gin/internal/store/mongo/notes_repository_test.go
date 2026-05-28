package mongostore

import (
	"reflect"
	"testing"
	"time"

	"example/web-service-gin/internal/notes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestOwnedNoteFilterScopesByUserAndID(t *testing.T) {
	id := bson.NewObjectID()

	filter := ownedNoteFilter("user-1", id)

	if filter["userId"] != "user-1" {
		t.Fatalf("expected user filter, got %#v", filter)
	}

	if filter["_id"] != id {
		t.Fatalf("expected id filter, got %#v", filter)
	}
}

func TestListFilterScopesByUser(t *testing.T) {
	filter := listFilter("user-1")

	if !reflect.DeepEqual(filter, bson.M{"userId": "user-1"}) {
		t.Fatalf("unexpected list filter: %#v", filter)
	}
}

func TestListSortPrioritizesPinnedThenUpdatedTime(t *testing.T) {
	sort := listSort()
	expected := bson.D{
		{Key: "pinned", Value: -1},
		{Key: "updatedAt", Value: -1},
		{Key: "_id", Value: -1},
	}

	if !reflect.DeepEqual(sort, expected) {
		t.Fatalf("unexpected list sort: %#v", sort)
	}
}

func TestUpdateSetIncludesOnlyRequestedFields(t *testing.T) {
	title := "Title"
	pinned := true
	updatedAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)

	set := updateSet(notes.NoteUpdate{
		Title:     &title,
		Pinned:    &pinned,
		UpdatedAt: updatedAt,
	})

	if set["title"] != "Title" {
		t.Fatalf("expected title update, got %#v", set)
	}

	if set["pinned"] != true {
		t.Fatalf("expected pinned update, got %#v", set)
	}

	if set["updatedAt"] != updatedAt {
		t.Fatalf("expected updatedAt update, got %#v", set)
	}

	if _, ok := set["content"]; ok {
		t.Fatalf("did not expect content update, got %#v", set)
	}
}
