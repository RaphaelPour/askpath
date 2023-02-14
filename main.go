package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

const INIT_DIR = ""

type model struct {
	currentDir string
	entries    []os.DirEntry
	cursor     int

	terminal *termenv.Output
}

func newModel(dir string) model {
	return model{
		currentDir: dir,
		terminal: termenv.NewOutput(os.Stdout),
	}.UpdateEntries()
}

func (m model) Init() tea.Cmd {
	m.terminal.AltScreen()
	m.terminal.ClearScreen()
	return nil
}

func (m model) UpdateEntries() model{
	var err error
	m.entries, err = os.ReadDir(m.currentDir)
	if err != nil {
		fmt.Printf("error reading dir %s: %s\n", m.currentDir, err)
		os.Exit(1)
	}
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.terminal.ExitAltScreen()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.entries) - 1
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "right", "l":
			// skip empty dirs and files
			if len(m.entries) == 0 || !m.entries[m.cursor].IsDir() {
				return m, nil
			}

			m.currentDir = filepath.Join(m.currentDir, m.entries[m.cursor].Name())
			m.cursor = 0
			return m.UpdateEntries(), nil
		case "left", "h":
			m.cursor = 0
			previousDir := m.currentDir
			m.currentDir = filepath.Dir(m.currentDir)
			m = m.UpdateEntries()

			for i, entry := range m.entries {
				if strings.HasSuffix(previousDir, fmt.Sprintf("/%s",entry.Name())) {
					m.cursor = i
					break
				}
			}
			return m, nil
		case "enter":
			m.terminal.ExitAltScreen()
			if len(m.entries) == 0 {
				fmt.Println(m.currentDir)
			} else {
				fmt.Println(filepath.Join(m.currentDir, m.entries[m.cursor].Name()))
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	selection := ""
	if len(m.entries) > 0 {
		selection = m.entries[m.cursor].Name()
	}
	s := m.terminal.String(filepath.Join(m.currentDir, selection)).Reverse().String()
	m.terminal.Reset()
	s += m.terminal.String("\n").String()

	for i, entry := range m.entries {
		item := entry.Name()
		if entry.IsDir() {
			item += "/"
		}
		if m.cursor == i {
			item = m.terminal.String(item).Bold().String()
			m.terminal.Reset()
			item += m.terminal.String("").String()
		}

		s += fmt.Sprintf("%s\n", item)
	}

	return s
}

func main() {
	var dir string
	if len(os.Args) == 2 {
		dir = os.Args[1]
	} else {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fmt.Printf("error getting current directory: %s\n", err)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(newModel(dir))
	if err := p.Start(); err != nil {
		fmt.Printf("error running program: %s\n", err)
		os.Exit(1)
	}
}
