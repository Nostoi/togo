package main

import (
	"testing"
	"github.com/prime-run/togo/model"
	"github.com/prime-run/togo/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// TestUIBoundsChecking tests that the UI properly handles edge cases
// that could cause array bounds panics.
func TestUIBoundsChecking(t *testing.T) {
	// Test with empty todo list
	todoList := model.NewTodoList()
	tableModel := ui.NewTodoTable(todoList)
	
	// Test enter key with empty table - should not panic
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := tableModel.Update(msg)
	if cmd != nil {
		t.Logf("Empty table enter handled correctly")
	}
	
	// Test toggle (t key) with empty table - should not panic
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	_, cmd = tableModel.Update(msg)
	if cmd != nil {
		t.Logf("Empty table toggle handled correctly")
	}
	
	// Test archive (n key) with empty table - should not panic
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd = tableModel.Update(msg)
	if cmd != nil {
		t.Logf("Empty table archive handled correctly")
	}
	
	// Test delete (d key) with empty table - should not panic
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, cmd = tableModel.Update(msg)
	if cmd != nil {
		t.Logf("Empty table delete handled correctly")
	}
}

// TestUIWithTasks tests that the UI properly handles scenarios with tasks
func TestUIWithTasks(t *testing.T) {
	// Create todo list with a task
	todoList := model.NewTodoList()
	todoList.Add("Test task")
	
	tableModel := ui.NewTodoTable(todoList)
	
	// Test various interactions that previously could cause crashes
	testKeys := []rune{'t', 'n', 'd'}
	
	for _, key := range testKeys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}}
		_, cmd := tableModel.Update(msg)
		if cmd != nil {
			t.Logf("Key '%c' handled correctly with tasks", key)
		}
	}
	
	// Test enter key specifically
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := tableModel.Update(msg)
	if cmd != nil {
		t.Logf("Enter key handled correctly with tasks")
	}
}

// TestTableResizeScenarios tests the table behavior during various resize events
// to prevent the index out of range crash that occurs during orientation changes.
func TestTableResizeScenarios(t *testing.T) {
	// Create todo list with multiple tasks to test different scenarios
	todoList := model.NewTodoList()
	todoList.Add("Task 1")
	todoList.Add("Task 2") 
	todoList.Add("Very long task title that should be handled properly")
	
	tableModel := ui.NewTodoTable(todoList)
	
	// Test sequence of window size changes simulating iPhone orientation changes
	resizeSequence := []struct {
		width  int
		height int
		name   string
	}{
		{120, 40, "Initial size"},
		{40, 80, "Rotate to narrow (portrait)"},
		{150, 30, "Rotate to wide (landscape)"},
		{30, 100, "Rotate to very narrow"},
		{200, 50, "Rotate to very wide"},
		{25, 60, "Extremely narrow"},
		{80, 24, "Back to normal"},
	}
	
	for _, scenario := range resizeSequence {
		t.Run(scenario.name, func(t *testing.T) {
			// Send WindowSizeMsg which previously caused the crash
			msg := tea.WindowSizeMsg{
				Width:  scenario.width,
				Height: scenario.height,
			}
			
			// This should not panic
			_, cmd := tableModel.Update(msg)
			if cmd != nil {
				t.Logf("Resize to %dx%d handled successfully", scenario.width, scenario.height)
			}
			
			// Test that basic operations still work after resize
			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			_, enterCmd := tableModel.Update(enterMsg)
			if enterCmd != nil {
				t.Logf("Enter key works after resize to %dx%d", scenario.width, scenario.height)
			}
		})
	}
}

// TestTableColumnRowConsistency tests that columns and rows always have consistent counts
func TestTableColumnRowConsistency(t *testing.T) {
	todoList := model.NewTodoList()
	todoList.Add("Test task 1")
	todoList.Add("Test task 2")
	
	tableModel := ui.NewTodoTable(todoList)
	
	// Test various widths that trigger different column layouts
	testWidths := []int{20, 30, 40, 50, 80, 100, 120, 150, 200}
	
	for _, width := range testWidths {
		t.Run("Width_"+string(rune(width+'0')), func(t *testing.T) {
			// Resize to trigger column layout changes
			msg := tea.WindowSizeMsg{Width: width, Height: 40}
			updatedModel, _ := tableModel.Update(msg)
			
			// Cast back to access internal state
			if tm, ok := updatedModel.(ui.TodoTableModel); ok {
				// This should not panic - accessing the model itself validates internal consistency
				_ = tm.View()
				t.Logf("Width %d handled without panic", width)
			}
		})
	}
}

// TestNormalizeCellsHelper tests the helper function that prevents index out of range
func TestNormalizeCellsHelper(t *testing.T) {
	// This test would require exposing the normalizeCells function or creating a wrapper
	// For now, we test indirectly through the resize scenarios above
	t.Logf("Cell normalization tested indirectly through resize scenarios")
}

// TestRapidResizeEvents tests rapid succession of resize events
func TestRapidResizeEvents(t *testing.T) {
	todoList := model.NewTodoList()
	todoList.Add("Test task")
	
	tableModel := ui.NewTodoTable(todoList)
	
	// Simulate rapid resize events like during orientation changes
	widths := []int{120, 40, 80, 30, 100, 50, 90, 35, 110}
	
	for i, width := range widths {
		msg := tea.WindowSizeMsg{Width: width, Height: 40}
		updatedModel, cmd := tableModel.Update(msg)
		if cmd != nil {
			t.Logf("Rapid resize event %d (width=%d) handled", i+1, width)
		}
		// Update tableModel for next iteration
		if tm, ok := updatedModel.(ui.TodoTableModel); ok {
			tableModel = tm
		}
	}
	
	// Verify the table is still functional after rapid resizing
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, enterCmd := tableModel.Update(enterMsg)
	if enterCmd != nil {
		t.Logf("Table remains functional after rapid resize events")
	}
}