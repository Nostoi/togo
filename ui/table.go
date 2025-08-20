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

// normalizeCells ensures that the row has exactly n cells, padding with empty strings
// or truncating as needed to prevent index out of range errors during table rendering.
func normalizeCells(cells []string, n int) []string {
	if len(cells) > n {
		return cells[:n]
	}
	if len(cells) < n {
		padded := make([]string, n)
		copy(padded, cells)
		for i := len(cells); i < n; i++ {
			padded[i] = ""
		}
		return padded
	}
	return cells
}

func (m TodoTableModel) updateRows() TodoTableModel {
	availableWidth := m.width - 8
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Guard against zero/negative widths - enforce minimum column widths
	checkboxColWidth := 5
	statusColWidth := 15
	createdAtColWidth := 15
	deadlineColWidth := 12
	
	// Calculate title column width with minimum constraint
	titleColWidth := availableWidth - checkboxColWidth - statusColWidth - createdAtColWidth - deadlineColWidth - 8
	if titleColWidth < 20 {
		titleColWidth = 20
		deadlineColWidth = 0 // Hide deadline column if space is too tight
	}
	
	// Ensure all column widths are positive
	if checkboxColWidth < 1 { checkboxColWidth = 1 }
	if statusColWidth < 1 { statusColWidth = 1 }
	if createdAtColWidth < 1 { createdAtColWidth = 1 }
	if titleColWidth < 1 { titleColWidth = 1 }

	// Build columns first to determine target layout
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

	// Get the number of columns for normalization
	numColumns := len(columns)

	// Build rows with proper cell count normalization BEFORE setting anything on the table
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
		
		// Build row with appropriate number of cells
		var rowCells []string
		if deadlineColWidth > 0 {
			deadline := model.FormatDeadline(todo.Deadline, todo.HardDeadline)
			rowCells = []string{checkbox, title, status, deadline, createdAt}
		} else {
			rowCells = []string{checkbox, title, status, createdAt}
		}
		
		// Normalize cells to match column count exactly
		normalizedCells := normalizeCells(rowCells, numColumns)
		rows = append(rows, table.Row(normalizedCells))
	}

	// CRITICAL FIX: We need to avoid SetColumns triggering UpdateViewport 
	// while the table has mismatched rows. The safest approach is to
	// construct a new table with the right columns and rows from the start.
	
	// Get the current table state we want to preserve
	currentCursor := m.table.Cursor()
	currentFocus := m.table.Focused()
	
	// Create a new table with the correct columns and rows from the beginning
	newTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(currentFocus),
	)
	
	// Apply the same styles as the original table
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
	newTable.SetStyles(s)
	
	// Restore cursor position safely
	if currentCursor < len(rows) {
		newTable.SetCursor(currentCursor)
	}
	
	// Replace the table entirely to avoid any inconsistent intermediate states
	m.table = newTable

	// Set the height after everything is consistent
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
