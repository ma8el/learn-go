package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type noteListItem struct {
	//	title     string
	//	createdAt time.Time
	title, desc string
}

func (i noteListItem) Title() string { return i.title }

// func (i noteListItem) CreatedAt() string   { return i.createdAt.Format(time.RFC3339) }
func (i noteListItem) Description() string { return i.desc }
func (i noteListItem) FilterValue() string { return i.title }

type apiNote struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
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
			title: note.Title,
			desc:  note.Content,
		}
	}
	return items
}

func main() {
	items := loadNotes()

	m := model{list: list.New(items, list.NewDefaultDelegate(), 100, 100)}
	m.list.Title = "Notes"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
