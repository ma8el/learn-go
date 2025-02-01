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
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listStyle       = lipgloss.NewStyle().Margin(1, 1).Border(lipgloss.RoundedBorder())
	noteHeaderStyle = lipgloss.NewStyle().Padding(0, 1).Width(80).Height(1).Border(lipgloss.RoundedBorder())
	noteStyle       = lipgloss.NewStyle().Padding(0, 1).Width(80).Height(29).Border(lipgloss.RoundedBorder())
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
	list       list.Model
	textarea   textarea.Model
	titleInput textinput.Model
	cursor     int
	focus      string
	creating   bool
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
		case "ctrl+n":
			if !m.creating && m.focus == "list" {
				m.creating = true
				m.titleInput.Reset()
				return m, textinput.Blink
			}
		case "esc":
			if m.creating {
				m.creating = false
				return m, nil
			} else {
				m.focus = "list"
			}
		case "enter":
			if m.creating {
				title := m.titleInput.Value()
				if title != "" {
					createNote(title, "")
					m.list.SetItems(loadNotes())
				}
				m.creating = false
				return m, nil
			} else {
				m.focus = "content"
				m.textarea.Focus()
			}
		case "ctrl+b":
			m.focus = "list"
			updateNote(m.list.SelectedItem().(noteListItem).ID(), m.list.SelectedItem().(noteListItem).title, m.textarea.Value())
		case "up", "k":
			if m.cursor > 0 && m.focus == "list" {
				m.cursor--
				m.textarea.SetValue(m.list.Items()[m.cursor].(noteListItem).content)
			}
		case "down", "j":
			if m.cursor < len(m.list.Items())-1 && m.focus == "list" {
				m.cursor++
				m.textarea.SetValue(m.list.Items()[m.cursor].(noteListItem).content)
			}
		case "d":
			if m.focus == "list" {
				deleteNote(m.list.SelectedItem().(noteListItem).ID())
				m.list.RemoveItem(m.cursor)
			}
		}

		if m.creating {
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	if m.focus == "list" {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.textarea, cmd = m.textarea.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	var header string

	if len(m.list.Items()) > 0 && m.cursor < len(m.list.Items()) {
		item := m.list.Items()[m.cursor].(noteListItem)
		header = fmt.Sprintf("ID: %s\nTitle: %s", item.id, item.title)
	} else {
		header = "No notes available"
	}

	listBorder := listStyle
	headerBorder := noteHeaderStyle
	contentBorder := noteStyle

	if m.focus == "list" {
		listBorder = listBorder.BorderForeground(lipgloss.Color("213"))
	} else {
		contentBorder = contentBorder.BorderForeground(lipgloss.Color("213"))
		headerBorder = headerBorder.BorderForeground(lipgloss.Color("213"))
	}

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Center,
		listBorder.Render(m.list.View()),
		lipgloss.JoinVertical(
			lipgloss.Center,
			headerBorder.Render(header),
			contentBorder.Render(m.textarea.View()),
		),
	)

	if m.creating {
		return lipgloss.Place(
			m.list.Width(),
			m.list.Height(),
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("213")).
				Padding(1, 2).
				Render(
					lipgloss.JoinVertical(
						lipgloss.Center,
						m.titleInput.View(),
						lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Enter to save • Esc to cancel"),
					),
				),
		)
	}

	return mainView
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

func initialModel() model {
	items := loadNotes()
	ta := textarea.New()
	if len(items) > 0 {
		ta.Placeholder = items[0].(noteListItem).content
		if ta.Placeholder == "" {
			ta.Placeholder = "This note is empty"
		}
	} else {
		ta.Placeholder = "No notes available"
	}

	ta.SetWidth(80)
	ta.SetHeight(20)
	ta.ShowLineNumbers = true

	ti := textinput.New()
	ti.Placeholder = "Title"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	return model{
		list:       list.New(items, list.NewDefaultDelegate(), 100, 100),
		textarea:   ta,
		cursor:     0,
		focus:      "list",
		titleInput: ti,
	}
}

func main() {
	m := initialModel()
	m.list.Title = "Notes"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
