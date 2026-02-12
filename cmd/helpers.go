package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"workbuddy/internal/note"

	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
)

// wrapText wraps text to a specified width
func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	lines := []string{}
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

func getDBPath() string {
	if p := os.Getenv("WORKBUDDY_DB"); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "workbuddy.db"
	}
	return filepath.Join(dir, "workbuddy", "workbuddy.db")
}

func openDatabase() (*sql.DB, error) {
	dbPath := getDBPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	return sql.Open("sqlite", dbPath)
}

// NoteStyles holds all the lipgloss styles for displaying notes
type NoteStyles struct {
	Header        lipgloss.Style
	ID            lipgloss.Style
	Content       lipgloss.Style
	Date          lipgloss.Style
	Status        lipgloss.Style
	StatusDone    lipgloss.Style
	StatusPending lipgloss.Style
	Tags          lipgloss.Style
	TagsDone      lipgloss.Style
	Box           lipgloss.Style
}

// getNoteStyles returns configured styles for displaying notes
func getNoteStyles() NoteStyles {
	return NoteStyles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			PaddingLeft(1),
		ID: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			PaddingLeft(2),
		Content: lipgloss.NewStyle().
			PaddingLeft(2),
		Date: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Status: lipgloss.NewStyle(),
		StatusDone: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
		StatusPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")),
		Tags: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")),
		TagsDone: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(1),
	}
}

// displayNotes displays a slice of notes with pretty formatting
func displayNotes(notes []note.Note, tagsByNoteID map[int64][]string, title string) {
	styles := getNoteStyles()

	fmt.Printf("\n%s\n", styles.Header.Render(title))

	for _, n := range notes {
		dateStr := n.CreatedAt.In(time.Local).Format("Mon, 02 Jan 2006 15:04")
		statusText := ""
		statusStyle := styles.StatusPending
		if n.IsTask {
			statusText = "[ ] Pending"
			if n.CompletedAt.Valid {
				completedAt := n.CompletedAt.Time.In(time.Local).Format("Mon, 02 Jan 2006 15:04")
				statusText = fmt.Sprintf("[x] Completed: %s", completedAt)
				statusStyle = styles.StatusDone
			}
		}
		noteLines := []string{styles.Date.Render(dateStr)}
		if n.IsTask {
			noteLines = append(noteLines, statusStyle.Render(statusText))
		}
		if tagsLine := renderTagsLine(tagsByNoteID[n.ID], n.CompletedAt.Valid, styles); tagsLine != "" {
			noteLines = append(noteLines, tagsLine)
		}
		noteLines = append(
			noteLines,
			renderDescriptionLine(n.Content, n.CompletedAt.Valid, styles),
			styles.ID.Render(fmt.Sprintf("ID: %d", n.ID)),
		)

		noteContent := strings.Join(noteLines, "\n")

		fmt.Println(styles.Box.Render(noteContent))
	}
}

func renderTagsLine(tags []string, completed bool, styles NoteStyles) string {
	label := styles.Tags
	if completed {
		label = styles.TagsDone
	}

	if len(tags) == 0 {
		return ""
	}

	return label.Render(fmt.Sprintf("Tags: %s", strings.Join(tags, ", ")))
}

func renderDescriptionLine(content string, completed bool, styles NoteStyles) string {
	label := styles.Tags
	if completed {
		label = styles.TagsDone
	}

	wrapped := wrapText(content, 70)
	wrapped = strings.ReplaceAll(wrapped, "\n", "\n  ")

	return label.Render("Description: ") + styles.Content.Render(wrapped)
}

func buildTagsByNoteID(ctx context.Context, repo note.Repository, notes []note.Note) (map[int64][]string, error) {
	tagsByNoteID := make(map[int64][]string, len(notes))
	for _, n := range notes {
		tags, err := repo.ListTagsByNoteID(ctx, n.ID)
		if err != nil {
			return nil, err
		}
		if len(tags) > 0 {
			tagsByNoteID[n.ID] = tags
		}
	}
	return tagsByNoteID, nil
}
