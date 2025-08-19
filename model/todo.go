package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Todo struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Completed    bool       `json:"completed"`
	Archived     bool       `json:"archived"`
	CreatedAt    time.Time  `json:"created_at"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	HardDeadline bool       `json:"hard_deadline"`
}

type TodoList struct {
	Todos    []Todo      `json:"todos"`
	NextID   int         `json:"next_id"`
	TodoByID map[int]int `json:"-"`
}

func NewTodoList() *TodoList {
	return &TodoList{
		Todos:    []Todo{},
		NextID:   1,
		TodoByID: make(map[int]int),
	}
}

func (tl *TodoList) rebuildIndex() {
	tl.TodoByID = make(map[int]int)
	for i, todo := range tl.Todos {
		tl.TodoByID[todo.ID] = i
	}
}

func (tl *TodoList) Add(title string) *Todo {
	todo := Todo{
		ID:           tl.NextID,
		Title:        title,
		Completed:    false,
		Archived:     false,
		CreatedAt:    time.Now(),
		Deadline:     nil,
		HardDeadline: false,
	}
	tl.Todos = append(tl.Todos, todo)
	tl.TodoByID[todo.ID] = len(tl.Todos) - 1
	tl.NextID++
	return &todo
}

// AddWithDeadline adds a new todo with an optional deadline
func (tl *TodoList) AddWithDeadline(title string, deadline *time.Time, hardDeadline bool) *Todo {
	todo := Todo{
		ID:           tl.NextID,
		Title:        title,
		Completed:    false,
		Archived:     false,
		CreatedAt:    time.Now(),
		Deadline:     deadline,
		HardDeadline: hardDeadline,
	}
	tl.Todos = append(tl.Todos, todo)
	tl.TodoByID[todo.ID] = len(tl.Todos) - 1
	tl.NextID++
	return &todo
}

// ParseDeadline parses deadline strings like "2h", "1d", "2024-01-15", "2024-01-15 15:30"
func ParseDeadline(deadlineStr string) (*time.Time, error) {
	if deadlineStr == "" {
		return nil, nil
	}

	// Try relative time format first (e.g., "2h", "1d", "30m")
	relativePattern := regexp.MustCompile(`^(\d+)([mhd])$`)
	if matches := relativePattern.FindStringSubmatch(deadlineStr); len(matches) == 3 {
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid time value: %s", matches[1])
		}

		var duration time.Duration
		switch matches[2] {
		case "m":
			duration = time.Duration(value) * time.Minute
		case "h":
			duration = time.Duration(value) * time.Hour
		case "d":
			duration = time.Duration(value) * 24 * time.Hour
		default:
			return nil, fmt.Errorf("invalid time unit: %s", matches[2])
		}

		deadline := time.Now().Add(duration)
		return &deadline, nil
	}

	// Try absolute date formats
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02",
		"01-02 15:04",
		"01-02",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, deadlineStr); err == nil {
			// For formats without year, use current year
			if strings.Count(format, "2006") == 0 {
				now := time.Now()
				parsed = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), 0, 0, time.Local)
			}
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("unable to parse deadline format: %s. Use formats like '2h', '1d', '2024-01-15' or '2024-01-15 15:30'", deadlineStr)
}

func (tl *TodoList) findIndexByID(id int) int {
	if idx, ok := tl.TodoByID[id]; ok {
		return idx
	}
	return -1
}

func (tl *TodoList) Toggle(id int) bool {
	idx := tl.findIndexByID(id)
	if idx == -1 {
		return false
	}
	tl.Todos[idx].Completed = !tl.Todos[idx].Completed
	return true
}

func (tl *TodoList) Archive(id int) bool {
	idx := tl.findIndexByID(id)
	if idx == -1 {
		return false
	}
	tl.Todos[idx].Archived = true
	return true
}

func (tl *TodoList) Unarchive(id int) bool {
	idx := tl.findIndexByID(id)
	if idx == -1 {
		return false
	}
	tl.Todos[idx].Archived = false
	return true
}

func (tl *TodoList) GetActiveTodos() []Todo {
	var activeTodos []Todo
	for _, todo := range tl.Todos {
		if !todo.Archived {
			activeTodos = append(activeTodos, todo)
		}
	}
	return activeTodos
}

func (tl *TodoList) GetArchivedTodos() []Todo {
	var archivedTodos []Todo
	for _, todo := range tl.Todos {
		if todo.Archived {
			archivedTodos = append(archivedTodos, todo)
		}
	}
	return archivedTodos
}

func (tl *TodoList) Delete(id int) bool {
	idx := tl.findIndexByID(id)
	if idx == -1 {
		return false
	}
	tl.Todos = append(tl.Todos[:idx], tl.Todos[idx+1:]...)
	tl.rebuildIndex()
	return true
}

func (tl *TodoList) Save(filename string) error {
	dataDir, err := getDataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(tl)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, filename), data, 0644)
}

func LoadTodoList(filename string) (*TodoList, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(dataDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return NewTodoList(), nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var tl TodoList
	if err := json.Unmarshal(data, &tl); err != nil {
		return nil, err
	}
	for i, todo := range tl.Todos {
		if todo.CreatedAt.IsZero() {
			tl.Todos[i].CreatedAt = time.Now()
		}
	}
	tl.TodoByID = make(map[int]int)
	for i, todo := range tl.Todos {
		tl.TodoByID[todo.ID] = i
	}
	return &tl, nil
}

func getDataDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	dataDir := filepath.Join(cacheDir, "togo")
	return dataDir, nil
}

func (tl *TodoList) GetTodoByID(id int) *Todo {
	idx := tl.findIndexByID(id)
	if idx == -1 {
		return nil
	}
	return &tl.Todos[idx]
}

func (tl *TodoList) FindByTitle(title string, caseSensitive bool) (*Todo, bool) {
	for i, todo := range tl.Todos {
		if caseSensitive {
			if todo.Title == title {
				return &tl.Todos[i], true
			}
		} else {
			if strings.EqualFold(todo.Title, title) {
				return &tl.Todos[i], true
			}
		}
	}
	return nil, false
}

func (tl *TodoList) DeleteByTitle(title string, caseSensitive bool) bool {
	for i, todo := range tl.Todos {
		var matches bool
		if caseSensitive {
			matches = todo.Title == title
		} else {
			matches = strings.EqualFold(todo.Title, title)
		}
		if matches {
			tl.Todos = append(tl.Todos[:i], tl.Todos[i+1:]...)
			tl.rebuildIndex()
			return true
		}
	}
	return false
}

func (tl *TodoList) GetTodoTitles() []string {
	titles := make([]string, len(tl.Todos))
	for i, todo := range tl.Todos {
		titles[i] = todo.Title
	}
	return titles
}

func (tl *TodoList) GetActiveAndArchivedTodoTitles() ([]string, []string) {
	var activeTitles, archivedTitles []string
	for _, todo := range tl.Todos {
		if todo.Archived {
			archivedTitles = append(archivedTitles, todo.Title)
		} else {
			activeTitles = append(activeTitles, todo.Title)
		}
	}
	return activeTitles, archivedTitles
}

func FormatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	} else {
		return "now"
	}
}

func FormatDeadline(deadline *time.Time, hardDeadline bool) string {
	if deadline == nil {
		return ""
	}
	
	now := time.Now()
	diff := deadline.Sub(now)
	
	prefix := ""
	if hardDeadline {
		prefix = "! "
	}
	
	if diff < 0 {
		// Overdue
		diff = -diff
		days := int(diff.Hours() / 24)
		hours := int(diff.Hours()) % 24
		if days > 0 {
			return fmt.Sprintf("%sOverdue %dd", prefix, days)
		} else if hours > 0 {
			return fmt.Sprintf("%sOverdue %dh", prefix, hours)
		} else {
			return fmt.Sprintf("%sOverdue", prefix)
		}
	}
	
	// Future deadline
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%s%dd", prefix, days)
	} else if hours > 0 {
		return fmt.Sprintf("%s%dh", prefix, hours)
	} else {
		return fmt.Sprintf("%sSoon", prefix)
	}
}
