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