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
		if task.Completed {
			t.Error("Expected new task to be incomplete")
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

func TestStore_ToggleCompleted(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Test Completion")

	t.Run("Mark as completed", func(t *testing.T) {
		err := store.ToggleCompleted(task.ID)
		if err != nil {
			t.Fatalf("Failed to toggle completion: %v", err)
		}

		tasks, _ := store.Read()
		if !tasks[0].Completed {
			t.Error("Expected task to be completed")
		}
	})

	t.Run("Mark as incomplete", func(t *testing.T) {
		err := store.ToggleCompleted(task.ID)
		if err != nil {
			t.Fatalf("Failed to toggle completion: %v", err)
		}

		tasks, _ := store.Read()
		if tasks[0].Completed {
			t.Error("Expected task to be incomplete")
		}
	})

	t.Run("Toggle non-existent task", func(t *testing.T) {
		err := store.ToggleCompleted(99)
		if err == nil {
			t.Error("Expected error when toggling non-existent task, got nil")
		}
	})
}

func TestStore_UpdateDescription(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Original Description")

	t.Run("Update existing task", func(t *testing.T) {
		err := store.UpdateDescription(task.ID, "Updated Description")
		if err != nil {
			t.Fatalf("Failed to update description: %v", err)
		}

		tasks, _ := store.Read()
		if tasks[0].Description != "Updated Description" {
			t.Errorf("Expected 'Updated Description', got '%s'", tasks[0].Description)
		}
	})

	t.Run("Update non-existent task", func(t *testing.T) {
		err := store.UpdateDescription(99, "Doesn't Matter")
		if err == nil {
			t.Error("Expected error when updating non-existent task, got nil")
		}
	})
}

func TestStore_ReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	tasks := []Task{
		{ID: 1, Description: "Task A", Completed: false},
		{ID: 2, Description: "Task B", Completed: true},
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
	if readTasks[1].Completed != true {
		t.Error("Expected second task to be completed")
	}
}

func TestMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "tasks.json")
	txtPath := filepath.Join(tmpDir, "tasks.txt")

	legacyContent := "1. Legacy Task 1\n2. Legacy Task 2\n3. Legacy Task 3\n"
	if err := os.WriteFile(txtPath, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("Failed to create legacy file: %v", err)
	}

	if err := Migrate(jsonPath, txtPath); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

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

	if _, err := os.Stat(txtPath); !os.IsNotExist(err) {
		t.Error("Legacy file should have been renamed to .bak")
	}
	if _, err := os.Stat(txtPath + ".bak"); os.IsNotExist(err) {
		t.Error("Backup file .bak should exist")
	}
}
