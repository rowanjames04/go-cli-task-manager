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

	if len(tasks) == 0 {
		fmt.Println("No tasks found 🎉")
		return
	}

	for _, task := range tasks {
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
	fmt.Println("  done <task number>")
	fmt.Println("  delete <task number>")
	fmt.Println("  list")
}
