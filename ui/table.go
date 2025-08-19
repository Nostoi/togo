package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prime-run/togo/model"
)

const (
	checkboxEmpty  = "[ ]"
	checkboxFilled = "[×]"
)

func NewTodoTable(todoList *model.TodoList) TodoTableModel {
	displayWidth := 80
	checkboxColWidth := 5
	statusColWidth := 15
	createdAtColWidth := 15
	deadlineColWidth := 12
	titleColWidth := displayWidth - checkboxColWidth - statusColWidth - createdAtColWidth - deadlineColWidth - 10
	
	var columns []table.Column
	if titleColWidth >= 20 {
		columns = []table.Column{
			{Title: "✓", Width: checkboxColWidth},
			{Title: "Title", Width: titleColWidth},
			{Title: "Status", Width: statusColWidth},
			{Title: "Deadline", Width: deadlineColWidth},
			{Title: "Created", Width: createdAtColWidth},
		}
	} else {
		titleColWidth = displayWidth - checkboxColWidth - statusColWidth - createdAtColWidth - 8
		columns = []table.Column{
			{Title: "✓", Width: checkboxColWidth},
			{Title: "Title", Width: titleColWidth},
			{Title: "Status", Width: statusColWidth},
			{Title: "Created", Width: createdAtColWidth},
		}
	}
	
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("252"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("236")).
		Bold(true)
	t.SetStyles(s)
	ti := textinput.New()
	ti.Placeholder = "Enter new task title"
	ti.Focus()
	ti.CharLimit = 120
	ti.Width = titleColWidth
	
	// Create deadline input
	di := textinput.New()
	di.Placeholder = "Enter deadline (e.g., 2h, 1d, 2026-01-15) or press Enter to skip"
	di.CharLimit = 50
	di.Width = titleColWidth
	
	showArchived := false
	for _, todo := range todoList.Todos {
		if todo.Archived {
			showArchived = true
			break
		}
	}
	m := TodoTableModel{
		todoList:         todoList,
		table:            t,
		mode:             ModeNormal,
		confirmAction:    "",
		actionTitle:      "",
		viewTaskID:       0,
		width:            displayWidth,
		height:           24,
		selectedTodoIDs:  make(map[int]bool),
		bulkActionActive: false,
		textInput:        ti,
		deadlineInput:    di,
		showArchived:     showArchived,
		showAll:          true,
		showArchivedOnly: false,
		statusMessage:    "",
		showHelp:         true,
		// Initialize new task fields
		newTaskTitle:        "",
		newTaskDeadline:     "",
		newTaskHardDeadline: false,
	}
	m = m.updateRows()
	return m
}

func (m *TodoTableModel) SetShowArchivedOnly(show bool) {
	m.showArchivedOnly = show
	m.showAll = false
	*m = m.updateRows()
}

func (m *TodoTableModel) SetShowAll(show bool) {
	m.showAll = show
	m.showArchivedOnly = false
	*m = m.updateRows()
}

func (m *TodoTableModel) SetShowActiveOnly(show bool) {
	m.showAll = false
	m.showArchivedOnly = false
	*m = m.updateRows()
}

func (m TodoTableModel) updateRows() TodoTableModel {
	availableWidth := m.width - 8
	if availableWidth < 40 {
		availableWidth = 40
	}

	checkboxColWidth := 5
	statusColWidth := 15
	createdAtColWidth := 15
	deadlineColWidth := 12
	titleColWidth := availableWidth - checkboxColWidth - statusColWidth - createdAtColWidth - deadlineColWidth - 8
	if titleColWidth < 20 {
		titleColWidth = 20
		deadlineColWidth = 0 // Hide deadline column if space is too tight
	}

	var columns []table.Column
	if deadlineColWidth > 0 {
		columns = []table.Column{
			{Title: "✓", Width: checkboxColWidth},
			{Title: "Title", Width: titleColWidth},
			{Title: "Status", Width: statusColWidth},
			{Title: "Deadline", Width: deadlineColWidth},
			{Title: "Created", Width: createdAtColWidth},
		}
	} else {
		columns = []table.Column{
			{Title: "✓", Width: checkboxColWidth},
			{Title: "Title", Width: titleColWidth},
			{Title: "Status", Width: statusColWidth},
			{Title: "Created", Width: createdAtColWidth},
		}
	}

	m.table.SetColumns(columns)

	var rows []table.Row
	var filteredTodos []model.Todo

	if m.showAll {
		filteredTodos = m.todoList.Todos
	} else if m.showArchivedOnly {
		filteredTodos = m.todoList.GetArchivedTodos()
	} else {
		filteredTodos = m.todoList.GetActiveTodos()
	}

	for _, todo := range filteredTodos {
		checkbox := checkboxEmpty
		if m.selectedTodoIDs[todo.ID] {
			checkbox = checkboxFilled
		}
		title := todo.Title
		if todo.Archived {
			title = archivedStyle.Render(title)
		}
		var status string
		if todo.Completed {
			status = statusCompleteStyle.Render("Completed")
		} else {
			status = statusPendingStyle.Render("Pending")
		}
		createdAt := model.FormatTimeAgo(todo.CreatedAt)
		
		if deadlineColWidth > 0 {
			deadline := model.FormatDeadline(todo.Deadline, todo.HardDeadline)
			rows = append(rows, table.Row{checkbox, title, status, deadline, createdAt})
		} else {
			rows = append(rows, table.Row{checkbox, title, status, createdAt})
		}
	}
	
	// Set the rows on the table
	m.table.SetRows(rows)
	
	// Update the columns to ensure the table refreshes completely
	m.table.SetColumns(columns)

	baseStyle.Width(availableWidth)

	extra := 4
	helpLines := 0
	if m.mode == ModeNormal {
		if m.showHelp {
			helpLines = 2
			if m.bulkActionActive {
				helpLines += 9
			} else {
				helpLines += 8
			}
		} else {
			helpLines = 1
		}
	}

	rowsHeight := m.height - extra - helpLines
	if rowsHeight < 3 {
		rowsHeight = 3
	}
	m.table.SetHeight(rowsHeight)
	return m
}

func (m TodoTableModel) findTodoByID(id int) *model.Todo {
	return m.todoList.GetTodoByID(id)
}

func (m TodoTableModel) findTodoByTitle(title string) *model.Todo {
	for i, todo := range m.todoList.Todos {
		if todo.Title == title {
			return &m.todoList.Todos[i]
		}
	}
	return nil
}

func (m TodoTableModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TodoTableModel) SetStatusMessage(message string) {
	m.statusMessage = message
}
