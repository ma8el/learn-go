package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"autocomplete/ollamastream"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type suggestionMsg struct {
	suggestion string
	err        error
}

func generate(prompt string) tea.Cmd {
	return func() tea.Msg {
		const ollamaEndpoint = "http://localhost:11434/api/generate"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		modelInstructions := `You are a helpful assistant that generates suggestions 
		for a text input. Complete the input with the most likely next few words.
		Return only the suggestion, no other text.`
		promptWithInstructions := fmt.Sprintf("%s\n\n%s", modelInstructions, prompt)

		var suggestionString string

		err := ollamastream.GenerateStream(
			ctx,
			promptWithInstructions,
			ollamaEndpoint,
			"llama3.2",
			0.7,
			80,
			func(token string) { suggestionString += token },
		)
		if err != nil {
			log.Fatalf("GenerateStream error: %v", err)
			return suggestionMsg{err: err}
		}

		return suggestionMsg{suggestion: suggestionString}
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	textInput  textinput.Model
	suggestion string
	err        error
	loading    bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Start typing..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 100
	return model{
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			if m.suggestion != "" {
				currentVal := m.textInput.Value()
				m.textInput.SetValue(currentVal + m.suggestion)
				m.suggestion = ""
			}

		default:
			oldVal := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)

			newVal := m.textInput.Value()
			if newVal != oldVal {
				return m, generate(newVal)
			}
			return m, cmd
		}

	case suggestionMsg:
		m.loading = false
		m.suggestion = msg.suggestion
		m.err = msg.err

		if msg.err != nil {
			log.Println("Error fetching suggestion:", msg.err)
		}
		return m, nil

	default:
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	ghostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#64748b"))

	var suggestionView string
	if m.suggestion != "" {
		suggestionView = ghostStyle.Render(m.suggestion)
	}

	s := fmt.Sprintf(
		"AI Autocomplete (type to generate suggestions)\n\n%s\n\n%s",
		m.textInput.View(),
		suggestionView,
	) + "\n"

	if m.loading {
		s += ghostStyle.Render("Generating suggestions...")
	}

	if m.err != nil {
		s += ghostStyle.Render(m.err.Error())
	}

	s += "Press Tab to accept suggestion, Ctrl-C or q to quit"

	return s
}
