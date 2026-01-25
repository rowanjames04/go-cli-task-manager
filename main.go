package main

import (
	"bufio"
	"os"
	"strings"
	"fmt"
)

type Task struct {
	Description string
}

func main() {
	tasks := make([]Task, 0)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nTask Manager")
		fmt.Println("1. Add Task")
		fmt.Println("2. List Tasks")
		fmt.Println("3. Exit")
		fmt.Print("Choose option: ")

		input,_ := reader.ReadString('\n')
		
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Println("Enter the task description: ")
			
			description,_ := reader.ReadString('\n')
			description = strings.TrimSpace(description)

			tasks = append(tasks, Task{Description: description})

			fmt.Println("Task added")

		case "2":
			if len(tasks) < 1{
				fmt.Println("No tasks")
			} else {
				fmt.Println("Tasks: ")
				for i, task := range tasks {
					fmt.Println("%d. %s\n", i + 1, task.Description)
				}
			}

		case "3":
			fmt.Println("Exiting")
			os.Exit(0)

		default:
			fmt.Println("Invalid option")
		}
	}
}