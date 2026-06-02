package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const taskFile = "tasks.json"
const legacyFile = "tasks.txt"

func main() {
	store := NewStore(taskFile)

	// Run migration from legacy tasks.txt if it exists
	if err := Migrate(taskFile, legacyFile); err != nil {
		fmt.Printf("Migration error: %v\n", err)
	}

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		handleAdd(store)
	case "edit":
		handleEdit(store)
	case "delete":
		handleDelete(store)
	case "done":
		handleDone(store)
	case "list":
		handleList(store)
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func handleAdd(store *Store) {
	if len(os.Args) < 3 {
		fmt.Println("Missing task description")
		return
	}

	taskDescription := strings.Join(os.Args[2:], " ")

	task, err := store.Add(taskDescription)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Task added: %s (ID: %d)\n", task.Description, task.ID)
}

func handleDelete(store *Store) {
	if len(os.Args) < 3 {
		fmt.Println("Missing task number")
		return
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Task number must be an integer")
		return
	}

	if err := store.Delete(num); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Deleted task", num)
}

func handleEdit(store *Store) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: edit <task number> \"new description\"")
		return
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Task number must be an integer")
		return
	}

	newDescription := strings.Join(os.Args[3:], " ")

	if err := store.UpdateDescription(num, newDescription); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Task %d updated to: %s\n", num, newDescription)
}

func handleDone(store *Store) {
	if len(os.Args) < 3 {
		fmt.Println("Missing task number")
		return
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Task number must be an integer")
		return
	}

	if err := store.ToggleCompleted(num); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Toggled completion for task %d\n", num)
}

func handleList(store *Store) {
	tasks, err := store.Read()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Filtering
	showAll := false
	filterPending := false
	filterCompleted := false

	for _, arg := range os.Args[2:] {
		switch arg {
		case "--pending":
			filterPending = true
			showAll = false
		case "--completed":
			filterCompleted = true
			showAll = false
		}
	}

	// If no specific filter flag was provided, we show all (default behavior)
	// but if we want to support a clear --all flag, we can.
	// For now, if neither is set, we just list everything.
	if !filterPending && !filterCompleted {
		showAll = true
	}

	var filteredTasks []Task
	for _, task := range tasks {
		if showAll || (filterPending && !task.Completed) || (filterCompleted && task.Completed) {
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
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add \"task description\"")
	fmt.Println("  edit <task number> \"new description\"")
	fmt.Println("  done <task number>")
	fmt.Println("  delete <task number>")
	fmt.Println("  list [--pending | --completed]")
}
