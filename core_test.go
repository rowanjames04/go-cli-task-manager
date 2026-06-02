package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_Add(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	t.Run("Add first task", func(t *testing.T) {
		task, err := store.Add("First Task", 2, nil, []string{"tag1", "tag2"}, nil)
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
		if task.Priority != 2 {
			t.Errorf("Expected priority 2, got %d", task.Priority)
		}
		if len(task.Tags) != 2 || task.Tags[0] != "tag1" || task.Tags[1] != "tag2" {
			t.Errorf("Expected tags [tag1, tag2], got %v", task.Tags)
		}
	})

	t.Run("Add second task", func(t *testing.T) {
		task, err := store.Add("Second Task", 3, nil, nil, nil)
		if err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}
		if task.ID != 2 {
			t.Errorf("Expected ID 2, got %d", task.ID)
		}
		if task.Priority != 3 {
			t.Errorf("Expected priority 3, got %d", task.Priority)
		}
	})
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	store.Add("Task 1", 2, nil, nil, nil)
	store.Add("Task 2", 2, nil, nil, nil)
	store.Add("Task 3", 2, nil, nil, nil)

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

	task, _ := store.Add("Test Completion", 2, nil, nil, nil)

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

	task, _ := store.Add("Original Description", 2, nil, nil, nil)

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

func TestStore_SetPriority(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Test Priority", 2, nil, nil, nil)

	t.Run("Set priority", func(t *testing.T) {
		err := store.SetPriority(task.ID, 3)
		if err != nil {
			t.Fatalf("Failed to set priority: %v", err)
		}

		tasks, _ := store.Read()
		if tasks[0].Priority != 3 {
			t.Errorf("Expected priority 3, got %d", tasks[0].Priority)
		}
	})

	t.Run("Set priority non-existent task", func(t *testing.T) {
		err := store.SetPriority(99, 1)
		if err == nil {
			t.Error("Expected error when setting priority for non-existent task, got nil")
		}
	})
}

func TestStore_SetDueDate(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Due Task", 2, nil, nil, nil)
	now := time.Now()

	t.Run("Set due date", func(t *testing.T) {
		err := store.SetDueDate(task.ID, &now)
		if err != nil {
			t.Fatalf("Failed to set due date: %v", err)
		}

		tasks, _ := store.Read()
		if tasks[0].DueDate == nil || !tasks[0].DueDate.Equal(now) {
			t.Errorf("Expected due date %v, got %v", now, tasks[0].DueDate)
		}
	})

	t.Run("Set due date non-existent task", func(t *testing.T) {
		err := store.SetDueDate(99, &now)
		if err == nil {
			t.Error("Expected error when setting due date for non-existent task, got nil")
		}
	})
}

func TestStore_AddTag(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Tag Test", 2, nil, []string{"work"}, nil)

	t.Run("Add new tag", func(t *testing.T) {
		err := store.AddTag(task.ID, "urgent")
		if err != nil {
			t.Fatalf("Failed to add tag: %v", err)
		}
		tasks, _ := store.Read()
		if len(tasks[0].Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tasks[0].Tags))
		}
	})

	t.Run("Add duplicate tag", func(t *testing.T) {
		err := store.AddTag(task.ID, "work")
		if err != nil {
			t.Fatalf("Failed to add duplicate tag: %v", err)
		}
		tasks, _ := store.Read()
		if len(tasks[0].Tags) != 2 {
			t.Errorf("Expected tags to remain at 2, got %d", len(tasks[0].Tags))
		}
	})

	t.Run("Add tag to non-existent task", func(t *testing.T) {
		err := store.AddTag(99, "ghost")
		if err == nil {
			t.Error("Expected error when adding tag to non-existent task, got nil")
		}
	})
}

func TestStore_RemoveTag(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Tag Test", 2, nil, []string{"work", "home"}, nil)

	t.Run("Remove existing tag", func(t *testing.T) {
		err := store.RemoveTag(task.ID, "home")
		if err != nil {
			t.Fatalf("Failed to remove tag: %v", err)
		}
		tasks, _ := store.Read()
		if len(tasks[0].Tags) != 1 || tasks[0].Tags[0] != "work" {
			t.Errorf("Expected tags to be [work], got %v", tasks[0].Tags)
		}
	})

	t.Run("Remove non-existent tag", func(t *testing.T) {
		err := store.RemoveTag(task.ID, "ghost")
		if err != nil {
			t.Errorf("Removing non-existent tag should be a no-op, but got error: %v", err)
		}
		tasks, _ := store.Read()
		if len(tasks[0].Tags) != 1 {
			t.Error("Removing non-existent tag should not affect other tags")
		}
	})

	t.Run("Remove tag from non-existent task", func(t *testing.T) {
		err := store.RemoveTag(99, "work")
		if err == nil {
			t.Error("Expected error when removing tag from non-existent task, got nil")
		}
	})
}

func TestStore_MoveTask(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	parent, _ := store.Add("Parent Task", 2, nil, nil, nil)
	child, _ := store.Add("Child Task", 2, nil, nil, &parent.ID)

	t.Run("Move task to new parent", func(t *testing.T) {
		parent2, _ := store.Add("Parent 2", 2, nil, nil, nil)
		err := store.MoveTask(child.ID, &parent2.ID)
		if err != nil {
			t.Fatalf("Failed to move task: %v", err)
		}
		tasks, _ := store.Read()
		for _, task := range tasks {
			if task.ID == child.ID {
				if task.ParentID == nil || *task.ParentID != parent2.ID {
					t.Errorf("Expected parent ID %d, got %v", parent2.ID, task.ParentID)
				}
			}
		}
	})

	t.Run("Move task to root", func(t *testing.T) {
		err := store.MoveTask(child.ID, nil)
		if err != nil {
			t.Fatalf("Failed to move task to root: %v", err)
		}
		tasks, _ := store.Read()
		for _, task := range tasks {
			if task.ID == child.ID {
				if task.ParentID != nil {
					t.Errorf("Expected no parent, got %v", task.ParentID)
				}
			}
		}
	})

	t.Run("Move non-existent task", func(t *testing.T) {
		err := store.MoveTask(99, nil)
		if err == nil {
			t.Error("Expected error when moving non-existent task, got nil")
		}
	})
}

func TestStore_ReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	tasks := []Task{
		{ID: 1, Description: "Task A", Completed: false, Priority: 1, Tags: []string{"a"}},
		{ID: 2, Description: "Task B", Completed: true, Priority: 3, Tags: []string{"b"}},
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
