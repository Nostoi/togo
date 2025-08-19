package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/prime-run/togo/model"
)

func (m TodoTableModel) View() string {
	if m.mode == ModeViewDetail {
		todo := m.findTodoByID(m.viewTaskID)
		if todo == nil {
			return fullScreenStyle.Width(m.width).Height(m.height).Render(
				fullTaskViewStyle.Render("Task not found."))
		}
		status := "Pending"
		if todo.Completed {
			status = statusCompleteStyle.Render("Completed")
		} else {
			status = statusPendingStyle.Render("Pending")
		}
		archivedStatus := ""
		if todo.Archived {
			archivedStatus = "\nArchived: " + archivedStyle.Render("Yes")
		}
		
		deadlineInfo := ""
		if todo.Deadline != nil {
			deadlineType := "Soft Deadline"
			if todo.HardDeadline {
				deadlineType = "Hard Deadline"
			}
			deadlineStr := todo.Deadline.Format("2006-01-02 15:04")
			deadlineFormatted := model.FormatDeadline(todo.Deadline, todo.HardDeadline)
			deadlineInfo = fmt.Sprintf("\n%s: %s (%s)", deadlineType, deadlineStr, deadlineFormatted)
		}
		
		createdAt := model.FormatTimeAgo(todo.CreatedAt)
		taskView := fullTaskViewStyle.Render(
			taskTitleStyle.Render(todo.Title) + "\n\n" +
				"Status: " + status + archivedStatus + deadlineInfo + "\n" +
				"Created: " + createdAtStyle.Render(createdAt) + "\n\n" +
				helpStyle.Render("Press Enter to go back"))
		return fullScreenStyle.Width(m.width).Height(m.height).Render(taskView)
	}
	if m.mode == ModeDeleteConfirm || m.mode == ModeArchiveConfirm {
		var confirmMessage string
		action := "delete"
		if m.mode == ModeArchiveConfirm {
			action = "archive"
		}
		if len(m.selectedTodoIDs) > 0 && m.bulkActionActive {
			confirmMessage = fmt.Sprintf("Are you sure you want to %s %d selected tasks?", action, len(m.selectedTodoIDs))
		} else {
			confirmMessage = fmt.Sprintf("Are you sure you want to %s task: \"%s\"?", action, m.actionTitle)
		}
		confirmBox := confirmStyle.Render(
			confirmTextStyle.Render(confirmMessage) + "\n\n" +
				confirmBtnStyle.Render("Y - Yes") + " " + cancelBtnStyle.Render("N - No"))
		return fullScreenStyle.Width(m.width).Height(m.height).Render(confirmBox)
	}
	if m.mode == ModeAddTask {
		inputView := inputStyle.Render(
			inputPromptStyle.Render("Add New Task") + "\n\n" +
				m.textInput.View() + "\n\n" +
				helpStyle.Render("Press Enter to continue, Esc to cancel"))
		return fullScreenStyle.Width(m.width).Height(m.height).Render(inputView)
	}
	if m.mode == ModeAddTaskDeadline {
		inputView := inputStyle.Render(
			inputPromptStyle.Render("Set Deadline (Optional)") + "\n\n" +
				m.deadlineInput.View() + "\n\n" +
				helpStyle.Render("Examples: 2h, 1d, 2026-01-15, 2026-01-15 15:30\nPress Enter to continue (or skip), Esc to cancel"))
		return fullScreenStyle.Width(m.width).Height(m.height).Render(inputView)
	}
	if m.mode == ModeAddTaskDeadlineType {
		inputView := inputStyle.Render(
			inputPromptStyle.Render("Deadline Type") + "\n\n" +
				fmt.Sprintf("Task: %s\nDeadline: %s\n\n", m.newTaskTitle, m.newTaskDeadline) +
				helpStyle.Render("H - Hard deadline (important!)\nS/Enter - Soft deadline\nEsc - Cancel"))
		return fullScreenStyle.Width(m.width).Height(m.height).Render(inputView)
	}
	if len(m.todoList.Todos) == 0 {
		return baseStyle.Render("No tasks found. Press 'a' to add a new task!")
	}

	var helpText string
	var listTitle string

	if m.showArchivedOnly {
		listTitle = "Archived Tasks"
	} else if m.showAll {
		listTitle = "All Tasks"
	} else {
		listTitle = "Active Tasks"
	}

	leftSide := titleBarStyle.Render(listTitle)
	rightSide := successMessageStyle.Render(m.statusMessage)

	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Center,
		leftSide,
		lipgloss.PlaceHorizontal(
			m.width-lipgloss.Width(leftSide)-4,
			lipgloss.Right,
			rightSide,
		),
	)

	if m.bulkActionActive {
		helpText = "\n" + statusBar + "\n" +
			"Bulk Mode:" +
			"\n→ t: toggle completion for all selected" +
			"\n→ n: toggle archive/unarchive for selected" +
			"\n→ d: delete selected" +
			"\n→ space: toggle selection" +
			"\n→ enter: view details" +
			"\n→ a: add new task" +
			"\n→ q: quit" +
			"\n→ .: toggle help"
	} else {
		helpText = "\n" + statusBar + "\n" +
			"→ t: toggle completion" +
			"\n→ n: toggle archive/unarchive" +
			"\n→ d: delete" +
			"\n→ space: select" +
			"\n→ enter: view details" +
			"\n→ a: add new task" +
			"\n→ q: quit" +
			"\n→ .: toggle help"
	}

	tableView := tableContainerStyle.Render(m.table.View())
	if m.mode == ModeNormal {
		if m.showHelp {
			help := helpStyle.Render(helpText)
			return tableView + help
		}
		hint := helpStyle.Render("\n→ .: toggle help")
		return tableView + hint
	}
	return tableView
}
