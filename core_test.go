package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_Add(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	t.Run("Add first task", func(t *testing.T) {
		task, err := store.Add("First Task")
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
		if task.ID != 1 {
			t.Errorf("Expected ID 1, got %d", task.ID)
		}
		if task.Description != "First Task" {
			t.Errorf("Expected description 'First Task', got '%s'", task.Description)
		}
	})

	t.Run("Add second task", func(t *testing.T) {
		task, err := store.Add("Second Task")
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
		if task.ID != 2 {
			t.Errorf("Expected ID 2, got %d", task.ID)
		}
	})
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Setup: Add some tasks
	store.Add("Task 1")
	store.Add("Task 2")
	store.Add("Task 3")

	t.Run("Delete existing task", func(t *testing.T) {
		err := store.Delete(2)
		if err != nil {
			t.Fatalf("Failed to delete task: %v", err)
		}

		tasks, _ := store.Read()
		if len(tasks) != 2 {
			t.Errorf("Expected 2 tasks remaining, got %d", len(tasks))
		}
		for _, task := range tasks {
			if task.ID == 2 {
				t.Error("Task with ID 2 should have been deleted")
			}
		}
	})

	t.Run("Delete non-existent task", func(t *testing.T) {
		err := store.Delete(99)
		if err == nil {
			t.Error("Expected error when deleting non-existent task, got nil")
		}
	})
}

func TestStore_ReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	tasks := []Task{
		{ID: 1, Description: "Task A"},
		{ID: 2, Description: "Task B"},
	}

	if err := store.Write(tasks); err != nil {
		t.Fatalf("Failed to write tasks: %v", err)
	}

	readTasks, err := store.Read()
	if err != nil {
		t.Fatalf("Failed to read tasks: %v", err)
	}

	if len(readTasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(readTasks))
	}
	if readTasks[0].Description != "Task A" || readTasks[1].Description != "Task B" {
		t.Errorf("Tasks read back do not match original data: %v", readTasks)
	}
}

func TestMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "tasks.json")
	txtPath := filepath.Join(tmpDir, "tasks.txt")

	// Create a legacy tasks.txt file
	legacyContent := "1. Legacy Task 1\n2. Legacy Task 2\n3. Legacy Task 3\n"
	if err := os.WriteFile(txtPath, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("Failed to create legacy file: %v", err)
	}

	if err := Migrate(jsonPath, txtPath); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify JSON file was created
	store := NewStore(jsonPath)
	tasks, err := store.Read()
	if err != nil {
		t.Fatalf("Failed to read migrated JSON: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 migrated tasks, got %d", len(tasks))
	}

	if tasks[0].Description != "Legacy Task 1" || tasks[2].ID != 3 {
		t.Errorf("Migrated data is incorrect: %v", tasks)
	}

	// Verify legacy file was backed up
	if _, err := os.Stat(txtPath); !os.IsNotExist(err) {
		t.Error("Legacy file should have been renamed to .bak")
	}
	if _, err := os.Stat(txtPath + ".bak"); os.IsNotExist(err) {
		t.Error("Backup file .bak should exist")
	}
}
