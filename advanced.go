package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Todo represents a single todo item
type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	Priority    string     `json:"priority"`
	Category    string     `json:"category"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// TodoList represents a collection of todos
type TodoList struct {
	Todos  []Todo `json:"todos"`
	NextID int    `json:"next_id"`
}

// Storage file path
const storageFile = "todos.json"

// Styles
var (
	titleStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	itemStyle           = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("#7D56F4"))
	completedStyle      = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("#757575"))
	pendingStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	highPriorityStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	mediumPriorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	lowPriorityStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6BCF7F"))
	helpStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	successStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	infoStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A90E2"))
	warningStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623"))
)

// List item implementation
type todoItem struct {
	todo Todo
}

func (i todoItem) Title() string {
	title := i.todo.Title
	if i.todo.Completed {
		title = completedStyle.Render(title)
	} else {
		title = pendingStyle.Render(title)
	}
	return title
}

func (i todoItem) Description() string {
	desc := i.todo.Description
	if desc == "" {
		desc = "No description"
	}

	// Add priority indicator
	priority := ""
	switch i.todo.Priority {
	case "high":
		priority = highPriorityStyle.Render("🔴 HIGH")
	case "medium":
		priority = mediumPriorityStyle.Render("🟡 MED")
	case "low":
		priority = lowPriorityStyle.Render("🟢 LOW")
	default:
		priority = lowPriorityStyle.Render("🟢 LOW")
	}

	status := "⏳ Pending"
	if i.todo.Completed {
		status = "✅ Completed"
	}

	// Add category
	category := ""
	if i.todo.Category != "" {
		category = fmt.Sprintf(" | 📁 %s", i.todo.Category)
	}

	// Add due date
	dueDate := ""
	if i.todo.DueDate != nil {
		now := time.Now()
		due := *i.todo.DueDate
		if due.Before(now) && !i.todo.Completed {
			dueDate = warningStyle.Render(fmt.Sprintf(" | ⚠️ Overdue (%s)", due.Format("Jan 2")))
		} else if due.Before(now.Add(24*time.Hour)) && !i.todo.Completed {
			dueDate = warningStyle.Render(fmt.Sprintf(" | ⏰ Due soon (%s)", due.Format("Jan 2")))
		} else {
			dueDate = infoStyle.Render(fmt.Sprintf(" | 📅 Due %s", due.Format("Jan 2")))
		}
	}

	return fmt.Sprintf("%s | %s%s%s | %s", priority, status, category, dueDate, desc)
}

func (i todoItem) FilterValue() string {
	return i.todo.Title + " " + i.todo.Description + " " + i.todo.Category
}

// Main model
type model struct {
	todos         []Todo
	list          list.Model
	textInput     textinput.Model
	descInput     textinput.Model
	categoryInput textinput.Model
	state         string // "list", "add", "edit", "add_desc", "add_category", "add_priority", "add_due"
	editingID     int
	nextID        int
	message       string
	messageType   string
	currentField  string
	priority      string
	dueDate       string
}

// Messages
type todoAddedMsg struct{}
type todoUpdatedMsg struct{}
type todoDeletedMsg struct{}
type messageMsg struct {
	text    string
	msgType string
}

// Initial model
func initialModel() model {
	todos, nextID := loadTodos()

	items := make([]list.Item, len(todos))
	for i, todo := range todos {
		items[i] = todoItem{todo: todo}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "📝 Advanced Todo List"
	l.SetShowStatusBar(true)
	l.SetShowFilter(true)
	l.SetShowHelp(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = helpStyle
	l.Styles.HelpStyle = helpStyle

	ti := textinput.New()
	ti.Placeholder = "Enter todo title..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	di := textinput.New()
	di.Placeholder = "Enter description (optional)..."
	di.CharLimit = 200
	di.Width = 50

	ci := textinput.New()
	ci.Placeholder = "Enter category (optional)..."
	ci.CharLimit = 50
	ci.Width = 50

	return model{
		todos:         todos,
		list:          l,
		textInput:     ti,
		descInput:     di,
		categoryInput: ci,
		state:         "list",
		nextID:        nextID,
		priority:      "low",
	}
}

// Commands
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := titleStyle.GetFrameSize()
		m.list.SetSize(msg.Width, msg.Height-h-v-2)

	case tea.KeyMsg:
		switch m.state {
		case "list":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
				m.state = "add"
				m.textInput.Reset()
				m.textInput.Placeholder = "Enter todo title..."
				m.textInput.Focus()
				return m, textinput.Blink

			case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
				if len(m.list.Items()) > 0 {
					selectedItem := m.list.SelectedItem().(todoItem)
					m.state = "edit"
					m.editingID = selectedItem.todo.ID
					m.textInput.SetValue(selectedItem.todo.Title)
					m.textInput.Focus()
					return m, textinput.Blink
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				if len(m.list.Items()) > 0 {
					selectedItem := m.list.SelectedItem().(todoItem)
					m = m.deleteTodo(selectedItem.todo.ID)
					return m, nil
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
				if len(m.list.Items()) > 0 {
					selectedItem := m.list.SelectedItem().(todoItem)
					m = m.toggleTodo(selectedItem.todo.ID)
					return m, nil
				}

			case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
				m = m.sortTodos()
				return m, nil

			case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
				m = m.showCategories()
				return m, nil

			case key.Matches(msg, key.NewBinding(key.WithKeys("q"))):
				return m, tea.Quit
			}

		case "add":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				title := strings.TrimSpace(m.textInput.Value())
				if title == "" {
					m = m.setMessage("Title cannot be empty", "error")
					return m, nil
				}
				m.currentField = "title"
				m.state = "add_desc"
				m.descInput.Reset()
				m.descInput.Focus()
				return m, textinput.Blink

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "list"
				m.textInput.Reset()
				return m, nil
			}

		case "add_desc":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m.state = "add_category"
				m.categoryInput.Reset()
				m.categoryInput.Focus()
				return m, textinput.Blink

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "add"
				return m, nil
			}

		case "add_category":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m.state = "add_priority"
				return m, nil

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "add_desc"
				return m, nil
			}

		case "add_priority":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
				m.priority = "low"
				m.state = "add_due"
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
				m.priority = "medium"
				m.state = "add_due"
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
				m.priority = "high"
				m.state = "add_due"
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m = m.finishAddTodo()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "add_category"
				return m, nil
			}

		case "add_due":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				m = m.finishAddTodo()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "add_priority"
				return m, nil
			}

		case "edit":
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				title := strings.TrimSpace(m.textInput.Value())
				if title == "" {
					m = m.setMessage("Title cannot be empty", "error")
					return m, nil
				}
				m = m.updateTodo(m.editingID, title)
				m.state = "list"
				m.textInput.Reset()
				return m, nil

			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = "list"
				m.textInput.Reset()
				return m, nil
			}
		}

	case messageMsg:
		m.message = msg.text
		m.messageType = msg.msgType
		return m, nil
	}

	// Update the appropriate component
	if m.state == "add" || m.state == "edit" {
		m.textInput, cmd = m.textInput.Update(msg)
	} else if m.state == "add_desc" {
		m.descInput, cmd = m.descInput.Update(msg)
	} else if m.state == "add_category" {
		m.categoryInput, cmd = m.categoryInput.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case "add":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("➕ Add New Todo"),
			m.textInput.View(),
			helpStyle.Render("Press Enter to continue, Esc to cancel"),
		)

	case "add_desc":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("📝 Add Description"),
			m.descInput.View(),
			helpStyle.Render("Press Enter to continue, Esc to go back"),
		)

	case "add_category":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("📁 Add Category"),
			m.categoryInput.View(),
			helpStyle.Render("Press Enter to continue, Esc to go back"),
		)

	case "add_priority":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("🎯 Select Priority"),
			"1. Low (Green)\n2. Medium (Yellow)\n3. High (Red)\n\nPress Enter to skip",
			helpStyle.Render("Press 1-3 to select priority, Enter to skip, Esc to go back"),
		)

	case "add_due":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("📅 Set Due Date"),
			"Enter due date (YYYY-MM-DD) or press Enter to skip",
			helpStyle.Render("Press Enter to skip, Esc to go back"),
		)

	case "edit":
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			titleStyle.Render("✏️ Edit Todo"),
			m.textInput.View(),
			helpStyle.Render("Press Enter to save, Esc to cancel"),
		)

	default:
		view := m.list.View()

		// Add message if any
		if m.message != "" {
			var style lipgloss.Style
			switch m.messageType {
			case "error":
				style = errorStyle
			case "success":
				style = successStyle
			case "info":
				style = infoStyle
			default:
				style = helpStyle
			}
			view = fmt.Sprintf("%s\n\n%s", view, style.Render(m.message))
		}

		// Add help text
		help := helpStyle.Render("Press 'a' to add, 'e' to edit, 'd' to delete, 'space' to toggle, 's' to sort, 'c' for categories, 'q' to quit")
		view = fmt.Sprintf("%s\n\n%s", view, help)

		return view
	}
}

// Todo operations
func (m model) finishAddTodo() model {
	title := strings.TrimSpace(m.textInput.Value())
	description := strings.TrimSpace(m.descInput.Value())
	category := strings.TrimSpace(m.categoryInput.Value())

	todo := Todo{
		ID:          m.nextID,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		Priority:    m.priority,
		Category:    category,
	}

	// Parse due date if provided
	if m.dueDate != "" {
		if due, err := time.Parse("2006-01-02", m.dueDate); err == nil {
			todo.DueDate = &due
		}
	}

	m.todos = append(m.todos, todo)
	m.nextID++
	m.saveTodos()
	m.updateList()
	m = m.setMessage(fmt.Sprintf("Added: %s", title), "success")
	m.state = "list"

	// Reset form
	m.textInput.Reset()
	m.descInput.Reset()
	m.categoryInput.Reset()
	m.priority = "low"
	m.dueDate = ""

	return m
}

func (m model) updateTodo(id int, title string) model {
	for i, todo := range m.todos {
		if todo.ID == id {
			m.todos[i].Title = title
			break
		}
	}

	m.saveTodos()
	m.updateList()
	m = m.setMessage(fmt.Sprintf("Updated: %s", title), "success")

	return m
}

func (m model) deleteTodo(id int) model {
	for i, todo := range m.todos {
		if todo.ID == id {
			title := todo.Title
			m.todos = append(m.todos[:i], m.todos[i+1:]...)
			m.saveTodos()
			m.updateList()
			m = m.setMessage(fmt.Sprintf("Deleted: %s", title), "success")
			break
		}
	}

	return m
}

func (m model) toggleTodo(id int) model {
	for i, todo := range m.todos {
		if todo.ID == id {
			m.todos[i].Completed = !m.todos[i].Completed
			status := "completed"
			if !m.todos[i].Completed {
				status = "pending"
			}
			m.saveTodos()
			m.updateList()
			m = m.setMessage(fmt.Sprintf("Marked as %s: %s", status, todo.Title), "success")
			break
		}
	}

	return m
}

func (m model) sortTodos() model {
	sort.Slice(m.todos, func(i, j int) bool {
		// First by completion status (incomplete first)
		if m.todos[i].Completed != m.todos[j].Completed {
			return !m.todos[i].Completed
		}

		// Then by priority (high to low)
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[m.todos[i].Priority] != priorityOrder[m.todos[j].Priority] {
			return priorityOrder[m.todos[i].Priority] > priorityOrder[m.todos[j].Priority]
		}

		// Then by due date (earliest first)
		if m.todos[i].DueDate != nil && m.todos[j].DueDate != nil {
			return m.todos[i].DueDate.Before(*m.todos[j].DueDate)
		}
		if m.todos[i].DueDate != nil {
			return true
		}

		// Finally by creation date (newest first)
		return m.todos[i].CreatedAt.After(m.todos[j].CreatedAt)
	})

	m.saveTodos()
	m.updateList()
	m = m.setMessage("Todos sorted by priority and due date", "info")

	return m
}

func (m model) showCategories() model {
	categories := make(map[string]int)
	for _, todo := range m.todos {
		if todo.Category != "" {
			categories[todo.Category]++
		}
	}

	if len(categories) == 0 {
		m = m.setMessage("No categories found", "info")
		return m
	}

	var categoryList []string
	for cat, count := range categories {
		categoryList = append(categoryList, fmt.Sprintf("%s: %d todos", cat, count))
	}

	m = m.setMessage(fmt.Sprintf("Categories: %s", strings.Join(categoryList, ", ")), "info")
	return m
}

func (m model) updateList() {
	items := make([]list.Item, len(m.todos))
	for i, todo := range m.todos {
		items[i] = todoItem{todo: todo}
	}
	m.list.SetItems(items)
}

func (m model) setMessage(text, msgType string) model {
	m.message = text
	m.messageType = msgType
	return m
}

// File operations
func loadTodos() ([]Todo, int) {
	data, err := os.ReadFile(storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo{}, 1
		}
		return []Todo{}, 1
	}

	var todoList TodoList
	err = json.Unmarshal(data, &todoList)
	if err != nil {
		return []Todo{}, 1
	}

	return todoList.Todos, todoList.NextID
}

func (m model) saveTodos() {
	todoList := TodoList{
		Todos:  m.todos,
		NextID: m.nextID,
	}

	data, err := json.MarshalIndent(todoList, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(storageFile, data, 0644)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
