package note

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"
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

func TestCreateNote_TableDriven(t *testing.T) {
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
			name:         "duplicate tags in input are deduplicated",
			content:      "Note with duplicate tags",
			tags:         []string{"work", "work"},
			isTask:       0,
			expectedTags: []string{"work"},
			expectedTask: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			service := NewService(db)
			ctx := context.Background()

			note, err := service.CreateNote(ctx, tt.content, tt.tags, tt.isTask)

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

func TestCreateNote_RejectsEmptyContent(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty string",
			content: "",
		},
		{
			name:    "whitespace only",
			content: "   \n\t  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateNote(ctx, tt.content, []string{"work"}, 0)
			if !errors.Is(err, errEmptyNoteContent) {
				t.Fatalf("expected errEmptyNoteContent, got %v", err)
			}
		})
	}
}

func TestCreateNote_ReusesExistingTag(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	// Create two notes with the same tag
	_, err := service.CreateNote(ctx, "First note", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create first note: %v", err)
	}

	note2, err := service.CreateNote(ctx, "Second note", []string{"work"}, 0)
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

func TestUpdateNoteContent(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createdNote, err := service.CreateNote(ctx, "Original content", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	updatedNote, err := service.UpdateNoteContent(ctx, createdNote.ID, "Updated content")
	if err != nil {
		t.Fatalf("failed to update note: %v", err)
	}

	if updatedNote.ID != createdNote.ID {
		t.Fatalf("updated wrong note: got %d, want %d", updatedNote.ID, createdNote.ID)
	}
	if updatedNote.Content != "Updated content" {
		t.Fatalf("updated content = %q, want %q", updatedNote.Content, "Updated content")
	}

	loadedNote, err := service.GetNoteByID(ctx, createdNote.ID)
	if err != nil {
		t.Fatalf("failed to reload note: %v", err)
	}
	if loadedNote.Content != "Updated content" {
		t.Fatalf("reloaded content = %q, want %q", loadedNote.Content, "Updated content")
	}
}

func TestUpdateNoteContent_RejectsEmptyContent(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	note, err := service.CreateNote(ctx, "Original content", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	_, err = service.UpdateNoteContent(ctx, note.ID, "   ")
	if !errors.Is(err, errEmptyNoteContent) {
		t.Fatalf("expected errEmptyNoteContent, got %v", err)
	}
}

func TestUpdateNote_Tags(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	n, err := service.CreateNote(ctx, "Original content", []string{"work", "urgent"}, 0)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	tests := []struct {
		name   string
		params UpdateParams
		want   []string
	}{
		{
			name: "adds tags without changing content",
			params: UpdateParams{
				AddTags: []string{"backend", "work"},
			},
			want: []string{"backend", "urgent", "work"},
		},
		{
			name: "removes tags",
			params: UpdateParams{
				RemoveTags: []string{"urgent"},
			},
			want: []string{"backend", "work"},
		},
		{
			name: "replaces tags",
			params: UpdateParams{
				SetTags: []string{"personal", "follow-up"},
			},
			want: []string{"follow-up", "personal"},
		},
		{
			name: "set add and remove combine deterministically",
			params: UpdateParams{
				SetTags:    []string{"team", "ops"},
				AddTags:    []string{"urgent"},
				RemoveTags: []string{"ops"},
			},
			want: []string{"team", "urgent"},
		},
	}

	repo := NewRepository(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateNote(ctx, n.ID, tt.params)
			if err != nil {
				t.Fatalf("UpdateNote returned error: %v", err)
			}

			got, err := repo.ListTagsByNoteID(ctx, n.ID)
			if err != nil {
				t.Fatalf("failed to list tags: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("tags = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNote_ContentAndTags(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	n, err := service.CreateNote(ctx, "Original content", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	newContent := "Updated content"
	updatedNote, err := service.UpdateNote(ctx, n.ID, UpdateParams{
		Content: &newContent,
		AddTags: []string{"urgent"},
	})
	if err != nil {
		t.Fatalf("UpdateNote returned error: %v", err)
	}

	if updatedNote.Content != newContent {
		t.Fatalf("updated content = %q, want %q", updatedNote.Content, newContent)
	}

	repo := NewRepository(db)
	gotTags, err := repo.ListTagsByNoteID(ctx, n.ID)
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}
	wantTags := []string{"urgent", "work"}
	if !reflect.DeepEqual(gotTags, wantTags) {
		t.Fatalf("tags = %v, want %v", gotTags, wantTags)
	}
}

func TestBuildContentQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "builds prefix query from words",
			input: "release notes",
			want:  "release* notes*",
		},
		{
			name:  "drops punctuation",
			input: "integrations, roadmap!",
			want:  "integrations* roadmap*",
		},
		{
			name:  "returns empty for blank query",
			input: "   ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildContentQuery(tt.input)
			if got != tt.want {
				t.Fatalf("buildContentQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchNotes_ByContent(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	releaseNote, err := service.CreateNote(ctx, "release planning for integrations", []string{"work"}, 0)
	if err != nil {
		t.Fatalf("failed to create release note: %v", err)
	}
	if _, err := service.CreateNote(ctx, "weekly shopping list", []string{"personal"}, 0); err != nil {
		t.Fatalf("failed to create personal note: %v", err)
	}
	pendingTask, err := service.CreateNote(ctx, "finish integrations documentation", []string{"work"}, 1)
	if err != nil {
		t.Fatalf("failed to create pending task: %v", err)
	}
	completedTask, err := service.CreateNote(ctx, "finish release notes", []string{"work"}, 1)
	if err != nil {
		t.Fatalf("failed to create completed task: %v", err)
	}

	repo := NewRepository(db)
	rows, err := repo.MarkNoteCompleted(ctx, completedTask.ID)
	if err != nil {
		t.Fatalf("failed to mark task completed: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected to mark 1 task completed, got %d", rows)
	}

	tests := []struct {
		name   string
		params SearchParams
		want   []int64
	}{
		{
			name: "matches exact term anywhere in content",
			params: SearchParams{
				Query: "integrations",
				Limit: 10,
			},
			want: []int64{pendingTask.ID, releaseNote.ID},
		},
		{
			name: "matches prefix searches",
			params: SearchParams{
				Query: "integ",
				Limit: 10,
			},
			want: []int64{pendingTask.ID, releaseNote.ID},
		},
		{
			name: "requires all terms to match",
			params: SearchParams{
				Query: "release notes",
				Limit: 10,
			},
			want: []int64{completedTask.ID},
		},
		{
			name: "combines content and tag filters",
			params: SearchParams{
				Query: "release",
				Tag:   "work",
				Limit: 10,
			},
			want: []int64{completedTask.ID, releaseNote.ID},
		},
		{
			name: "supports pending task filter",
			params: SearchParams{
				Query:     "finish",
				TasksOnly: true,
				Completed: boolPtr(false),
				Limit:     10,
			},
			want: []int64{pendingTask.ID},
		},
		{
			name: "supports completed task filter",
			params: SearchParams{
				Query:     "finish",
				Completed: boolPtr(true),
				Limit:     10,
			},
			want: []int64{completedTask.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notes, err := service.SearchNotes(ctx, tt.params)
			if err != nil {
				t.Fatalf("SearchNotes returned error: %v", err)
			}

			got := make([]int64, 0, len(notes))
			for _, note := range notes {
				got = append(got, note.ID)
			}
			slices.Sort(got)

			want := append([]int64(nil), tt.want...)
			slices.Sort(want)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("SearchNotes IDs = %v, want %v", got, want)
			}
		})
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func TestCreateNoteWithParams_RecurringValidation(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	dueDate := mustDate(t, "2026-03-16")
	weeklyRule := RecurrenceWeekly
	monthlyRule := RecurrenceMonthly

	tests := []struct {
		name    string
		params  CreateParams
		wantErr error
	}{
		{
			name: "rejects recurrence on non-task notes",
			params: CreateParams{
				Content:        "Plan the week",
				RecurrenceRule: weeklyRule,
				DueDate:        &dueDate,
				RecurrenceWeekday: func() *time.Weekday {
					day := time.Monday
					return &day
				}(),
			},
			wantErr: errRecurringRequiresTask,
		},
		{
			name: "requires due date for recurring tasks",
			params: CreateParams{
				Content:        "Daily standup",
				IsTask:         true,
				RecurrenceRule: RecurrenceDaily,
			},
			wantErr: errRecurringRequiresDueDate,
		},
		{
			name: "weekly recurrence requires weekday",
			params: CreateParams{
				Content:        "Weekly planning",
				IsTask:         true,
				DueDate:        &dueDate,
				RecurrenceRule: weeklyRule,
			},
			wantErr: errWeeklyRequiresWeekday,
		},
		{
			name: "monthly recurrence requires day of month",
			params: CreateParams{
				Content:        "Close books",
				IsTask:         true,
				DueDate:        dueDatePtr(t, "2026-03-31"),
				RecurrenceRule: monthlyRule,
			},
			wantErr: errMonthlyRequiresDayOfMonth,
		},
		{
			name: "weekly recurrence validates due weekday",
			params: CreateParams{
				Content:        "Weekly planning",
				IsTask:         true,
				DueDate:        dueDatePtr(t, "2026-03-17"),
				RecurrenceRule: weeklyRule,
				RecurrenceWeekday: func() *time.Weekday {
					day := time.Monday
					return &day
				}(),
			},
			wantErr: errDueDateWeekdayMismatch,
		},
		{
			name: "monthly recurrence validates due day",
			params: CreateParams{
				Content:              "Close books",
				IsTask:               true,
				DueDate:              dueDatePtr(t, "2026-03-30"),
				RecurrenceRule:       monthlyRule,
				RecurrenceDayOfMonth: intPtr(31),
			},
			wantErr: errDueDateDayOfMonthMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateNoteWithParams(ctx, tt.params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateNoteWithParams_CreatesRecurringTasks(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	tests := []struct {
		name   string
		params CreateParams
	}{
		{
			name: "creates daily recurring task",
			params: CreateParams{
				Content:        "Daily standup",
				Tags:           []string{"work", "team"},
				IsTask:         true,
				DueDate:        dueDatePtr(t, "2026-03-16"),
				RecurrenceRule: RecurrenceDaily,
			},
		},
		{
			name: "creates weekly recurring task",
			params: CreateParams{
				Content:           "Weekly planning",
				Tags:              []string{"work"},
				IsTask:            true,
				DueDate:           dueDatePtr(t, "2026-03-16"),
				RecurrenceRule:    RecurrenceWeekly,
				RecurrenceWeekday: weekdayPtr(time.Monday),
			},
		},
		{
			name: "creates monthly recurring task",
			params: CreateParams{
				Content:              "Monthly close",
				Tags:                 []string{"finance"},
				IsTask:               true,
				DueDate:              dueDatePtr(t, "2026-03-31"),
				RecurrenceRule:       RecurrenceMonthly,
				RecurrenceDayOfMonth: intPtr(31),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := service.CreateNoteWithParams(ctx, tt.params)
			if err != nil {
				t.Fatalf("CreateNoteWithParams returned error: %v", err)
			}

			if !created.TaskSeriesID.Valid {
				t.Fatal("expected recurring task to be linked to a task series")
			}
			if !created.DueAt.Valid {
				t.Fatal("expected recurring task to have a due date")
			}
			if created.RecurrenceRule.String != string(tt.params.RecurrenceRule) {
				t.Fatalf("recurrence rule = %q, want %q", created.RecurrenceRule.String, tt.params.RecurrenceRule)
			}

			assertTaskSeriesState(t, db, created.TaskSeriesID.Int64, taskSeriesExpectation{
				Content:        tt.params.Content,
				Rule:           string(tt.params.RecurrenceRule),
				Weekday:        weekdayInt(tt.params.RecurrenceWeekday),
				DayOfMonth:     int64Ptr(tt.params.RecurrenceDayOfMonth),
				Active:         true,
				ExpectedTags:   normalizeTags(tt.params.Tags),
				PendingNoteID:  created.ID,
				PendingDueDate: tt.params.DueDate.Format(dateLayout),
			})
		})
	}
}

func TestUpdateNote_RecurringSeriesSyncsPendingOccurrence(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	created, err := service.CreateNoteWithParams(ctx, CreateParams{
		Content:        "Weekly planning",
		Tags:           []string{"work"},
		IsTask:         true,
		DueDate:        dueDatePtr(t, "2026-03-16"),
		RecurrenceRule: RecurrenceWeekly,
		RecurrenceWeekday: func() *time.Weekday {
			day := time.Monday
			return &day
		}(),
	})
	if err != nil {
		t.Fatalf("failed to create recurring task: %v", err)
	}

	newContent := "Weekly planning and staffing"
	newRule := RecurrenceMonthly
	updated, err := service.UpdateNote(ctx, created.ID, UpdateParams{
		Content:              &newContent,
		AddTags:              []string{"urgent"},
		DueDate:              dueDatePtr(t, "2026-03-31"),
		RecurrenceRule:       &newRule,
		RecurrenceDayOfMonth: intPtr(31),
	})
	if err != nil {
		t.Fatalf("UpdateNote returned error: %v", err)
	}

	if updated.Content != newContent {
		t.Fatalf("updated content = %q, want %q", updated.Content, newContent)
	}
	if updated.RecurrenceRule.String != string(newRule) {
		t.Fatalf("updated rule = %q, want %q", updated.RecurrenceRule.String, newRule)
	}
	if got := updated.DueAt.Time.Format(dateLayout); got != "2026-03-31" {
		t.Fatalf("updated due date = %q, want %q", got, "2026-03-31")
	}

	assertTaskSeriesState(t, db, created.TaskSeriesID.Int64, taskSeriesExpectation{
		Content:        newContent,
		Rule:           string(newRule),
		DayOfMonth:     int64Ptr(intPtr(31)),
		Active:         true,
		ExpectedTags:   []string{"urgent", "work"},
		PendingNoteID:  created.ID,
		PendingDueDate: "2026-03-31",
	})
}

func TestCompleteTask_RecurringTasksGenerateOneNextOccurrence(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	nonRecurring, err := service.CreateNoteWithParams(ctx, CreateParams{
		Content: "Ship release",
		Tags:    []string{"ops"},
		IsTask:  true,
		DueDate: dueDatePtr(t, "2026-03-16"),
	})
	if err != nil {
		t.Fatalf("failed to create non-recurring task: %v", err)
	}

	result, err := service.CompleteTask(ctx, nonRecurring.ID)
	if err != nil {
		t.Fatalf("CompleteTask returned error: %v", err)
	}
	if result.Next != nil {
		t.Fatal("expected non-recurring task completion to not create a successor")
	}

	recurring, err := service.CreateNoteWithParams(ctx, CreateParams{
		Content:        "Daily standup",
		Tags:           []string{"team"},
		IsTask:         true,
		DueDate:        dueDatePtr(t, "2026-03-10"),
		RecurrenceRule: RecurrenceDaily,
	})
	if err != nil {
		t.Fatalf("failed to create recurring task: %v", err)
	}

	completeRecurring, err := service.CompleteTask(ctx, recurring.ID)
	if err != nil {
		t.Fatalf("CompleteTask returned error: %v", err)
	}
	if completeRecurring.Next == nil {
		t.Fatal("expected recurring task completion to create a successor")
	}
	if got := completeRecurring.Next.DueAt.Time.Format(dateLayout); got != "2026-03-11" {
		t.Fatalf("next due date = %q, want %q", got, "2026-03-11")
	}

	secondAttempt, err := service.CompleteTask(ctx, recurring.ID)
	if err == nil {
		t.Fatalf("expected second completion attempt to fail, got result=%+v", secondAttempt)
	}

	assertPendingCount(t, db, recurring.TaskSeriesID.Int64, 1)
	assertTaskSeriesState(t, db, recurring.TaskSeriesID.Int64, taskSeriesExpectation{
		Content:        "Daily standup",
		Rule:           string(RecurrenceDaily),
		Active:         true,
		ExpectedTags:   []string{"team"},
		PendingNoteID:  completeRecurring.Next.ID,
		PendingDueDate: "2026-03-11",
	})
}

func TestCompleteTask_MonthlyRolloverClampsToMonthEnd(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	recurring, err := service.CreateNoteWithParams(ctx, CreateParams{
		Content:              "Monthly close",
		IsTask:               true,
		DueDate:              dueDatePtr(t, "2026-01-31"),
		RecurrenceRule:       RecurrenceMonthly,
		RecurrenceDayOfMonth: intPtr(31),
	})
	if err != nil {
		t.Fatalf("failed to create recurring task: %v", err)
	}

	first, err := service.CompleteTask(ctx, recurring.ID)
	if err != nil {
		t.Fatalf("failed first completion: %v", err)
	}
	if got := first.Next.DueAt.Time.Format(dateLayout); got != "2026-02-28" {
		t.Fatalf("first rollover due date = %q, want %q", got, "2026-02-28")
	}

	second, err := service.CompleteTask(ctx, first.Next.ID)
	if err != nil {
		t.Fatalf("failed second completion: %v", err)
	}
	if got := second.Next.DueAt.Time.Format(dateLayout); got != "2026-03-31" {
		t.Fatalf("second rollover due date = %q, want %q", got, "2026-03-31")
	}
}

func TestRemoveNote_RecurringSeriesPreservesCompletedHistory(t *testing.T) {
	db := newTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	recurring, err := service.CreateNoteWithParams(ctx, CreateParams{
		Content:        "Weekly planning",
		Tags:           []string{"work"},
		IsTask:         true,
		DueDate:        dueDatePtr(t, "2026-03-16"),
		RecurrenceRule: RecurrenceWeekly,
		RecurrenceWeekday: func() *time.Weekday {
			day := time.Monday
			return &day
		}(),
	})
	if err != nil {
		t.Fatalf("failed to create recurring task: %v", err)
	}

	completion, err := service.CompleteTask(ctx, recurring.ID)
	if err != nil {
		t.Fatalf("failed to complete recurring task: %v", err)
	}

	removed, err := service.RemoveNote(ctx, completion.Next.ID)
	if err != nil {
		t.Fatalf("RemoveNote returned error: %v", err)
	}
	if !removed.RemovedSeries {
		t.Fatal("expected recurring series removal to be reported")
	}

	assertTaskSeriesState(t, db, recurring.TaskSeriesID.Int64, taskSeriesExpectation{
		Content:      "Weekly planning",
		Rule:         string(RecurrenceWeekly),
		Weekday:      int64Ptr(intPtr(int(time.Monday))),
		Active:       false,
		ExpectedTags: []string{"work"},
	})
	assertPendingCount(t, db, recurring.TaskSeriesID.Int64, 0)

	history, err := service.GetNoteByID(ctx, recurring.ID)
	if err != nil {
		t.Fatalf("failed to reload completed occurrence: %v", err)
	}
	if !history.CompletedAt.Valid {
		t.Fatal("expected completed recurring occurrence to remain in history")
	}
}

const dateLayout = "2006-01-02"

type taskSeriesExpectation struct {
	Content        string
	Rule           string
	Weekday        *int64
	DayOfMonth     *int64
	Active         bool
	ExpectedTags   []string
	PendingNoteID  int64
	PendingDueDate string
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", value, err)
	}
	return parsed
}

func dueDatePtr(t *testing.T, value string) *time.Time {
	t.Helper()
	parsed := mustDate(t, value)
	return &parsed
}

func weekdayPtr(day time.Weekday) *time.Weekday {
	return &day
}

func intPtr(v int) *int {
	return &v
}

func weekdayInt(day *time.Weekday) *int64 {
	if day == nil {
		return nil
	}
	value := int64(*day)
	return &value
}

func int64Ptr(value *int) *int64 {
	if value == nil {
		return nil
	}
	out := int64(*value)
	return &out
}

func assertTaskSeriesState(t *testing.T, db *sql.DB, seriesID int64, want taskSeriesExpectation) {
	t.Helper()

	row := db.QueryRowContext(t.Context(), `
SELECT content, recurrence_rule, recurrence_weekday, recurrence_day_of_month, active
FROM task_series
WHERE id = ?`, seriesID)

	var (
		content    string
		rule       string
		weekday    sql.NullInt64
		dayOfMonth sql.NullInt64
		active     int64
	)
	if err := row.Scan(&content, &rule, &weekday, &dayOfMonth, &active); err != nil {
		t.Fatalf("failed to load task_series %d: %v", seriesID, err)
	}

	if content != want.Content {
		t.Fatalf("series content = %q, want %q", content, want.Content)
	}
	if rule != want.Rule {
		t.Fatalf("series rule = %q, want %q", rule, want.Rule)
	}
	if got, expected := nullInt64Ptr(weekday), want.Weekday; !reflect.DeepEqual(got, expected) {
		t.Fatalf("series weekday = %v, want %v", got, expected)
	}
	if got, expected := nullInt64Ptr(dayOfMonth), want.DayOfMonth; !reflect.DeepEqual(got, expected) {
		t.Fatalf("series day_of_month = %v, want %v", got, expected)
	}
	if got := active == 1; got != want.Active {
		t.Fatalf("series active = %v, want %v", got, want.Active)
	}

	tagRows, err := db.QueryContext(t.Context(), `
SELECT t.name
FROM task_series_tags tst
INNER JOIN tags t ON t.id = tst.tag_id
WHERE tst.task_series_id = ?
ORDER BY t.name`, seriesID)
	if err != nil {
		t.Fatalf("failed to load task_series tags: %v", err)
	}
	defer tagRows.Close()

	var gotTags []string
	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			t.Fatalf("failed to scan task_series tag: %v", err)
		}
		gotTags = append(gotTags, tag)
	}
	if err := tagRows.Err(); err != nil {
		t.Fatalf("failed to iterate task_series tags: %v", err)
	}

	if !reflect.DeepEqual(gotTags, want.ExpectedTags) {
		t.Fatalf("series tags = %v, want %v", gotTags, want.ExpectedTags)
	}

	if want.PendingNoteID == 0 {
		return
	}

	pendingRow := db.QueryRowContext(t.Context(), `
SELECT id, due_at
FROM note
WHERE task_series_id = ? AND completed_at IS NULL`, seriesID)
	var pendingID int64
	var dueAt sql.NullTime
	if err := pendingRow.Scan(&pendingID, &dueAt); err != nil {
		t.Fatalf("failed to load pending occurrence for task_series %d: %v", seriesID, err)
	}
	if pendingID != want.PendingNoteID {
		t.Fatalf("pending note id = %d, want %d", pendingID, want.PendingNoteID)
	}
	if !dueAt.Valid {
		t.Fatal("expected pending occurrence to have a due date")
	}
	if got := dueAt.Time.Format(dateLayout); got != want.PendingDueDate {
		t.Fatalf("pending due date = %q, want %q", got, want.PendingDueDate)
	}
}

func assertPendingCount(t *testing.T, db *sql.DB, seriesID int64, want int) {
	t.Helper()

	row := db.QueryRowContext(t.Context(), `
SELECT COUNT(*)
FROM note
WHERE task_series_id = ? AND completed_at IS NULL`, seriesID)

	var got int
	if err := row.Scan(&got); err != nil {
		t.Fatalf("failed to count pending occurrences: %v", err)
	}
	if got != want {
		t.Fatalf("pending occurrences = %d, want %d", got, want)
	}
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	value := v.Int64
	return &value
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

func (m *mockRepositoryWithNotes) DeleteAllTagsFromNote(ctx context.Context, noteID int64) error {
	return nil
}

func (m *mockRepositoryWithNotes) DeleteNoteByID(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) DeleteTaskByID(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) GetNoteByID(ctx context.Context, id int64) (Note, error) {
	return Note{}, nil
}

func (m *mockRepositoryWithNotes) ListTags(ctx context.Context, pattern string) ([]Tag, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) ListTagsByNoteID(ctx context.Context, noteID int64) ([]string, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) ListTagsByNoteIDs(ctx context.Context, noteIDs []int64) (map[int64][]string, error) {
	return nil, nil
}

func (m *mockRepositoryWithNotes) MarkNoteCompleted(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockRepositoryWithNotes) UpdateNoteContentByID(ctx context.Context, arg UpdateNoteContentByIDParams) (Note, error) {
	return Note{}, nil
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

func (m *mockRepositoryWithNotes) SearchNotesByTagAndContent(ctx context.Context, arg SearchNotesByTagAndContentParams) ([]Note, error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	return m.notesToReturn, nil
}

func (m *mockRepositoryWithNotes) SearchNotesByContent(ctx context.Context, arg SearchNotesByContentParams) ([]Note, error) {
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
