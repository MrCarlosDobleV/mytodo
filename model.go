package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var (
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			BorderForeground(lipgloss.Color("63"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

	inputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))
)

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

	b.WriteString(titleStyle.Render("mytodo"))
	b.WriteString("\n\n")

	if len(m.tasks) == 0 {
		b.WriteString(dimStyle.Render("No tasks yet."))
		b.WriteString("\n")
	} else {
		for i, task := range m.tasks {
			check := " "
			if task.Done {
				check = "x"
			}

			prefix := "  "
			if i == m.cursor {
				prefix = "> "
			}

			line := fmt.Sprintf("%s[%s] %s", prefix, check, task.Text)

			if task.Done {
				line = dimStyle.Render(line)
			}

			if i == m.cursor {
				line = selectedStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	switch m.mode {
	case modeAdding:
		b.WriteString(inputLabelStyle.Render("Add task: "))
		b.WriteString(inputStyle.Render(m.input + "█"))
		b.WriteString("\n\n")
	case modeEditing:
		b.WriteString(inputLabelStyle.Render("Edit task: "))
		b.WriteString(inputStyle.Render(m.input + "█"))
		b.WriteString("\n\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m model) renderDeleteModal() string {
	taskText := ""
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		taskText = m.tasks[m.cursor].Text
	}

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Delete this task?"),
		taskText,
		dimStyle.Render("[y] confirm    [n] cancel"),
	)

	box := modalStyle.Render(content)

	// center it
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
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
