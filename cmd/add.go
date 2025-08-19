package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prime-run/togo/model"
	"github.com/spf13/cobra"
)

var (
	deadline     string
	hardDeadline bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new todo",
	Long:  `Add a new todo to your list. The todo will be marked as pending by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: Todo title is required")
			fmt.Println("Usage: togo add <title> [--deadline <deadline>] [--hard-deadline]")
			os.Exit(1)
		}
		title := strings.Join(args, " ")

		todoList := loadTodoListOrExit()
		
		// Parse deadline if provided
		var parsedDeadline *time.Time
		if deadline != "" {
			var err error
			parsedDeadline, err = model.ParseDeadline(deadline)
			if err != nil {
				fmt.Printf("Error parsing deadline: %v\n", err)
				os.Exit(1)
			}
		}

		// Add todo with deadline
		var todo *model.Todo
		if parsedDeadline != nil {
			todo = todoList.AddWithDeadline(title, parsedDeadline, hardDeadline)
		} else {
			todo = todoList.Add(title)
		}
		
		saveTodoListOrExit(todoList)

		fmt.Printf("Todo added successfully with ID: %d\n", todo.ID)
		fmt.Printf("Title: %s\n", todo.Title)
		if todo.Deadline != nil {
			deadlineStr := model.FormatDeadline(todo.Deadline, todo.HardDeadline)
			fmt.Printf("Deadline: %s\n", deadlineStr)
		}
	},
}

func init() {
	addCmd.Flags().StringVarP(&deadline, "deadline", "d", "", "Set deadline (e.g., '2h', '1d', '2024-01-15', '2024-01-15 15:30')")
	addCmd.Flags().BoolVarP(&hardDeadline, "hard-deadline", "", false, "Mark as hard deadline (shown with ! prefix)")
	rootCmd.AddCommand(addCmd)
}
