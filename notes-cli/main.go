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

var (
	listStyle = lipgloss.NewStyle().Margin(1, 2)
	noteStyle = lipgloss.NewStyle().Padding(0, 1).Width(80).Height(10).Border(lipgloss.RoundedBorder())
)

type noteListItem struct {
	title, desc string
}

func (i noteListItem) Title() string { return i.title }

func (i noteListItem) Description() string { return i.desc }
func (i noteListItem) FilterValue() string { return i.title }

type apiNote struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type model struct {
	list     list.Model
	selected string // Add this field to track selected content
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
			m.selected = item.desc
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
	// Update this line to show the selected content
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		listStyle.Render(m.list.View()),
		noteStyle.Render(m.selected),
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
