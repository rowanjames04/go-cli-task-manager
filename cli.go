package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const taskFile = "tasks.json"
const legacyFile = "tasks.txt"

func main() {
	store := NewStore(taskFile)

	// Run migration from legacy tasks.txt if it exists
	if err := Migrate(taskFile, legacyFile); err != nil {
		fmt.Printf("Migration error: %v\n", err)
	}

	rootCmd := &cobra.Command{
		Use:   "taskmanager",
		Short: "A simple CLI task manager",
	}

	// Add command
	addCmd := &cobra.Command{
		Use:   "add [description]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskDescription := strings.Join(args, " ")
			task, err := store.Add(taskDescription)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task added: %s (ID: %d)\n", task.Description, task.ID)
		},
	}

	// Done command
	doneCmd := &cobra.Command{
		Use:   "done [id]",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			if err := store.ToggleCompleted(num); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Toggled completion for task %d\n", num)
		},
	}

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			if err := store.Delete(num); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Println("Deleted task", num)
		},
	}

	// Edit command
	editCmd := &cobra.Command{
		Use:   "edit [id] [description]",
		Short: "Edit a task description",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			newDescription := strings.Join(args[1:], " ")
			if err := store.UpdateDescription(num, newDescription); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task %d updated to: %s\n", num, newDescription)
		},
	}

	// List command
	var filterPending bool
	var filterCompleted bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := store.Read()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			var filteredTasks []Task
			for _, task := range tasks {
				if (!filterPending && !filterCompleted) || (filterPending && !task.Completed) || (filterCompleted && task.Completed) {
					filteredTasks = append(filteredTasks, task)
				}
			}

			if len(filteredTasks) == 0 {
				fmt.Println("No matching tasks found 🎉")
				return
			}

			for _, task := range filteredTasks {
				status := "[ ]"
				if task.Completed {
					status = "[x]"
				}
				fmt.Printf("%d. %s %s\n", task.ID, status, task.Description)
			}
		},
	}
	listCmd.Flags().BoolVarP(&filterPending, "pending", "p", false, "Show only pending tasks")
	listCmd.Flags().BoolVarP(&filterCompleted, "completed", "c", false, "Show only completed tasks")

	rootCmd.AddCommand(addCmd, doneCmd, deleteCmd, editCmd, listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
