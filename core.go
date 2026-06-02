package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Task represents a single task in the system.
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Store handles the persistence of tasks to a JSON file.
type Store struct {
	filePath string
}

// NewStore initializes a new Store with the given file path.
func NewStore(path string) *Store {
	return &Store{filePath: path}
}

// Read retrieves all tasks from the JSON file.
func (s *Store) Read() ([]Task, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var tasks []Task
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		if err.Error() == "EOF" {
			return []Task{}, nil
		}
		return nil, err
	}

	return tasks, nil
}

// Write saves the provided slice of tasks to the JSON file.
func (s *Store) Write(tasks []Task) error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

// Add creates a new task and persists the updated list.
func (s *Store) Add(description string) (Task, error) {
	tasks, err := s.Read()
	if err != nil {
		return Task{}, err
	}

	newID := 1
	if len(tasks) > 0 {
		maxID := 0
		for _, t := range tasks {
			if t.ID > maxID {
				maxID = t.ID
			}
		}
		newID = maxID + 1
	}

	newTask := Task{
		ID:          newID,
		Description: description,
		Completed:   false,
	}

	tasks = append(tasks, newTask)
	if err := s.Write(tasks); err != nil {
		return Task{}, err
	}

	return newTask, nil
}

// Delete removes a task by its ID and persists the updated list.
func (s *Store) Delete(id int) error {
	tasks, err := s.Read()
	if err != nil {
		return err
	}

	found := false
	var updatedTasks []Task
	for _, t := range tasks {
		if t.ID == id {
			found = true
		} else {
			updatedTasks = append(updatedTasks, t)
		}
	}

	if !found {
		return fmt.Errorf("task %d does not exist", id)
	}

	return s.Write(updatedTasks)
}

// ToggleCompleted flips the completion status of a task.
func (s *Store) ToggleCompleted(id int) error {
	tasks, err := s.Read()
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Completed = !tasks[i].Completed
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task %d does not exist", id)
	}

	return s.Write(tasks)
}

// Migrate converts a legacy tasks.txt file to the new JSON format.
func Migrate(jsonPath, txtPath string) error {
	if _, err := os.Stat(jsonPath); err == nil {
		return nil
	}

	if _, err := os.Stat(txtPath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(txtPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	count := 1
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if idx := strings.Index(line, ". "); idx != -1 {
			line = line[idx+2:]
		}
		tasks = append(tasks, Task{ID: count, Description: line, Completed: false})
		count++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	store := NewStore(jsonPath)
	if err := store.Write(tasks); err != nil {
		return err
	}

	return os.Rename(txtPath, txtPath+".bak")
}
