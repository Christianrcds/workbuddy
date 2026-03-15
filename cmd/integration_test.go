package cmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"workbuddy/internal/note"

	_ "modernc.org/sqlite"
)

func TestRunMigrations_Idempotent(t *testing.T) {
	repoRoot := projectRoot(t)
	SetMigrations(os.DirFS(repoRoot))

	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")
	t.Setenv("WORKBUDDY_DB", dbPath)

	if err := runMigrations(); err != nil {
		t.Fatalf("first runMigrations() failed: %v", err)
	}

	if err := runMigrations(); err != nil {
		t.Fatalf("second runMigrations() failed: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open migrated database: %v", err)
	}
	defer db.Close()

	columns := make(map[string]bool)
	rows, err := db.Query("PRAGMA table_info(note);")
	if err != nil {
		t.Fatalf("failed to inspect note table: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultVal sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			t.Fatalf("failed to scan note column: %v", err)
		}
		columns[name] = true
	}

	if !columns["completed_at"] {
		t.Fatal("expected completed_at column to exist after migrations")
	}
	if !columns["is_task"] {
		t.Fatal("expected is_task column to exist after migrations")
	}
	if !columns["due_at"] {
		t.Fatal("expected due_at column to exist after migrations")
	}
	if !columns["task_series_id"] {
		t.Fatal("expected task_series_id column to exist after migrations")
	}
	if !columns["recurrence_rule"] {
		t.Fatal("expected recurrence_rule column to exist after migrations")
	}
	if !columns["recurrence_weekday"] {
		t.Fatal("expected recurrence_weekday column to exist after migrations")
	}
	if !columns["recurrence_day_of_month"] {
		t.Fatal("expected recurrence_day_of_month column to exist after migrations")
	}

	assertTableExists(t, db, "task_series")
	assertTableExists(t, db, "task_series_tags")
	assertIndexExists(t, db, "idx_note_pending_task_series")
}

func TestCLI_CreateNormalizesContentAndDeduplicatesTags(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")

	result := runCLI(t, dbPath, "", "create", "-c", "  Ship release notes  ", "-t", "work,work,urgent, work ")
	if result.exitCode != 0 {
		t.Fatalf("create command failed with exit code %d\nstdout: %s\nstderr: %s", result.exitCode, result.stdout, result.stderr)
	}

	if !strings.Contains(result.stdout, "Created note with ID 1") {
		t.Fatalf("expected success output, got stdout=%q stderr=%q", result.stdout, result.stderr)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repo := note.NewRepository(db)
	createdNote, err := repo.GetNoteByID(t.Context(), 1)
	if err != nil {
		t.Fatalf("failed to load created note: %v", err)
	}
	if createdNote.Content != "Ship release notes" {
		t.Fatalf("created note content = %q, want %q", createdNote.Content, "Ship release notes")
	}

	tags, err := repo.ListTagsByNoteID(t.Context(), createdNote.ID)
	if err != nil {
		t.Fatalf("failed to load created note tags: %v", err)
	}
	if fmt.Sprint(tags) != fmt.Sprint([]string{"urgent", "work"}) {
		t.Fatalf("created note tags = %v, want %v", tags, []string{"urgent", "work"})
	}
}

func TestCLI_NoteAndTaskWorkflows(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")

	createNote := runCLI(t, dbPath, "", "create", "-c", "Draft project brief", "-t", "work")
	if createNote.exitCode != 0 {
		t.Fatalf("create note failed: stdout=%s stderr=%s", createNote.stdout, createNote.stderr)
	}

	updateNote := runCLI(t, dbPath, "", "update", "1", "-c", "Draft launch brief", "--add-tag", "urgent")
	if updateNote.exitCode != 0 {
		t.Fatalf("update note failed: stdout=%s stderr=%s", updateNote.stdout, updateNote.stderr)
	}

	searchNote := runCLI(t, dbPath, "", "search", "launch", "-t", "urgent")
	if searchNote.exitCode != 0 {
		t.Fatalf("search note failed: stdout=%s stderr=%s", searchNote.stdout, searchNote.stderr)
	}
	if !strings.Contains(searchNote.stdout, "Draft launch brief") {
		t.Fatalf("expected updated note content in search output, got %q", searchNote.stdout)
	}
	if !strings.Contains(searchNote.stdout, "urgent") || !strings.Contains(searchNote.stdout, "work") {
		t.Fatalf("expected note tags in search output, got %q", searchNote.stdout)
	}

	createTask := runCLI(t, dbPath, "", "create", "-c", "Ship release", "--task", "-t", "ops")
	if createTask.exitCode != 0 {
		t.Fatalf("create task failed: stdout=%s stderr=%s", createTask.stdout, createTask.stderr)
	}

	listTasks := runCLI(t, dbPath, "", "list", "tasks")
	if listTasks.exitCode != 0 {
		t.Fatalf("list tasks failed: stdout=%s stderr=%s", listTasks.stdout, listTasks.stderr)
	}
	if !strings.Contains(listTasks.stdout, "Ship release") {
		t.Fatalf("expected task content in list output, got %q", listTasks.stdout)
	}

	checkTask := runCLI(t, dbPath, "", "check", "2")
	if checkTask.exitCode != 0 {
		t.Fatalf("check task failed: stdout=%s stderr=%s", checkTask.stdout, checkTask.stderr)
	}

	completedTasks := runCLI(t, dbPath, "", "search", "--tasks", "--completed")
	if completedTasks.exitCode != 0 {
		t.Fatalf("search completed tasks failed: stdout=%s stderr=%s", completedTasks.stdout, completedTasks.stderr)
	}
	if !strings.Contains(completedTasks.stdout, "Ship release") {
		t.Fatalf("expected completed task in search output, got %q", completedTasks.stdout)
	}

	removeTask := runCLI(t, dbPath, "y\n", "remove", "2")
	if removeTask.exitCode != 0 {
		t.Fatalf("remove task failed: stdout=%s stderr=%s", removeTask.stdout, removeTask.stderr)
	}

	taskResults := runCLI(t, dbPath, "", "search", "--tasks")
	if taskResults.exitCode != 0 {
		t.Fatalf("search tasks failed: stdout=%s stderr=%s", taskResults.stdout, taskResults.stderr)
	}
	if !strings.Contains(taskResults.stdout, "No notes found.") {
		t.Fatalf("expected empty task search output, got %q", taskResults.stdout)
	}
}

func TestCLI_InvalidInputErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")

	createNote := runCLI(t, dbPath, "", "create", "-c", "Regular note")
	if createNote.exitCode != 0 {
		t.Fatalf("create note failed: stdout=%s stderr=%s", createNote.stdout, createNote.stderr)
	}

	conflictingSearch := runCLI(t, dbPath, "", "search", "--completed", "--pending")
	if conflictingSearch.exitCode == 0 {
		t.Fatalf("expected conflicting search flags to fail, got stdout=%q stderr=%q", conflictingSearch.stdout, conflictingSearch.stderr)
	}
	if !strings.Contains(conflictingSearch.stderr, "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got stdout=%q stderr=%q", conflictingSearch.stdout, conflictingSearch.stderr)
	}

	checkNonTask := runCLI(t, dbPath, "", "check", "1")
	if checkNonTask.exitCode == 0 {
		t.Fatalf("expected check on non-task to fail, got stdout=%q stderr=%q", checkNonTask.stdout, checkNonTask.stderr)
	}
	if !strings.Contains(checkNonTask.stderr, "not a task") {
		t.Fatalf("expected task-only error, got stdout=%q stderr=%q", checkNonTask.stdout, checkNonTask.stderr)
	}
}

func TestCLI_RecurringTaskLifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")

	createDaily := runCLI(t, dbPath, "", "create", "-c", "Daily standup", "--task", "--due", "2026-03-10", "--recur", "daily", "-t", "work,team")
	if createDaily.exitCode != 0 {
		t.Fatalf("create daily recurring task failed: stdout=%s stderr=%s", createDaily.stdout, createDaily.stderr)
	}
	if !strings.Contains(createDaily.stdout, "Created task with ID 1") {
		t.Fatalf("expected created task output, got stdout=%q stderr=%q", createDaily.stdout, createDaily.stderr)
	}

	createWeekly := runCLI(t, dbPath, "", "create", "-c", "Weekly planning", "--task", "--due", "2026-03-16", "--recur", "weekly", "--weekday", "mon", "-t", "ops")
	if createWeekly.exitCode != 0 {
		t.Fatalf("create weekly recurring task failed: stdout=%s stderr=%s", createWeekly.stdout, createWeekly.stderr)
	}

	createMonthly := runCLI(t, dbPath, "", "create", "-c", "Monthly close", "--task", "--due", "2026-03-31", "--recur", "monthly", "--day-of-month", "31")
	if createMonthly.exitCode != 0 {
		t.Fatalf("create monthly recurring task failed: stdout=%s stderr=%s", createMonthly.stdout, createMonthly.stderr)
	}

	listTasks := runCLI(t, dbPath, "", "list", "tasks")
	if listTasks.exitCode != 0 {
		t.Fatalf("list tasks failed: stdout=%s stderr=%s", listTasks.stdout, listTasks.stderr)
	}
	if !strings.Contains(listTasks.stdout, "2026-03-10") || !strings.Contains(listTasks.stdout, "daily") {
		t.Fatalf("expected daily recurrence metadata in list output, got %q", listTasks.stdout)
	}
	if !strings.Contains(listTasks.stdout, "weekly(mon)") {
		t.Fatalf("expected weekly recurrence metadata in list output, got %q", listTasks.stdout)
	}
	if !strings.Contains(listTasks.stdout, "monthly(31)") {
		t.Fatalf("expected monthly recurrence metadata in list output, got %q", listTasks.stdout)
	}

	updateSeries := runCLI(t, dbPath, "", "update", "2", "-c", "Weekly planning and staffing", "--recur", "monthly", "--day-of-month", "31", "--due", "2026-03-31", "--add-tag", "urgent")
	if updateSeries.exitCode != 0 {
		t.Fatalf("update recurring task failed: stdout=%s stderr=%s", updateSeries.stdout, updateSeries.stderr)
	}

	checkDaily := runCLI(t, dbPath, "", "check", "1")
	if checkDaily.exitCode != 0 {
		t.Fatalf("check recurring daily task failed: stdout=%s stderr=%s", checkDaily.stdout, checkDaily.stderr)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	assertRecurringTaskState(t, db, recurringTaskExpectation{
		ID:           4,
		Content:      "Daily standup",
		DueDate:      "2026-03-11",
		Rule:         "daily",
		TaskSeriesID: 1,
		ExpectedTags: []string{"team", "work"},
		ShouldBeOpen: true,
	})
	assertRecurringTaskState(t, db, recurringTaskExpectation{
		ID:           2,
		Content:      "Weekly planning and staffing",
		DueDate:      "2026-03-31",
		Rule:         "monthly",
		DayOfMonth:   int64PtrCLI(31),
		TaskSeriesID: 2,
		ExpectedTags: []string{"ops", "urgent"},
		ShouldBeOpen: true,
	})

	checkOriginalAgain := runCLI(t, dbPath, "", "check", "1")
	if checkOriginalAgain.exitCode == 0 {
		t.Fatalf("expected second completion attempt to fail, got stdout=%q stderr=%q", checkOriginalAgain.stdout, checkOriginalAgain.stderr)
	}

	removeSeries := runCLI(t, dbPath, "y\n", "remove", "2")
	if removeSeries.exitCode != 0 {
		t.Fatalf("remove recurring series failed: stdout=%s stderr=%s", removeSeries.stdout, removeSeries.stderr)
	}

	searchCompleted := runCLI(t, dbPath, "", "search", "--tasks", "--completed")
	if searchCompleted.exitCode != 0 {
		t.Fatalf("search completed tasks failed: stdout=%s stderr=%s", searchCompleted.stdout, searchCompleted.stderr)
	}
	if !strings.Contains(searchCompleted.stdout, "Daily standup") {
		t.Fatalf("expected completed recurring history in search output, got %q", searchCompleted.stdout)
	}

	searchPending := runCLI(t, dbPath, "", "search", "--tasks", "--pending")
	if searchPending.exitCode != 0 {
		t.Fatalf("search pending tasks failed: stdout=%s stderr=%s", searchPending.stdout, searchPending.stderr)
	}
	if strings.Contains(searchPending.stdout, "Weekly planning and staffing") {
		t.Fatalf("expected removed recurring series to have no pending occurrence, got %q", searchPending.stdout)
	}
}

func TestCLI_RecurringTaskValidationErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "workbuddy.db")

	tests := []struct {
		name   string
		args   []string
		expect string
	}{
		{
			name:   "recurrence requires task",
			args:   []string{"create", "-c", "Just a note", "--recur", "daily", "--due", "2026-03-10"},
			expect: "recurrence is only available for tasks",
		},
		{
			name:   "weekly requires weekday",
			args:   []string{"create", "-c", "Weekly planning", "--task", "--recur", "weekly", "--due", "2026-03-16"},
			expect: "weekly recurrence requires --weekday",
		},
		{
			name:   "monthly requires day of month",
			args:   []string{"create", "-c", "Monthly close", "--task", "--recur", "monthly", "--due", "2026-03-31"},
			expect: "monthly recurrence requires --day-of-month",
		},
		{
			name:   "daily rejects extra scheduling flags",
			args:   []string{"create", "-c", "Daily standup", "--task", "--recur", "daily", "--due", "2026-03-10", "--weekday", "mon"},
			expect: "daily recurrence does not accept --weekday or --day-of-month",
		},
		{
			name:   "weekly due date must match weekday",
			args:   []string{"create", "-c", "Weekly planning", "--task", "--recur", "weekly", "--due", "2026-03-17", "--weekday", "mon"},
			expect: "due date must match the selected recurrence weekday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCLI(t, dbPath, "", tt.args...)
			if result.exitCode == 0 {
				t.Fatalf("expected CLI command to fail, got stdout=%q stderr=%q", result.stdout, result.stderr)
			}
			if !strings.Contains(result.stderr, tt.expect) {
				t.Fatalf("expected error %q, got stdout=%q stderr=%q", tt.expect, result.stdout, result.stderr)
			}
		})
	}
}

type recurringTaskExpectation struct {
	ID           int64
	Content      string
	DueDate      string
	Rule         string
	Weekday      *int64
	DayOfMonth   *int64
	TaskSeriesID int64
	ExpectedTags []string
	ShouldBeOpen bool
}

func assertRecurringTaskState(t *testing.T, db *sql.DB, want recurringTaskExpectation) {
	t.Helper()

	row := db.QueryRowContext(t.Context(), `
SELECT content, due_at, recurrence_rule, recurrence_weekday, recurrence_day_of_month, task_series_id, completed_at
FROM note
WHERE id = ?`, want.ID)

	var (
		content      string
		dueAt        sql.NullTime
		rule         sql.NullString
		weekday      sql.NullInt64
		dayOfMonth   sql.NullInt64
		taskSeriesID sql.NullInt64
		completedAt  sql.NullTime
	)
	if err := row.Scan(&content, &dueAt, &rule, &weekday, &dayOfMonth, &taskSeriesID, &completedAt); err != nil {
		t.Fatalf("failed to load note %d: %v", want.ID, err)
	}

	if content != want.Content {
		t.Fatalf("note content = %q, want %q", content, want.Content)
	}
	if !dueAt.Valid {
		t.Fatal("expected note due date to be set")
	}
	if got := dueAt.Time.Format("2006-01-02"); got != want.DueDate {
		t.Fatalf("note due date = %q, want %q", got, want.DueDate)
	}
	if rule.String != want.Rule {
		t.Fatalf("note recurrence rule = %q, want %q", rule.String, want.Rule)
	}
	if got, expected := nullInt64PtrCLI(weekday), want.Weekday; !reflect.DeepEqual(got, expected) {
		t.Fatalf("note recurrence weekday = %v, want %v", got, expected)
	}
	if got, expected := nullInt64PtrCLI(dayOfMonth), want.DayOfMonth; !reflect.DeepEqual(got, expected) {
		t.Fatalf("note recurrence day_of_month = %v, want %v", got, expected)
	}
	if !taskSeriesID.Valid || taskSeriesID.Int64 != want.TaskSeriesID {
		t.Fatalf("task_series_id = %v, want %d", taskSeriesID, want.TaskSeriesID)
	}
	if got := !completedAt.Valid; got != want.ShouldBeOpen {
		t.Fatalf("note open state = %v, want %v", got, want.ShouldBeOpen)
	}

	tagRows, err := db.QueryContext(t.Context(), `
SELECT t.name
FROM note_tags nt
INNER JOIN tags t ON t.id = nt.tag_id
WHERE nt.note_id = ?
ORDER BY t.name`, want.ID)
	if err != nil {
		t.Fatalf("failed to load note tags: %v", err)
	}
	defer tagRows.Close()

	var gotTags []string
	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			t.Fatalf("failed to scan note tag: %v", err)
		}
		gotTags = append(gotTags, tag)
	}
	if err := tagRows.Err(); err != nil {
		t.Fatalf("failed to iterate note tags: %v", err)
	}

	if fmt.Sprint(gotTags) != fmt.Sprint(want.ExpectedTags) {
		t.Fatalf("note tags = %v, want %v", gotTags, want.ExpectedTags)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to inspect sqlite_master for table %q: %v", name, err)
	}
	if count != 1 {
		t.Fatalf("expected table %q to exist", name)
	}
}

func assertIndexExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?", name)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to inspect sqlite_master for index %q: %v", name, err)
	}
	if count != 1 {
		t.Fatalf("expected index %q to exist", name)
	}
}

func int64PtrCLI(v int64) *int64 {
	return &v
}

func nullInt64PtrCLI(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	value := v.Int64
	return &value
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func runCLI(t *testing.T, dbPath, stdin string, args ...string) cliResult {
	t.Helper()

	repoRoot := projectRoot(t)
	cmdArgs := append([]string{"-test.run=TestCLIHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"WORKBUDDY_CLI_HELPER=1",
		"WORKBUDDY_DB="+dbPath,
		"WORKBUDDY_PROJECT_ROOT="+repoRoot,
		"NO_COLOR=1",
	)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return cliResult{
			stdout:   stdout.String(),
			stderr:   stderr.String(),
			exitCode: 0,
		}
	}

	var exitErr *exec.ExitError
	if !errorsAs(err, &exitErr) {
		t.Fatalf("failed to run helper CLI process: %v", err)
	}

	return cliResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitErr.ExitCode(),
	}
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv("WORKBUDDY_CLI_HELPER") != "1" {
		return
	}

	repoRoot := os.Getenv("WORKBUDDY_PROJECT_ROOT")
	if repoRoot == "" {
		fmt.Fprintln(os.Stderr, "missing WORKBUDDY_PROJECT_ROOT")
		os.Exit(2)
	}

	SetMigrations(os.DirFS(repoRoot))

	args := helperProcessArgs(os.Args)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func helperProcessArgs(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	return nil
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	return filepath.Dir(wd)
}

func errorsAs(err error, target interface{}) bool {
	switch v := target.(type) {
	case **exec.ExitError:
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return false
		}
		*v = exitErr
		return true
	default:
		return false
	}
}
