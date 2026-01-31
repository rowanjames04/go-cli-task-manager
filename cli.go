package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const taskFile = "tasks.txt"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		handleAdd()
	case "delete":
		handleDelete()
	case "list":
		handleList()
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func handleAdd() {
	if len(os.Args) < 3 {
		fmt.Println("Missing task description")
		return
	}

	task := strings.Join(os.Args[2:], " ")

	if err := addTask(taskFile, task); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Task added:", task)
}

func handleDelete() {
	if len(os.Args) < 3 {
		fmt.Println("Missing task number")
		return
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Task number must be an integer")
		return
	}

	if err := deleteTask(taskFile, num); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Deleted task", num)
}

func handleList() {
	tasks, err := readTasks(taskFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found 🎉")
		return
	}

	for i, task := range tasks {
		fmt.Printf("%d. %s\n", i+1, task)
	}
}


func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add \"task description\"")
	fmt.Println("  delete <task number>")
	fmt.Println("  list")
}

