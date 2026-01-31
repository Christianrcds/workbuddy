package cmd

import (
	"database/sql"
	"fmt"
	"os"
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

// openDatabase opens the SQLite database
// Uses WORKBUDDY_DB environment variable, falls back to default location
func openDatabase() (*sql.DB, error) {
	dbPath := os.Getenv("WORKBUDDY_DB")
	if dbPath == "" {
		// Default: internal/database/workbuddy.db in project root
		dbPath = "internal/database/workbuddy.db"
	}
	return sql.Open("sqlite", dbPath)
}

// NoteStyles holds all the lipgloss styles for displaying notes
type NoteStyles struct {
	Header  lipgloss.Style
	ID      lipgloss.Style
	Content lipgloss.Style
	Date    lipgloss.Style
	Box     lipgloss.Style
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
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(1),
	}
}

// displayNotes displays a slice of notes with pretty formatting
func displayNotes(notes []note.Note, title string) {
	styles := getNoteStyles()

	fmt.Printf("\n%s\n", styles.Header.Render(title))

	for _, n := range notes {
		dateStr := n.CreatedAt.In(time.Local).Format("Mon, 02 Jan 2006 15:04")

		noteContent := fmt.Sprintf(
			"%s\n%s\n%s",
			styles.Date.Render(dateStr),
			styles.Content.Render(wrapText(n.Content, 70)),
			styles.ID.Render(fmt.Sprintf("ID: %d", n.ID)),
		)

		fmt.Println(styles.Box.Render(noteContent))
	}
}
