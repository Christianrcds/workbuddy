package cmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
