package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"autocomplete/ollamastream"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func generate(prompt string) string {
	const ollamaEndpoint = "http://localhost:11434/api/generate"
	var fullText string

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := ollamastream.GenerateStream(
		ctx,
		prompt,
		ollamaEndpoint,
		"llama3.2",
		0.7,
		80,
		func(token string) {
			fullText += token
		},
	)
	if err != nil {
		log.Fatalf("GenerateStream error: %v", err)
	}

	return fullText
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput.Placeholder = generate(m.textInput.Value())
	fmt.Println(m.textInput.Placeholder)
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"Let AI autocomplete your text\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
