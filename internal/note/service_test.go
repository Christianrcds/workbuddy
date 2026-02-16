package note

import (
	"context"
	"database/sql"
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

	notes, err := service.GetNotesByTag(context.Background(), "", 10)

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

			notes, err := service.GetNotesByTag(context.Background(), tt.tag, tt.limit)

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

func (m *mockRepositoryWithNotes) GetNotesByTagWithLimit(ctx context.Context, arg GetNotesByTagWithLimitParams) ([]Note, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	return m.notesToReturn, nil
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

type mockRepository struct {
	createNoteCalled bool
	createNoteArg    string
	createNoteReturn Note
	createNoteError  error
	tags             map[string]Tag
}

func (m *mockRepository) CreateNote(ctx context.Context, content string) (Note, error) {
	m.createNoteCalled = true
	m.createNoteArg = content

	if m.createNoteError != nil {
		return Note{}, m.createNoteError
	}

	return m.createNoteReturn, nil
}

func (m *mockRepository) ListNotes(ctx context.Context) ([]Note, error) {
	return nil, nil
}

func (m *mockRepository) CreateTag(ctx context.Context, name string) (Tag, error) {
	tag := Tag{
		ID:   int64(len(m.tags) + 1),
		Name: name,
	}
	m.tags[name] = tag
	return tag, nil
}

func (m *mockRepository) GetTag(ctx context.Context, name string) (Tag, error) {
	tag, exists := m.tags[name]
	if !exists {
		return Tag{}, sql.ErrNoRows
	}
	return tag, nil
}

func (m *mockRepository) AddTagToNote(ctx context.Context, arg AddTagToNoteParams) error {
	return nil
}

func (m *mockRepository) GetNotesByTagWithLimit(ctx context.Context, arg GetNotesByTagWithLimitParams) ([]Note, error) {
	return nil, nil
}

func (m *mockRepository) ListTags(ctx context.Context, pattern string) ([]Tag, error) {
	return nil, nil
}
