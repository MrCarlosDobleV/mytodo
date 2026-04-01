package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeNormal mode = iota
	modeAdding
	modeEditing
	modeConfirmDelete
)

type model struct {
	tasks    []Task
	cursor   int
	mode     mode
	input    string
	err      error
	quitting bool

	width  int
	height int
}

func initialModel() model {
	tasks, err := loadTask()
	if err != nil {
		tasks = []Task{}
	}
	return model{
		tasks:  tasks,
		cursor: 0,
		mode:   modeNormal,
		err:    err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeNormal:
			return updateNormalMode(m, msg)
		case modeAdding:
			return updateAddingMode(m, msg)
		case modeEditing:
			return updateEditingMode(m, msg)
		case modeConfirmDelete:
			return updateConfirmDeleteMode(m, msg)
		}
	}
	return m, nil
}

func updateNormalMode(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "x", " ":
		m.tasks = toggleTask(m.tasks, m.cursor)
		m.err = saveTasks(m.tasks)

	case "d":
		if len(m.tasks) > 0 {
			m.mode = modeConfirmDelete
		}

	case "a":
		m.mode = modeAdding
		m.input = ""

	case "e":
		if len(m.tasks) > 0 {
			m.mode = modeEditing
			m.input = m.tasks[m.cursor].Text
		}
	}

	return m, nil
}

func updateAddingMode(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.input = ""

	case "enter":
		text := strings.TrimSpace(m.input)
		if text != "" {
			m.tasks = addTask(m.tasks, text)
			m.cursor = len(m.tasks) - 1
			m.err = saveTasks(m.tasks)
		}
		m.mode = modeNormal
		m.input = ""

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}

	return m, nil
}

func updateEditingMode(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.input = ""

	case "enter":
		text := strings.TrimSpace(m.input)
		if text != "" && len(m.tasks) > 0 {
			m.tasks = editTask(m.tasks, m.cursor, text)
			m.err = saveTasks(m.tasks)
		}
		m.mode = modeNormal
		m.input = ""

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}

	return m, nil
}

func updateConfirmDeleteMode(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		if len(m.tasks) > 0 {
			m.tasks = deleteTask(m.tasks, m.cursor)
			if m.cursor >= len(m.tasks) && m.cursor > 0 {
				m.cursor--
			}
			m.err = saveTasks(m.tasks)
		}
		m.mode = modeNormal
	case "n", "esc":
		m.mode = modeNormal
	}
	return m, nil
}

func (m model) renderMainView() string {
	var b strings.Builder

	b.WriteString("mytodo\n\n")

	if len(m.tasks) == 0 {
		b.WriteString("No tasks yet.\n")
	} else {
		for i, task := range m.tasks {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			check := " "
			if task.Done {
				check = "x"
			}

			b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, check, task.Text))
		}
	}

	b.WriteString("\n")

	switch m.mode {
	case modeAdding:
		b.WriteString("Add task: " + m.input + "█\n\n")
	case modeEditing:
		b.WriteString("Edit task: " + m.input + "█\n\n")
	}

	b.WriteString(m.renderStatusBar())

	if m.err != nil {
		b.WriteString("\nError: " + m.err.Error() + "\n")
	}

	return b.String()
}

func (m model) renderDeleteModal() string {
	taskText := ""
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		taskText = m.tasks[m.cursor].Text
	}

	return fmt.Sprintf(
		"┌──────────────────────────────┐\n"+
			"│ Delete this task?            │\n"+
			"│                              │\n"+
			"│ %s │\n"+
			"│                              │\n"+
			"│ [y] confirm    [n] cancel    │\n"+
			"└──────────────────────────────┘",
		truncate(taskText, 28),
	)
}

func (m model) renderStatusBar() string {
	modeLabel := "NORMAL"
	help := "j/k move • a add • e edit • d delete • x toggle • q quit"

	switch m.mode {
	case modeAdding:
		modeLabel = "ADDING"
		help = "enter save • esc cancel"
	case modeEditing:
		modeLabel = "EDITING"
		help = "enter save • esc cancel"
	case modeConfirmDelete:
		modeLabel = "CONFIRM DELETE"
		help = "y confirm • n cancel"
	}

	position := "0/0"
	if len(m.tasks) > 0 {
		position = fmt.Sprintf("%d/%d", m.cursor+1, len(m.tasks))
	}

	return fmt.Sprintf("%s | %s | %s", modeLabel, position, help)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s + strings.Repeat(" ", max-len(s))
	}
	return s[:max-3] + "..."
}

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	main := m.renderMainView()

	if m.mode == modeConfirmDelete {
		return main + "\n\n" + m.renderDeleteModal()
	}
	return main
}
