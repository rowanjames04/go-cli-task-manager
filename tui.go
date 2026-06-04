package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TaskApp represents the TUI application
type TaskApp struct {
	app           *tview.Application
	pages         *tview.Pages
	store         *Store
	taskList      *tview.List
	taskDetails   *tview.TextView
	helpText      *tview.TextView
	filterInput   *tview.InputField
	tasks         []Task
	selectedID    int
}

// NewTaskApp creates a new TUI application
func NewTaskApp(store *Store) *TaskApp {
	return &TaskApp{
		app:   tview.NewApplication(),
		store: store,
		tasks: []Task{},
	}
}

// Init initializes the TUI components
func (ta *TaskApp) Init() {
	// Create the main list
	ta.taskList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedBackgroundColor(tcell.ColorDarkCyan)

	// Create the details panel
	ta.taskDetails = tview.NewTextView().
		SetDynamicColors(true)
	ta.taskDetails.SetBorder(true).
		SetTitle(" Task Details ")

	// Create help text
	ta.helpText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Create filter input
	ta.filterInput = tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(30).
		SetChangedFunc(func(text string) {
			ta.filterTasks(text)
		})

	// Set up list selection handler
	ta.taskList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ta.selectedID = ta.tasks[index].ID
		ta.updateDetails()
	})

	// Load initial tasks
	ta.loadTasks()

	// Create the main layout
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ta.filterInput, 1, 0, false).
		AddItem(ta.taskList, 0, 1, true)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftPanel, 40, 1, true).
		AddItem(ta.taskDetails, 0, 1, false)

	// Create footer with help
	ta.updateHelpText()

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(ta.helpText, 1, 0, false)

	// Create pages for modals
	ta.pages = tview.NewPages().
		AddPage("main", mainLayout, true, true)

	// Set up global key bindings
	ta.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle key events when not in a modal
		if ta.pages.HasPage("modal") {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlQ:
			ta.app.Stop()
			return nil
		case tcell.KeyCtrlN:
			ta.showAddTaskModal()
			return nil
		case tcell.KeyCtrlD:
			ta.showDeleteConfirmModal()
			return nil
		case tcell.KeyCtrlE:
			ta.showEditTaskModal()
			return nil
		case tcell.KeyCtrlT:
			ta.showTagModal()
			return nil
		case tcell.KeyCtrlP:
			ta.showPriorityModal()
			return nil
		case tcell.KeyCtrlF:
			ta.showDueDateModal()
			return nil
		case tcell.KeyCtrlL:
			ta.app.SetFocus(ta.taskList)
			return nil
		case tcell.KeyCtrlH:
			ta.toggleHelp()
			return nil
		case tcell.KeyCtrlC:
			ta.toggleCompletion()
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			// Let the list handle navigation
			return event
		}

		// Number keys for quick selection
		if event.Rune() >= '1' && event.Rune() <= '9' {
			index := int(event.Rune() - '1')
			if index < len(ta.tasks) {
				ta.taskList.SetCurrentItem(index)
				ta.selectedID = ta.tasks[index].ID
				ta.updateDetails()
			}
			return nil
		}

		return event
	})

	ta.app.SetRoot(ta.pages, true)
}

func (ta *TaskApp) loadTasks() {
	tasks, err := ta.store.Read()
	if err != nil {
		ta.taskDetails.SetText(fmt.Sprintf("[red]Error loading tasks: %v[-]", err))
		return
	}
	ta.tasks = tasks
	ta.updateList()
}

func (ta *TaskApp) updateList() {
	ta.taskList.Clear()

	// Sort tasks: pending first, then by priority (high to low)
	sortedTasks := make([]Task, len(ta.tasks))
	copy(sortedTasks, ta.tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		// Completed tasks go to the bottom
		if sortedTasks[i].Completed != sortedTasks[j].Completed {
			return !sortedTasks[i].Completed
		}
		// Then by priority (higher first)
		return sortedTasks[i].Priority > sortedTasks[j].Priority
	})

	for i, task := range sortedTasks {
		status := "[ ]"
		color := "[white]"
		if task.Completed {
			status = "[green][x][-]"
			color = "[darkgray]"
		}

		priorityLabel := ""
		priorityColor := "[yellow]"
		switch task.Priority {
		case 1:
			priorityLabel = "Low"
			priorityColor = "[green]"
		case 2:
			priorityLabel = "Med"
		case 3:
			priorityLabel = "High"
			priorityColor = "[red]"
		}

		// Check for overdue
		overdue := ""
		if task.DueDate != nil && !task.Completed {
			now := time.Now().Truncate(24 * time.Hour)
			if task.DueDate.Before(now) {
				overdue = " [red]⚠️ OVERDUE[-]"
			}
		}

		dueLabel := ""
		if task.DueDate != nil {
			dueLabel = fmt.Sprintf(" [blue]📅 %s[-]", task.DueDate.Format("2006-01-02"))
		}

		tagsLabel := ""
		if len(task.Tags) > 0 {
			tagsLabel = fmt.Sprintf(" [%s]", strings.Join(task.Tags, ","))
		}

		mainText := fmt.Sprintf("%s %s%d. %s", status, color, task.ID, task.Description)
		secondaryText := fmt.Sprintf("%sPriority: %s%s%s%s%s",
			priorityColor, priorityLabel, "[-]",
			dueLabel, overdue, tagsLabel)

		ta.taskList.AddItem(mainText, secondaryText, rune('0'+i%10), nil)
	}

	ta.tasks = sortedTasks
}

func (ta *TaskApp) filterTasks(filter string) {
	filter = strings.ToLower(filter)
	if filter == "" {
		ta.loadTasks()
		return
	}

	var filtered []Task
	for _, task := range ta.tasks {
		descLower := strings.ToLower(task.Description)
		tagsLower := ""
		for _, tag := range task.Tags {
			tagsLower += strings.ToLower(tag) + " "
		}

		if strings.Contains(descLower, filter) || strings.Contains(tagsLower, filter) {
			filtered = append(filtered, task)
		}
	}

	ta.taskList.Clear()
	for i, task := range filtered {
		status := "[ ]"
		if task.Completed {
			status = "[green][x][-]"
		}

		priorityLabel := ""
		switch task.Priority {
		case 1:
			priorityLabel = "[green]Low[-]"
		case 2:
			priorityLabel = "Med"
		case 3:
			priorityLabel = "[red]High[-]"
		}

		dueLabel := ""
		if task.DueDate != nil {
			dueLabel = fmt.Sprintf(" [blue]📅 %s[-]", task.DueDate.Format("2006-01-02"))
		}

		tagsLabel := ""
		if len(task.Tags) > 0 {
			tagsLabel = fmt.Sprintf(" [%s]", strings.Join(task.Tags, ","))
		}

		mainText := fmt.Sprintf("%s %d. %s", status, task.ID, task.Description)
		secondaryText := fmt.Sprintf("Priority: %s%s%s", priorityLabel, dueLabel, tagsLabel)
		ta.taskList.AddItem(mainText, secondaryText, rune('0'+i%10), nil)
	}
}

func (ta *TaskApp) updateDetails() {
	if ta.selectedID == 0 && len(ta.tasks) > 0 {
		ta.selectedID = ta.tasks[0].ID
	}

	var task *Task
	for i := range ta.tasks {
		if ta.tasks[i].ID == ta.selectedID {
			task = &ta.tasks[i]
			break
		}
	}

	if task == nil {
		ta.taskDetails.SetText("[gray]No task selected[-]")
		return
	}

	status := "Pending"
	statusColor := "yellow"
	if task.Completed {
		status = "Completed"
		statusColor = "green"
	}

	priorityLabel := ""
	priorityColor := "yellow"
	switch task.Priority {
	case 1:
		priorityLabel = "Low"
		priorityColor = "green"
	case 2:
		priorityLabel = "Medium"
	case 3:
		priorityLabel = "High"
		priorityColor = "red"
	}

	dueDate := "Not set"
	if task.DueDate != nil {
		dueDate = task.DueDate.Format("January 2, 2006")
		now := time.Now().Truncate(24 * time.Hour)
		if task.DueDate.Before(now) && !task.Completed {
			dueDate = fmt.Sprintf("[red]%s (OVERDUE)[-]", dueDate)
		}
	}

	tags := "None"
	if len(task.Tags) > 0 {
		tags = strings.Join(task.Tags, ", ")
	}

	parentInfo := "None"
	if task.ParentID != nil {
		parentInfo = fmt.Sprintf("Task #%d", *task.ParentID)
	}

	details := fmt.Sprintf(`[%s]●[%s] [bold]%s[-]

[white]Status:[-]      [%s]%s[-]
[white]Priority:[-]    [%s]%s[-]
[white]Due Date:[-]    %s
[white]Tags:[-]        %s
[white]Parent:[-]      %s
[white]Created:[-]     ID #%d

[gray]Press Ctrl+E to edit | Ctrl+D to delete | Ctrl+C to toggle completion[-]`,
		statusColor, statusColor, task.Description,
		statusColor, status,
		priorityColor, priorityLabel,
		dueDate,
		tags,
		parentInfo,
		task.ID,
	)

	ta.taskDetails.SetText(details)
}

func (ta *TaskApp) updateHelpText() {
	help := `[gray]
 Ctrl+N: Add | Ctrl+E: Edit | Ctrl+D: Delete | Ctrl+C: Toggle Done |
 Ctrl+P: Priority | Ctrl+F: Due Date | Ctrl+T: Tags | Ctrl+Q: Quit |
 ↑/↓: Navigate | 1-9: Quick Select | Ctrl+H: Toggle Help
[-]`
	ta.helpText.SetText(help)
}

func (ta *TaskApp) toggleHelp() {
	currentText := ta.helpText.GetText(false)
	if strings.Contains(currentText, "Ctrl+N") {
		ta.helpText.SetText("")
	} else {
		ta.updateHelpText()
	}
}

func (ta *TaskApp) showAddTaskModal() {
	var form *tview.Form
	form = tview.NewForm().
		AddInputField("Description", "", 50, nil, nil).
		AddInputField("Tags (comma-separated)", "", 30, nil, nil).
		AddDropDown("Priority", []string{"Low", "Medium", "High"}, 1, nil).
		AddButton("Add", func() {
			desc := form.GetFormItem(0).(*tview.InputField).GetText()
			tagsStr := form.GetFormItem(1).(*tview.InputField).GetText()
			dropDown := form.GetFormItem(2).(*tview.DropDown)
			priorityIdx, _ := dropDown.GetCurrentOption()

			// priorityIdx is the index: 0=Low, 1=Medium, 2=High
			// convert to our priority: 1=Low, 2=Medium, 3=High
			priority := priorityIdx + 1

			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}

			task, err := ta.store.Add(desc, priority, nil, tags, nil)
			if err != nil {
				ta.showErrorModal(fmt.Sprintf("Error adding task: %v", err))
			} else {
				ta.loadTasks()
				ta.selectedID = task.ID
				ta.updateDetails()
			}
			ta.pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			ta.pages.RemovePage("modal")
		})

	form.SetBorder(true).
		SetTitle(" Add New Task ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ta.pages.RemovePage("modal")
				return nil
			}
			return event
		})

	ta.pages.AddPage("modal", form, true, true)
	ta.app.SetFocus(form)
}

func (ta *TaskApp) showEditTaskModal() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	var task *Task
	for i := range ta.tasks {
		if ta.tasks[i].ID == ta.selectedID {
			task = &ta.tasks[i]
			break
		}
	}

	if task == nil {
		ta.showErrorModal("Task not found")
		return
	}

	var form *tview.Form
	form = tview.NewForm().
		AddInputField("Description", task.Description, 50, nil, nil).
		AddButton("Save", func() {
			desc := form.GetFormItem(0).(*tview.InputField).GetText()
			if err := ta.store.UpdateDescription(task.ID, desc); err != nil {
				ta.showErrorModal(fmt.Sprintf("Error updating task: %v", err))
			} else {
				ta.loadTasks()
				ta.updateDetails()
			}
			ta.pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			ta.pages.RemovePage("modal")
		})

	form.SetBorder(true).
		SetTitle(" Edit Task ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ta.pages.RemovePage("modal")
				return nil
			}
			return event
		})

	ta.pages.AddPage("modal", form, true, true)
	ta.app.SetFocus(form)
}

func (ta *TaskApp) showDeleteConfirmModal() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete task #%d? This cannot be undone.", ta.selectedID)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := ta.store.Delete(ta.selectedID); err != nil {
					ta.showErrorModal(fmt.Sprintf("Error deleting task: %v", err))
				} else {
					ta.selectedID = 0
					ta.loadTasks()
					ta.updateDetails()
				}
			}
			ta.pages.RemovePage("modal")
		})

	modal.SetBorder(true).SetTitle(" Confirm Delete ")
	ta.pages.AddPage("modal", modal, true, true)
	ta.app.SetFocus(modal)
}

func (ta *TaskApp) toggleCompletion() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	if err := ta.store.ToggleCompleted(ta.selectedID); err != nil {
		ta.showErrorModal(fmt.Sprintf("Error toggling completion: %v", err))
		return
	}

	ta.loadTasks()
	ta.updateDetails()
}

func (ta *TaskApp) showPriorityModal() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	var task *Task
	for i := range ta.tasks {
		if ta.tasks[i].ID == ta.selectedID {
			task = &ta.tasks[i]
			break
		}
	}

	if task == nil {
		ta.showErrorModal("Task not found")
		return
	}

	currentPriority := task.Priority - 1 // Convert to 0-indexed

	var form *tview.Form
	form = tview.NewForm().
		AddDropDown("Priority", []string{"1 - Low", "2 - Medium", "3 - High"}, currentPriority, nil).
		AddButton("Set", func() {
			dropDown := form.GetFormItem(0).(*tview.DropDown)
			priorityIdx, _ := dropDown.GetCurrentOption()

			// priorityIdx is the index: 0=Low, 1=Medium, 2=High
			// convert to our priority: 1=Low, 2=Medium, 3=High
			priority := priorityIdx + 1

			if err := ta.store.SetPriority(task.ID, priority); err != nil {
				ta.showErrorModal(fmt.Sprintf("Error setting priority: %v", err))
			} else {
				ta.loadTasks()
				ta.updateDetails()
			}
			ta.pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			ta.pages.RemovePage("modal")
		})

	form.SetBorder(true).
		SetTitle(" Set Priority ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ta.pages.RemovePage("modal")
				return nil
			}
			return event
		})

	ta.pages.AddPage("modal", form, true, true)
	ta.app.SetFocus(form)
}

func (ta *TaskApp) showDueDateModal() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	var task *Task
	for i := range ta.tasks {
		if ta.tasks[i].ID == ta.selectedID {
			task = &ta.tasks[i]
			break
		}
	}

	if task == nil {
		ta.showErrorModal("Task not found")
		return
	}

	currentDate := ""
	if task.DueDate != nil {
		currentDate = task.DueDate.Format("2006-01-02")
	}

	var form *tview.Form
	form = tview.NewForm().
		AddInputField("Due Date (YYYY-MM-DD)", currentDate, 20, nil, nil).
		AddButton("Set", func() {
			dateStr := form.GetFormItem(0).(*tview.InputField).GetText()
			if dateStr == "" {
				// Clear due date
				if err := ta.store.SetDueDate(task.ID, nil); err != nil {
					ta.showErrorModal(fmt.Sprintf("Error clearing due date: %v", err))
				} else {
					ta.loadTasks()
					ta.updateDetails()
				}
				ta.pages.RemovePage("modal")
				return
			}

			dueDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				ta.showErrorModal("Invalid date format. Use YYYY-MM-DD")
				return
			}

			if err := ta.store.SetDueDate(task.ID, &dueDate); err != nil {
				ta.showErrorModal(fmt.Sprintf("Error setting due date: %v", err))
			} else {
				ta.loadTasks()
				ta.updateDetails()
			}
			ta.pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			ta.pages.RemovePage("modal")
		})

	form.SetBorder(true).
		SetTitle(" Set Due Date ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ta.pages.RemovePage("modal")
				return nil
			}
			return event
		})

	ta.pages.AddPage("modal", form, true, true)
	ta.app.SetFocus(form)
}

func (ta *TaskApp) showTagModal() {
	if ta.selectedID == 0 {
		ta.showErrorModal("No task selected")
		return
	}

	var task *Task
	for i := range ta.tasks {
		if ta.tasks[i].ID == ta.selectedID {
			task = &ta.tasks[i]
			break
		}
	}

	if task == nil {
		ta.showErrorModal("Task not found")
		return
	}

	tagsStr := strings.Join(task.Tags, ", ")

	var form *tview.Form
	form = tview.NewForm().
		AddInputField("Tags (comma-separated)", tagsStr, 40, nil, nil).
		AddButton("Save", func() {
			newTagsStr := form.GetFormItem(0).(*tview.InputField).GetText()

			// Get current tags for comparison
			currentTags := make(map[string]bool)
			for _, tag := range task.Tags {
				currentTags[tag] = true
			}

			// Parse new tags
			var newTags []string
			if newTagsStr != "" {
				newTags = strings.Split(newTagsStr, ",")
				for i := range newTags {
					newTags[i] = strings.TrimSpace(newTags[i])
				}
			}

			// Add new tags
			for _, tag := range newTags {
				if tag != "" && !currentTags[tag] {
					ta.store.AddTag(task.ID, tag)
				}
			}

			// Remove tags that are no longer in the list
			newTagsMap := make(map[string]bool)
			for _, tag := range newTags {
				if tag != "" {
					newTagsMap[tag] = true
				}
			}
			for tag := range currentTags {
				if !newTagsMap[tag] {
					ta.store.RemoveTag(task.ID, tag)
				}
			}

			ta.loadTasks()
			ta.updateDetails()
			ta.pages.RemovePage("modal")
		}).
		AddButton("Cancel", func() {
			ta.pages.RemovePage("modal")
		})

	form.SetBorder(true).
		SetTitle(" Edit Tags ").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ta.pages.RemovePage("modal")
				return nil
			}
			return event
		})

	ta.pages.AddPage("modal", form, true, true)
	ta.app.SetFocus(form)
}

func (ta *TaskApp) showErrorModal(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ta.pages.RemovePage("modal")
		})

	modal.SetBorder(true).SetTitle(" Error ")
	ta.pages.AddPage("modal", modal, true, true)
	ta.app.SetFocus(modal)
}

// Run starts the TUI application
func (ta *TaskApp) Run() error {
	return ta.app.Run()
}
