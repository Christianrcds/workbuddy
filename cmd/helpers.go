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
	"github.com/charmbracelet/lipgloss/table"
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

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// BorderType define o tipo de borda da tabela
type BorderType int

const (
	BorderTypeNormal BorderType = iota
	BorderTypeASCII
	BorderTypeMarkdown
)

// getBorder retorna o lipgloss.Border baseado no tipo
// ✅ Boa prática: centralizar configuração de bordas em uma função
func getBorder(bt BorderType) lipgloss.Border {
	switch bt {
	case BorderTypeASCII:
		return lipgloss.ASCIIBorder()
	case BorderTypeMarkdown:
		return lipgloss.MarkdownBorder()
	default:
		return lipgloss.NormalBorder()
	}
}

// displayNotesAsTable renderiza notas em formato de tabela bonita com lipgloss/table
// borderType: especifica o tipo de borda (Normal, ASCII, Markdown)
func displayNotesAsTable(notes []note.Note, tagsByNoteID map[int64][]string, title string) {
	// Define cores
	purple := lipgloss.Color("99")
	gray := lipgloss.Color("245")
	lightGray := lipgloss.Color("241")

	// Define estilos para cabeçalho
	headerStyle := lipgloss.NewStyle().
		Foreground(purple).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 1).
		Width(14)

	// Estilos para célula comum
	cellStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(14)

	oddRowStyle := cellStyle.Foreground(gray)
	evenRowStyle := cellStyle.Foreground(lightGray)

	// Cria array 2D de strings (rows)
	// ⚠️ IMPORTANTE: [][]string, não []string!
	var rows [][]string

	showStatus := false
	if len(notes) > 0 && notes[0].IsTask {
		showStatus = true
	}

	for _, n := range notes {

		// Formata data
		dateStr := n.CreatedAt.In(time.Local).Format("02/01 15:04")

		// Determina status
		statusStr := "-"
		if n.IsTask {
			if n.CompletedAt.Valid {
				statusStr = "✅ Done"
			} else {
				statusStr = "⏳ Pending"
			}
		}

		descStr := n.Content

		// Formata tags
		tagsStr := "-"
		if tags, ok := tagsByNoteID[n.ID]; ok && len(tags) > 0 {
			tagsStr = strings.Join(tags, ",")
			tagsStr = truncateString(tagsStr, 15)
		}

		// ID
		idStr := fmt.Sprintf("%d", n.ID)

		// Adiciona row como []string
		if showStatus {
			rows = append(rows, []string{idStr, dateStr, descStr, tagsStr, statusStr})
		} else {
			rows = append(rows, []string{idStr, dateStr, descStr, tagsStr})
		}
	}

	// Cria tabela com StyleFunc
	// ⚠️ IMPORTANTE: Rows ANTES de Offset! Ordem importa!
	// ⚠️ IMPORTANTE: StyleFunc é uma função de callback que recebe (row, col)

	var headers []string
	if showStatus {
		headers = []string{"ID", "Date", "Description", "Tags", "Status"}
	} else {
		headers = []string{"ID", "Date", "Description", "Tags"}
	}

	t := table.New().
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			// HeaderRow é uma constante especial do package table
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers(headers...).
		Rows(rows...) // ✅ Rows VEM PRIMEIRO

	// Imprime título + tabela
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	fmt.Printf("\n%s\n", titleStyle.Render(title))
	fmt.Println(t.Render())
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
