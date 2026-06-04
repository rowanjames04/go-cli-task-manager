package main

import (
	"path/filepath"
	"testing"
	"time"
)

// TestTaskApp_loadTasks verifies that the TUI can load tasks from the store
func TestTaskApp_loadTasks(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Add some test tasks
	store.Add("Task 1", 1, nil, []string{"work"}, nil)
	store.Add("Task 2", 3, nil, []string{"home"}, nil)
	store.Add("Task 3", 2, nil, nil, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	if len(app.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(app.tasks))
	}
}

// TestTaskApp_updateList_sorting verifies tasks are sorted correctly
func TestTaskApp_updateList_sorting(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Add tasks with different priorities and completion status
	store.Add("Low Pending", 1, nil, nil, nil)
	store.Add("High Pending", 3, nil, nil, nil)
	store.Add("Med Pending", 2, nil, nil, nil)

	task4, _ := store.Add("Completed High", 3, nil, nil, nil)
	store.ToggleCompleted(task4.ID)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	// Verify sorting: pending first (by priority), then completed
	if len(app.tasks) != 4 {
		t.Fatalf("Expected 4 tasks, got %d", len(app.tasks))
	}

	// First task should be High priority pending (priority 3)
	if app.tasks[0].Description != "High Pending" {
		t.Errorf("Expected first task to be 'High Pending', got '%s'", app.tasks[0].Description)
	}

	// Last task should be the completed one
	if !app.tasks[3].Completed {
		t.Error("Expected last task to be completed")
	}
}

// TestTaskApp_filterTasks verifies the filter functionality
func TestTaskApp_filterTasks(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	store.Add("Work task", 2, nil, []string{"work", "urgent"}, nil)
	store.Add("Home task", 2, nil, []string{"home"}, nil)
	store.Add("Work project", 3, nil, []string{"work"}, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	// Verify initial load
	if len(app.tasks) != 3 {
		t.Errorf("Expected 3 tasks initially, got %d", len(app.tasks))
	}

	// Test that filter updates the list display (items count)
	app.filterTasks("work")
	if app.taskList.GetItemCount() != 2 {
		t.Errorf("Expected 2 tasks matching 'work' in list, got %d", app.taskList.GetItemCount())
	}

	// Reload all tasks
	app.loadTasks()

	// Test filtering by tag
	app.filterTasks("urgent")
	if app.taskList.GetItemCount() != 1 {
		t.Errorf("Expected 1 task matching 'urgent' in list, got %d", app.taskList.GetItemCount())
	}

	// Test empty filter reloads all
	app.filterTasks("")
	if app.taskList.GetItemCount() != 3 {
		t.Errorf("Expected 3 tasks after clearing filter, got %d", app.taskList.GetItemCount())
	}
}

// TestTaskApp_updateDetails verifies task details are generated correctly
func TestTaskApp_updateDetails(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Test Details", 3, nil, []string{"test"}, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()
	app.selectedID = task.ID
	app.updateDetails()

	details := app.taskDetails.GetText(false)

	if details == "" {
		t.Error("Expected task details to be non-empty")
	}

	// Verify key information is present
	expectedStrings := []string{"Test Details", "High", "Pending", "test"}
	for _, expected := range expectedStrings {
		if !contains(details, expected) {
			t.Errorf("Expected details to contain '%s', got: %s", expected, details)
		}
	}
}

// TestTaskApp_toggleCompletion verifies toggling task completion works
func TestTaskApp_toggleCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	task, _ := store.Add("Toggle Test", 2, nil, nil, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()
	app.selectedID = task.ID

	// Initially should be incomplete
	tasks, _ := store.Read()
	if tasks[0].Completed {
		t.Error("Expected task to be initially incomplete")
	}

	// Toggle to complete
	app.toggleCompletion()

	tasks, _ = store.Read()
	if !tasks[0].Completed {
		t.Error("Expected task to be completed after toggle")
	}

	// Toggle back to incomplete
	app.toggleCompletion()

	tasks, _ = store.Read()
	if tasks[0].Completed {
		t.Error("Expected task to be incomplete after second toggle")
	}
}

// TestTaskApp_showErrorModal verifies error modal creation
func TestTaskApp_showErrorModal(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	app.showErrorModal("Test error message")

	if !app.pages.HasPage("modal") {
		t.Error("Expected modal page to be added")
	}
}

// TestTaskApp_selectedIDAutoSelect verifies selectedID auto-selects first task
func TestTaskApp_selectedIDAutoSelect(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	store.Add("First Task", 2, nil, nil, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()
	app.selectedID = 0 // Reset to 0
	app.updateDetails()

	if app.selectedID == 0 {
		t.Error("Expected selectedID to auto-select first task")
	}
}

// TestTaskApp_overdueDetection verifies overdue tasks are detected
func TestTaskApp_overdueDetection(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Create a past due date
	pastDate := time.Now().AddDate(0, 0, -7)
	store.Add("Overdue Task", 2, &pastDate, nil, nil)

	// Create a future due date
	futureDate := time.Now().AddDate(0, 0, 7)
	store.Add("Future Task", 2, &futureDate, nil, nil)

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	// Check that overdue task is detected
	app.selectedID = app.tasks[0].ID
	app.updateDetails()
	details := app.taskDetails.GetText(false)

	if !contains(details, "OVERDUE") {
		t.Errorf("Expected overdue indicator in details, got: %s", details)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkTaskApp_loadTasks benchmarks loading tasks
func BenchmarkTaskApp_loadTasks(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Add 100 tasks
	for i := 0; i < 100; i++ {
		store.Add("Benchmark Task", 2, nil, []string{"tag1", "tag2"}, nil)
	}

	app := NewTaskApp(store)
	app.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.loadTasks()
	}
}

// BenchmarkTaskApp_filterTasks benchmarks filtering tasks
func BenchmarkTaskApp_filterTasks(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "tasks.json")
	store := NewStore(tmpFile)

	// Add 100 tasks
	for i := 0; i < 100; i++ {
		store.Add("Benchmark Task", 2, nil, []string{"tag"}, nil)
	}

	app := NewTaskApp(store)
	app.Init()
	app.loadTasks()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.filterTasks("Benchmark")
	}
}
