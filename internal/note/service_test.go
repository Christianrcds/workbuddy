package note

import (
	"context"
	"errors"
	"testing"
)

func TestGetNotesByTag(t *testing.T) {
	testNotes := []Note{
		{ID: 1, Content: "Learning Go"},
		{ID: 2, Content: "Go interfaces"},
	}

	mockRepo := &mockRepositoryWithNotes{
		notesToReturn: testNotes,
	}

	service := &Service{
		db:   nil,
		repo: mockRepo,
	}

	notes, err := service.SearchNotes(context.Background(), SearchParams{Tag: "", Limit: 10})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(notes) != 2 {
		t.Errorf("Expected 2 notes, got %d", len(notes))
	}

	if notes[0].Content != "Learning Go" {
		t.Errorf("Expected first note content 'Learning Go', got '%s'", notes[0].Content)
	}
}

func TestGetNotesByTag_TableDriven(t *testing.T) {
	var errRepoFailure = errors.New("repository failure")

	tests := []struct {
		name          string
		tag           string
		limit         int
		mockNotes     []Note
		expectedCount int
		expectError   bool
		errorToReturn error
	}{
		{
			name:  "finds multiple notes",
			tag:   "golang",
			limit: 10,
			mockNotes: []Note{
				{ID: 1, Content: "Learning Go"},
				{ID: 2, Content: "Go is great"},
			},
			expectedCount: 2,
			expectError:   false,
			errorToReturn: nil,
		},
		{
			name:          "finds no notes",
			tag:           "nonexistent",
			limit:         10,
			mockNotes:     []Note{},
			expectedCount: 0,
			expectError:   false,
			errorToReturn: nil,
		},
		{
			name:  "respects limit",
			tag:   "popular",
			limit: 1,
			mockNotes: []Note{
				{ID: 1, Content: "First note"},
			},
			expectedCount: 1,
			expectError:   false,
			errorToReturn: nil,
		},
		{
			name:          "Throws error",
			tag:           "popular",
			limit:         1,
			mockNotes:     []Note{},
			expectedCount: 0,
			expectError:   true,
			errorToReturn: errRepoFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepositoryWithNotes{
				notesToReturn: tt.mockNotes,
				errorToReturn: tt.errorToReturn,
			}
			service := &Service{repo: mockRepo}

			notes, err := service.SearchNotes(context.Background(), SearchParams{Tag: tt.tag, Limit: int64(tt.limit)})

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				if !errors.Is(err, tt.errorToReturn) {
					t.Fatalf("expected error %v, got %v", tt.errorToReturn, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if len(notes) != tt.expectedCount {
				t.Errorf("Expected %d notes, got %d", tt.expectedCount, len(notes))
			}
		})
	}
}

func TestCreateNoteWithTags_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		tags         []string
		isTask       int64
		expectError  bool
		expectedTags []string // must be in alphabetical order (ListTagsByNoteID sorts by name)
		expectedTask int64
	}{
		{
			name:         "creates note with single tag",
			content:      "Buy milk",
			tags:         []string{"shopping"},
			isTask:       0,
			expectedTags: []string{"shopping"},
			expectedTask: 0,
		},
		{
			name:         "creates note with multiple tags",
			content:      "Team standup",
			tags:         []string{"work", "daily"},
			isTask:       0,
			expectedTags: []string{"daily", "work"}, // alphabetical order
			expectedTask: 0,
		},
		{
			name:         "creates note with no tags",
			content:      "Just a thought",
			tags:         []string{},
			isTask:       0,
			expectedTags: []string{},
			expectedTask: 0,
		},
		{
			name:         "creates a task",
			content:      "Finish the report",
			tags:         []string{"work"},
			isTask:       1,
			expectedTags: []string{"work"},
			expectedTask: 1,
		},
		{
			name:         "trims whitespace from tags",
			content:      "Note with padded tags",
			tags:         []string{" work ", " personal "},
			isTask:       0,
			expectedTags: []string{"personal", "work"}, // trimmed and alphabetical
			expectedTask: 0,
		},
		{
			// This exposes a current limitation: the service does not deduplicate
			// tags before processing, so inserting the same tag twice violates
			// the PRIMARY KEY constraint on note_tags and causes a rollback.
			name:        "duplicate tags in input cause error",
			content:     "Note with duplicate tags",
			tags:        []string{"work", "work"},
			isTask:      0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			service := NewService(db)
			ctx := context.Background()

			note, err := service.CreateNoteWithTags(ctx, tt.content, tt.tags, tt.isTask)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if note.Content != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, note.Content)
			}
			if note.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if note.IsTask != tt.expectedTask {
				t.Errorf("expected isTask %d, got %d", tt.expectedTask, note.IsTask)
			}

			repo := NewRepository(db)
			tags, err := repo.ListTagsByNoteID(ctx, note.ID)
			if err != nil {
				t.Fatalf("failed to list tags: %v", err)
			}
			if len(tags) != len(tt.expectedTags) {
				t.Fatalf("expected %d tags, got %d: %v", len(tt.expectedTags), len(tags), tags)
			}
			for i, tag := range tags {
				if tag != tt.expectedTags[i] {
					t.Errorf("tag[%d]: expected %q, got %q", i, tt.expectedTags[i], tag)
				}
			}
		})
	}
}

func TestCreateNoteWithTags_ReusesExistingTag(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create two notes with the same tag
	_, err := service.CreateNoteWithTags(ctx, "First note", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create first note: %v", err)
	}

	note2, err := service.CreateNoteWithTags(ctx, "Second note", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create second note: %v", err)
	}

	// The second note should still be tagged with "work"
	repo := NewRepository(db)
	tags, err := repo.ListTagsByNoteID(ctx, note2.ID)
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "work" {
		t.Errorf("expected tag 'work', got %v", tags)
	}

	// Most importantly: only ONE "work" tag should exist in the database
	// This verifies the get-or-create logic didn't duplicate the tag
	allTags, err := repo.ListTags(ctx, "work")
	if err != nil {
		t.Fatalf("failed to list all tags: %v", err)
	}
	if len(allTags) != 1 {
		t.Errorf("expected 1 'work' tag in DB, got %d — tag was duplicated!", len(allTags))
	}
}

// Mock implementations

type mockRepositoryWithNotes struct {
	notesToReturn []Note
	errorToReturn error
}

func (m *mockRepositoryWithNotes) CreateNote(ctx context.Context, content CreateNoteParams) (Note, error) {
	return Note{}, nil
}

func (m *mockRepositoryWithNotes) ListNotes(ctx context.Context) ([]Note, error) {
	return m.notesToReturn, nil
}

func (m *mockRepositoryWithNotes) CreateTag(ctx context.Context, name string) (Tag, error) {
	return Tag{}, nil
}

func (m *mockRepositoryWithNotes) GetTag(ctx context.Context, name string) (Tag, error) {
	return Tag{}, nil
}

func (m *mockRepositoryWithNotes) AddTagToNote(ctx context.Context, arg AddTagToNoteParams) error {
	return nil
}

func (m *mockRepositoryWithNotes) DeleteNoteByID(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) DeleteTaskByID(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) ListTags(ctx context.Context, pattern string) ([]Tag, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) ListTagsByNoteID(ctx context.Context, noteID int64) ([]string, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) MarkNoteCompleted(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) ListTasks(ctx context.Context) ([]Note, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) SearchNotesByTag(ctx context.Context, arg SearchNotesByTagParams) ([]Note, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	return m.notesToReturn, nil
}

func (m *mockRepositoryWithNotes) SearchNotes(ctx context.Context, arg SearchNotesParams) ([]Note, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	return m.notesToReturn, nil
}

