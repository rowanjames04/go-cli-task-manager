// Package main provides core functionalities for a task management application,
// including reading from and writing to a task file, adding new tasks, and deleting existing tasks.
package main

import (
	"fmt"
	"bufio"
	"strings"
	"os"
)

// readTasks reads tasks from the specified file and returns them as a slice of strings.
// It handles potential file opening errors and strips numerical prefixes from the task lines.
func readTasks(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// strip "1. " prefix
		if idx := strings.Index(line, ". "); idx != -1 {
			line = line[idx+2:]
		}

		tasks = append(tasks, line)
	}

	return tasks, scanner.Err()
}

// writeTasks writes the given slice of tasks to the specified file,
// prefixing each task with its 1-based index.
func writeTasks(filename string, tasks []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for i, task := range tasks {
		fmt.Fprintf(writer, "%d. %s\n", i+1, task)
	}

	return writer.Flush()
}

// addTask reads existing tasks, appends the new task, and writes all tasks back to the file.
// It creates the file if it does not exist.
func addTask(filename string, task string) error {
	tasks, err := readTasks(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	tasks = append(tasks, task)
	return writeTasks(filename, tasks)
}

// deleteTask removes a task by its 1-based number from the specified file.
// It returns an error if the task number is out of bounds.
func deleteTask(filename string, taskNumber int) error {
	tasks, err := readTasks(filename)
	if err != nil {
		return err
	}

	index := taskNumber - 1

	if index < 0 || index >= len(tasks) {
		return fmt.Errorf("task %d does not exist", taskNumber)
	}

	tasks = append(tasks[:index], tasks[index+1:]...)
	return writeTasks(filename, tasks)
}


