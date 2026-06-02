package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	var priority int
	var dueStr string
	addCmd := &cobra.Command{
		Use:   "add [description]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskDescription := strings.Join(args, " ")

			var dueDate *time.Time
			if dueStr != "" {
				t, err := time.Parse("2006-01-02", dueStr)
				if err != nil {
					fmt.Println("Invalid date format. Use YYYY-MM-DD")
					return
				}
				dueDate = &t
			}

			task, err := store.Add(taskDescription, priority, dueDate)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task added: %s (ID: %d, Priority: %d)\n", task.Description, task.ID, task.Priority)
		},
	}
	addCmd.Flags().IntVarP(&priority, "priority", "p", 2, "Priority level (1=Low, 2=Medium, 3=High)")
	addCmd.Flags().StringVarP(&dueStr, "due", "d", "", "Due date in YYYY-MM-DD format")

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

	// Priority command
	priorityCmd := &cobra.Command{
		Use:   "priority [id] [level]",
		Short: "Change a task's priority",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			p, err := strconv.Atoi(args[1])
			if err != nil || p < 1 || p > 3 {
				fmt.Println("Priority level must be 1 (Low), 2 (Medium), or 3 (High)")
				return
			}
			if err := store.SetPriority(num, p); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task %d priority set to %d\n", num, p)
		},
	}

	// Due command
	dueCmd := &cobra.Command{
		Use:   "due [id] [date]",
		Short: "Set a due date for a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			t, err := time.Parse("2006-01-02", args[1])
			if err != nil {
				fmt.Println("Invalid date format. Use YYYY-MM-DD")
				return
			}
			if err := store.SetDueDate(num, &t); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task %d due date set to %s\n", num, args[1])
		},
	}

	// List command
	var filterPending bool
	var filterCompleted bool
	var sortByPriority bool

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

			if sortByPriority {
				for i := 0; i < len(filteredTasks); i++ {
					for j := i + 1; j < len(filteredTasks); j++ {
						if filteredTasks[i].Priority < filteredTasks[j].Priority {
							filteredTasks[i], filteredTasks[j] = filteredTasks[j], filteredTasks[i]
						}
					}
				}
			}

			if len(filteredTasks) == 0 {
				fmt.Println("No matching tasks found 🎉")
				return
			}

			now := time.Now().Truncate(24 * time.Hour)
			for _, task := range filteredTasks {
				status := "[ ]"
				if task.Completed {
					status = "[x]"
				}
				priorityLabel := "Med"
				switch task.Priority {
				case 1:
					priorityLabel = "Low"
				case 3:
					priorityLabel = "High"
				}

				dueLabel := ""
				if task.DueDate != nil {
					dueLabel = fmt.Sprintf(" (Due: %s)", task.DueDate.Format("2006-01-02"))
					if task.DueDate.Before(now) && !task.Completed {
						dueLabel = " ⚠️ " + dueLabel
					}
				}

				fmt.Printf("%d. %s [%s] %s%s\n", task.ID, status, priorityLabel, task.Description, dueLabel)
			}
		},
	}
	listCmd.Flags().BoolVarP(&filterPending, "pending", "p", false, "Show only pending tasks")
	listCmd.Flags().BoolVarP(&filterCompleted, "completed", "c", false, "Show only completed tasks")
	listCmd.Flags().BoolVarP(&sortByPriority, "priority", "s", false, "Sort by priority (High to Low)")

	rootCmd.AddCommand(addCmd, doneCmd, deleteCmd, editCmd, priorityCmd, dueCmd, listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
