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
	var tagsStr string
	var parentID int
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

			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
			}

			var pID *int
			if cmd.Flags().Changed("parent") {
				p := parentID
				pID = &p
			}

			task, err := store.Add(taskDescription, priority, dueDate, tags, pID)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task added: %s (ID: %d, Priority: %d)\n", task.Description, task.ID, task.Priority)
		},
	}
	addCmd.Flags().IntVarP(&priority, "priority", "p", 2, "Priority level (1=Low, 2=Medium, 3=High)")
	addCmd.Flags().StringVarP(&dueStr, "due", "d", "", "Due date in YYYY-MM-DD format")
	addCmd.Flags().StringVarP(&tagsStr, "tags", "t", "", "Comma-separated list of tags")
	addCmd.Flags().IntVar(&parentID, "parent", 0, "ID of the parent task")

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

	// Tag commands
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage task tags",
	}
	tagAddCmd := &cobra.Command{
		Use:   "add [id] [tag]",
		Short: "Add a tag to a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			if err := store.AddTag(num, args[1]); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Tag '%s' added to task %d\n", args[1], num)
		},
	}
	tagRemCmd := &cobra.Command{
		Use:   "remove [id] [tag]",
		Short: "Remove a tag from a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task number must be an integer")
				return
			}
			if err := store.RemoveTag(num, args[1]); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Tag '%s' removed from task %d\n", args[1], num)
		},
	}
	tagCmd.AddCommand(tagAddCmd, tagRemCmd)

	// Move command
	moveCmd := &cobra.Command{
		Use:   "move [id] [new_parent_id]",
		Short: "Change the parent of a task",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Task ID must be an integer")
				return
			}
			pIDInt, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("Parent ID must be an integer")
				return
			}

			var pID *int
			if pIDInt != 0 {
				p := pIDInt
				pID = &p
			}

			if err := store.MoveTask(id, pID); err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("Task %d moved under parent %d\n", id, pIDInt)
		},
	}

	// List command
	var filterPending bool
	var filterCompleted bool
	var sortByPriority bool
	var filterTag string

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
				match := (!filterPending && !filterCompleted) || (filterPending && !task.Completed) || (filterCompleted && task.Completed)
				if match && filterTag != "" {
					hasTag := false
					for _, t := range task.Tags {
						if t == filterTag {
							hasTag = true
							break
						}
					}
					if !hasTag {
						match = false
					}
				}
				if match {
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

			// For hierarchical display, we need to process the tasks.
			// A simple approach: recursively print tasks.
			printed := make(map[int]bool)

			var printTask func(int, int)
			printTask = func(id int, indent int) {
				for _, t := range filteredTasks {
					if t.ID == id {
						status := "[ ]"
						if t.Completed {
							status = "[x]"
						}
						priorityLabel := "Med"
						switch t.Priority {
						case 1:
							priorityLabel = "Low"
						case 3:
							priorityLabel = "High"
						}
						dueLabel := ""
						if t.DueDate != nil {
							dueLabel = fmt.Sprintf(" (Due: %s)", t.DueDate.Format("2006-01-02"))
							now := time.Now().Truncate(24 * time.Hour)
							if t.DueDate.Before(now) && !t.Completed {
								dueLabel = " ⚠️ " + dueLabel
							}
						}
						tagsLabel := ""
						if len(t.Tags) > 0 {
							tagsLabel = fmt.Sprintf(" [%s]", strings.Join(t.Tags, ","))
						}

						indentStr := strings.Repeat("  ", indent)
						fmt.Printf("%s%d. %s [%s] %s%s%s\n", indentStr, t.ID, status, priorityLabel, t.Description, dueLabel, tagsLabel)
						printed[t.ID] = true

						// Print children
						for _, child := range filteredTasks {
							if child.ParentID != nil && *child.ParentID == t.ID {
								printTask(child.ID, indent+1)
							}
						}
						break
					}
				}
			}

			// Print all top-level tasks
			for _, t := range filteredTasks {
				if t.ParentID == nil {
					printTask(t.ID, 0)
				}
			}

			// If we have an empty list but we only have child tasks (which shouldn't happen normally)
			// the above handles it.
		},
	}
	listCmd.Flags().BoolVarP(&filterPending, "pending", "p", false, "Show only pending tasks")
	listCmd.Flags().BoolVarP(&filterCompleted, "completed", "c", false, "Show only completed tasks")
	listCmd.Flags().BoolVarP(&sortByPriority, "priority", "s", false, "Sort by priority (High to Low)")
	listCmd.Flags().StringVarP(&filterTag, "tag", "t", "", "Filter by tag")

	// TUI command
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI",
		Run: func(cmd *cobra.Command, args []string) {
			tuiApp := NewTaskApp(store)
			tuiApp.Init()
			if err := tuiApp.Run(); err != nil {
				fmt.Printf("Error running TUI: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(addCmd, doneCmd, deleteCmd, editCmd, priorityCmd, dueCmd, tagCmd, moveCmd, listCmd, tuiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
