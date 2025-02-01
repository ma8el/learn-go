package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listStyle       = lipgloss.NewStyle().Margin(1, 2)
	noteHeaderStyle = lipgloss.NewStyle().Padding(0, 1).Width(80).Height(3).Border(lipgloss.RoundedBorder())
	noteStyle       = lipgloss.NewStyle().Padding(0, 1).Width(80).Height(10).Border(lipgloss.RoundedBorder())
)

type noteListItem struct {
	id, title, content string
	createdAt          time.Time
}

func (i noteListItem) ID() string    { return i.id }
func (i noteListItem) Title() string { return i.title }
func (i noteListItem) Description() string {
	return fmt.Sprintf("Created: %s", i.createdAt.Format("2006-01-02 15:04"))
}
func (i noteListItem) FilterValue() string { return i.title }

type apiNote struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type model struct {
	list     list.Model
	selected string
	cursor   int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			item := m.list.SelectedItem().(noteListItem)
			m.selected = item.content
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.list.Items())-1 {
				m.cursor++
			}
		case "d":
			deleteNote(m.list.SelectedItem().(noteListItem).ID())
			m.list.RemoveItem(m.cursor)
		}
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var header string
	var content string

	if len(m.list.Items()) > 0 && m.cursor < len(m.list.Items()) {
		item := m.list.Items()[m.cursor].(noteListItem)
		header = fmt.Sprintf("ID: %s\nTitle: %s", item.id, item.title)
		content = item.content
	} else {
		header = "No notes available"
		content = ""
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		listStyle.Render(m.list.View()),
		lipgloss.JoinVertical(
			lipgloss.Center,
			noteHeaderStyle.Render(header),
			noteStyle.Render(content),
		),
	)
}

func loadNotes() []list.Item {
	resp, err := http.Get("http://localhost:3000/notes")
	if err != nil {
		log.Printf("Error fetching notes: %v", err)
		return nil
	}
	defer resp.Body.Close()
	var apiNotes []apiNote
	if err := json.NewDecoder(resp.Body).Decode(&apiNotes); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil
	}

	items := make([]list.Item, len(apiNotes))
	for i, note := range apiNotes {
		items[i] = noteListItem{
			id:        note.ID,
			title:     note.Title,
			content:   note.Content,
			createdAt: note.CreatedAt,
		}
	}
	return items
}

func createNote(title, content string) {
	resp, err := http.Post("http://localhost:3000/notes", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"title": "%s", "content": "%s"}`, title, content))))
	if err != nil {
		log.Printf("Error creating note: %v", err)
	}
	defer resp.Body.Close()
}

func updateNote(id, title, content string) {
	req, _ := http.NewRequest(http.MethodPut, "http://localhost:3000/notes/"+id, bytes.NewBuffer([]byte(fmt.Sprintf(`{"title": "%s", "content": "%s"}`, title, content))))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error updating note: %v", err)
	}
	defer resp.Body.Close()
}

func deleteNote(id string) {
	req, _ := http.NewRequest(http.MethodDelete, "http://localhost:3000/notes/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error deleting note: %v", err)
	}
	defer resp.Body.Close()
}

func main() {
	items := loadNotes()
	m := model{list: list.New(items, list.NewDefaultDelegate(), 100, 100), cursor: 0}
	m.list.Title = "Notes"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
