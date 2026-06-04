# CLI Task Manager

A powerful command-line task manager with both CLI and interactive TUI (Terminal User Interface) modes.

## Features

- **Task Management**: Add, edit, delete, and complete tasks
- **Priority Levels**: Low (1), Medium (2), High (3)
- **Due Dates**: Set and track task deadlines with overdue highlighting
- **Tags**: Organize tasks with customizable tags
- **Hierarchical Tasks**: Support for parent/child task relationships
- **Interactive TUI**: Full terminal UI for visual task management

## Installation

```bash
go build -o taskmanager .
```

## Usage

### Interactive TUI

Launch the interactive terminal UI:

```bash
./taskmanager tui
```

#### TUI Key Bindings

| Key | Action |
|-----|--------|
| `Ctrl+N` | Add new task |
| `Ctrl+E` | Edit task description |
| `Ctrl+D` | Delete task |
| `Ctrl+C` | Toggle completion |
| `Ctrl+P` | Set priority |
| `Ctrl+F` | Set due date |
| `Ctrl+T` | Edit tags |
| `Ctrl+Q` | Quit |
| `Ctrl+H` | Toggle help |
| `↑/↓` | Navigate tasks |
| `1-9` | Quick select tasks |

### CLI Commands

#### Add a task
```bash
./taskmanager add "My task description" -p 2 -d 2026-12-31 -t work,urgent
```

Options:
- `-p, --priority`: Priority level (1=Low, 2=Medium, 3=High). Default: 2
- `-d, --due`: Due date in YYYY-MM-DD format
- `-t, --tags`: Comma-separated list of tags
- `--parent`: ID of the parent task

#### List tasks
```bash
./taskmanager list
./taskmanager list -p          # Pending only
./taskmanager list -c          # Completed only
./taskmanager list -s          # Sort by priority
./taskmanager list -t work     # Filter by tag
```

#### Mark as complete
```bash
./taskmanager done 1
```

#### Delete a task
```bash
./taskmanager delete 1
```

#### Edit a task
```bash
./taskmanager edit 1 "New description"
```

#### Set priority
```bash
./taskmanager priority 1 3     # Set task 1 to High priority
```

#### Set due date
```bash
./taskmanager due 1 2026-12-31
```

#### Manage tags
```bash
./taskmanager tag add 1 work
./taskmanager tag remove 1 work
```

#### Move task (change parent)
```bash
./taskmanager move 5 3         # Move task 5 under parent 3
```

## Data Storage

Tasks are stored in `tasks.json` in the current directory. The format is compatible with the legacy `tasks.txt` format - existing tasks.txt files are automatically migrated on first run.

## License

MIT
