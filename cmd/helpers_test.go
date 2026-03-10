package cmd

import (
	"context"
	"fmt"
	"os"
	"testing"

	"workbuddy/internal/note"
)

func TestOpenEditorWithRunner(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		runner      editorRunner
		wantContent string
		wantErr     bool
	}{
		{
			name:    "returns content written by runner",
			initial: "",
			runner: func(path string) error {
				return os.WriteFile(path, []byte("my test note"), 0644)
			},
			wantContent: "my test note",
		},
		{
			name:    "trims surrounding whitespace",
			initial: "",
			runner: func(path string) error {
				return os.WriteFile(path, []byte("  note with spaces  \n"), 0644)
			},
			wantContent: "note with spaces",
		},
		{
			name:    "returns error when content is empty",
			initial: "",
			runner: func(path string) error {
				return nil // writes nothing, file stays empty
			},
			wantErr: true,
		},
		{
			name:    "seeds initial content before editing",
			initial: "existing note",
			runner: func(path string) error {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if string(data) != "existing note" {
					return fmt.Errorf("unexpected seeded content: %q", string(data))
				}
				return os.WriteFile(path, []byte("edited note"), 0644)
			},
			wantContent: "edited note",
		},
		{
			name:    "returns error when runner fails",
			initial: "",
			runner: func(path string) error {
				return fmt.Errorf("editor crashed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := openEditorWithRunner(tt.initial, tt.runner)

			if tt.wantErr {
				if err == nil {
					t.Error("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if content != tt.wantContent {
				t.Errorf("got %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestOpenEditorForContent_NoEditorSet(t *testing.T) {
	t.Setenv("EDITOR", "") // t.Setenv restores the original value after the test

	_, err := openEditorForContent()
	if err == nil {
		t.Fatal("expected an error when $EDITOR is not set, got nil")
	}
}

func TestBuildTagsByNoteID_UsesBatchLookup(t *testing.T) {
	repo := &batchTagRepo{
		tagsByNoteID: map[int64][]string{
			2: {"backend", "work"},
			3: {"personal"},
		},
	}

	notes := []note.Note{
		{ID: 1, Content: "untagged"},
		{ID: 2, Content: "multiple tags"},
		{ID: 3, Content: "single tag"},
	}

	got, err := buildTagsByNoteID(context.Background(), repo, notes)
	if err != nil {
		t.Fatalf("buildTagsByNoteID returned error: %v", err)
	}

	if repo.batchCalls != 1 {
		t.Fatalf("expected exactly one batched tag lookup, got %d", repo.batchCalls)
	}

	if repo.singleCalls != 0 {
		t.Fatalf("expected no per-note tag lookups, got %d", repo.singleCalls)
	}

	wantIDs := []int64{1, 2, 3}
	if len(repo.receivedIDs) != len(wantIDs) {
		t.Fatalf("batched lookup note IDs = %v, want %v", repo.receivedIDs, wantIDs)
	}
	for i, id := range wantIDs {
		if repo.receivedIDs[i] != id {
			t.Fatalf("batched lookup note IDs = %v, want %v", repo.receivedIDs, wantIDs)
		}
	}

	if _, ok := got[1]; ok {
		t.Fatalf("expected note without tags to be omitted from result map, got %v", got[1])
	}

	if fmt.Sprint(got[2]) != fmt.Sprint([]string{"backend", "work"}) {
		t.Fatalf("tags for note 2 = %v, want %v", got[2], []string{"backend", "work"})
	}

	if fmt.Sprint(got[3]) != fmt.Sprint([]string{"personal"}) {
		t.Fatalf("tags for note 3 = %v, want %v", got[3], []string{"personal"})
	}
}

type batchTagRepo struct {
	note.Repository
	tagsByNoteID map[int64][]string
	receivedIDs  []int64
	batchCalls   int
	singleCalls  int
}

func (r *batchTagRepo) ListTagsByNoteIDs(_ context.Context, noteIDs []int64) (map[int64][]string, error) {
	r.batchCalls++
	r.receivedIDs = append([]int64(nil), noteIDs...)

	result := make(map[int64][]string, len(r.tagsByNoteID))
	for noteID, tags := range r.tagsByNoteID {
		result[noteID] = append([]string(nil), tags...)
	}

	return result, nil
}

func (r *batchTagRepo) ListTagsByNoteID(_ context.Context, _ int64) ([]string, error) {
	r.singleCalls++
	return nil, nil
}
