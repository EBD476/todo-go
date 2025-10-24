# Todo App with Bubble Tea

A beautiful, interactive terminal-based todo application built with Go and Bubble Tea framework.

## Features

- 🎨 **Beautiful Terminal UI** - Modern, colorful interface with smooth animations
- ⌨️ **Keyboard Navigation** - Intuitive keyboard shortcuts for all operations
- 📝 **Interactive Forms** - Inline editing and adding with text input
- ✅ **Todo Management** - Add, edit, delete, and toggle completion status
- 🎯 **Priority System** - Visual priority indicators (High, Medium, Low)
- 💾 **Persistent Storage** - JSON file storage for data persistence
- 🔍 **Search & Filter** - Built-in search functionality
- 📊 **Status Indicators** - Clear visual feedback for todo status

## Installation

1. **Install Go** (version 1.21 or later)
2. **Clone or download** this repository
3. **Install dependencies:**
   ```bash
   go mod tidy
   ```
4. **Build the application:**
   ```bash
   go build -o todo main.go
   ```

## Usage

### Running the App
```bash
./todo
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `a` | Add new todo |
| `e` | Edit selected todo |
| `d` | Delete selected todo |
| `Space` | Toggle completion status |
| `↑/↓` | Navigate through todos |
| `Enter` | Confirm action (in forms) |
| `Esc` | Cancel action (in forms) |
| `q` | Quit application |

### Features Overview

#### 🎨 **Visual Design**
- **Color-coded priorities**: Red (High), Yellow (Medium), Green (Low)
- **Status indicators**: ✅ Completed, ⏳ Pending
- **Smooth animations**: Elegant transitions and effects
- **Responsive layout**: Adapts to terminal size

#### 📝 **Todo Management**
- **Add todos**: Press `a` and type your todo
- **Edit todos**: Press `e` to edit the selected todo
- **Delete todos**: Press `d` to delete the selected todo
- **Toggle status**: Press `Space` to mark as complete/incomplete

#### 🔍 **Search & Navigation**
- **Built-in search**: Type to filter todos
- **Keyboard navigation**: Use arrow keys to navigate
- **Status bar**: Shows current selection and total count

## Data Storage

Todos are automatically saved to `todos.json` in the same directory as the executable. The file is created automatically when you add your first todo.

### File Format
```json
{
  "todos": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "",
      "completed": false,
      "created_at": "2024-01-01T12:00:00Z",
      "priority": "medium"
    }
  ],
  "next_id": 2
}
```

## Technical Details

### Built With
- **Go 1.21+** - Programming language
- **Bubble Tea** - Terminal UI framework
- **Lip Gloss** - Styling and layout
- **Bubbles** - UI components (list, textinput)

### Architecture
- **Model-View-Update (MVU)** pattern
- **State management** with immutable updates
- **Component-based** UI architecture
- **Event-driven** message passing

## Screenshots

The app features:
- A beautiful list view with color-coded priorities
- Interactive forms for adding/editing
- Smooth animations and transitions
- Responsive design that adapts to terminal size
- Clear visual feedback for all actions

## Development

### Project Structure
```
todo-bubbletea/
├── main.go          # Main application code
├── go.mod           # Go module file
├── README.md        # This file
└── todos.json       # Data storage (created automatically)
```

### Building for Different Platforms
```bash
# Windows
go build -o todo.exe main.go

# Linux
GOOS=linux go build -o todo main.go

# macOS
GOOS=darwin go build -o todo main.go
```

## Comparison with Traditional CLI

| Feature | Traditional CLI | Bubble Tea App |
|---------|----------------|----------------|
| **Interface** | Static text | Interactive UI |
| **Navigation** | Command-based | Keyboard shortcuts |
| **Visual Feedback** | Basic | Rich animations |
| **User Experience** | Functional | Delightful |
| **Learning Curve** | Steep | Intuitive |

## Contributing

Feel free to submit issues and enhancement requests!

## License

This project is open source and available under the MIT License.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components