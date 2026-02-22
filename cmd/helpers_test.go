package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestOpenEditorWithRunner(t *testing.T) {
	tests := []struct {
		name        string
		runner      editorRunner
		wantContent string
		wantErr     bool
	}{
		{
			name: "returns content written by runner",
			runner: func(path string) error {
				return os.WriteFile(path, []byte("my test note"), 0644)
			},
			wantContent: "my test note",
		},
		{
			name: "trims surrounding whitespace",
			runner: func(path string) error {
				return os.WriteFile(path, []byte("  note with spaces  \n"), 0644)
			},
			wantContent: "note with spaces",
		},
		{
			name: "returns error when content is empty",
			runner: func(path string) error {
				return nil // writes nothing, file stays empty
			},
			wantErr: true,
		},
		{
			name: "returns error when runner fails",
			runner: func(path string) error {
				return fmt.Errorf("editor crashed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := openEditorWithRunner(tt.runner)

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
