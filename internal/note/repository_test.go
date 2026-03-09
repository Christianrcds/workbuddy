package note

import (
	"context"
	"reflect"
	"testing"
)

func TestRepository_ListTagsByNoteIDs(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	repo := NewRepository(db)
	ctx := context.Background()

	noteWithoutTags, err := service.CreateNote(ctx, "untagged", nil, 0)
	if err != nil {
		t.Fatalf("failed to create untagged note: %v", err)
	}

	noteWithOneTag, err := service.CreateNote(ctx, "single tag", []string{"personal"}, 0)
	if err != nil {
		t.Fatalf("failed to create note with one tag: %v", err)
	}

	noteWithManyTags, err := service.CreateNote(ctx, "many tags", []string{"work", "backend"}, 0)
	if err != nil {
		t.Fatalf("failed to create note with many tags: %v", err)
	}

	got, err := repo.ListTagsByNoteIDs(ctx, []int64{
		noteWithoutTags.ID,
		noteWithOneTag.ID,
		noteWithManyTags.ID,
	})
	if err != nil {
		t.Fatalf("ListTagsByNoteIDs returned error: %v", err)
	}

	if _, ok := got[noteWithoutTags.ID]; ok {
		t.Fatalf("expected untagged note to be omitted from result map, got %v", got[noteWithoutTags.ID])
	}

	if want := []string{"personal"}; !reflect.DeepEqual(got[noteWithOneTag.ID], want) {
		t.Fatalf("tags for note %d = %v, want %v", noteWithOneTag.ID, got[noteWithOneTag.ID], want)
	}

	if want := []string{"backend", "work"}; !reflect.DeepEqual(got[noteWithManyTags.ID], want) {
		t.Fatalf("tags for note %d = %v, want %v", noteWithManyTags.ID, got[noteWithManyTags.ID], want)
	}
}
