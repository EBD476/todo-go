package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Todo represents a single todo item
type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TodoList represents a collection of todos
type TodoList struct {
	Todos  []Todo `json:"todos"`
	NextID int    `json:"next_id"`
}

// Storage file path
const storageFile = "todos.json"
const googleDriveFile = "todos-backup.json"

// Network configuration
const (
	defaultServerURL = "http://localhost:8080"
	apiEndpoint      = "/api/todos"
)

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	ServerURL string
	Username  string
	Password  string
}

// GoogleDriveConfig holds Google Drive configuration
type GoogleDriveConfig struct {
	CredentialsFile string
	TokenFile       string
}

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorItalic = "\033[3m"

	// Background colors
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"
	BgWhite  = "\033[47m"
)

// Utility functions for beautiful formatting
func printHeader() {
	fmt.Printf("%s%s╔══════════════════════════════════════════════════════════════════════════════╗%s\n", ColorCyan, ColorBold, ColorReset)
	fmt.Printf("%s%s║%s %s%sTODO CLI%s %s- A Beautiful Command-Line Todo Manager%s %s║%s\n", ColorCyan, ColorBold, ColorReset, ColorYellow, ColorBold, ColorReset, ColorDim, ColorCyan, ColorBold, ColorReset)
	fmt.Printf("%s%s╚══════════════════════════════════════════════════════════════════════════════╝%s\n", ColorCyan, ColorBold, ColorReset)
	fmt.Println()
}

func printSuccess(message string) {
	fmt.Printf("%s✅ %s%s%s\n", ColorGreen, ColorBold, message, ColorReset)
}

func printError(message string) {
	fmt.Printf("%s❌ %s%s%s\n", ColorRed, ColorBold, message, ColorReset)
}

func printWarning(message string) {
	fmt.Printf("%s⚠️  %s%s%s\n", ColorYellow, ColorBold, message, ColorReset)
}

func printInfo(message string) {
	fmt.Printf("%sℹ️  %s%s%s\n", ColorBlue, ColorBold, message, ColorReset)
}

func printProgress(message string) {
	fmt.Printf("%s🔄 %s%s%s\n", ColorCyan, ColorBold, message, ColorReset)
}

func centerText(text string, width int) string {
	textLen := utf8.RuneCountInString(text)
	if textLen >= width {
		return text
	}
	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-textLen-padding)
}

func printSeparator(char string, length int) {
	fmt.Printf("%s%s%s\n", ColorDim, strings.Repeat(char, length), ColorReset)
}

func printBoxedText(text string, color string) {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		if utf8.RuneCountInString(line) > maxWidth {
			maxWidth = utf8.RuneCountInString(line)
		}
	}

	width := maxWidth + 4
	fmt.Printf("%s┌%s┐%s\n", color, strings.Repeat("─", width-2), ColorReset)
	for _, line := range lines {
		padding := width - 2 - utf8.RuneCountInString(line)
		fmt.Printf("%s│ %s%s%s │%s\n", color, line, strings.Repeat(" ", padding), ColorReset)
	}
	fmt.Printf("%s└%s┘%s\n", color, strings.Repeat("─", width-2), ColorReset)
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]

	// Load existing todos
	todoList, err := loadTodos()
	if err != nil {
		fmt.Printf("Error loading todos: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "add", "a":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo add <title> [description]")
			return
		}
		title := os.Args[2]
		description := ""
		if len(os.Args) > 3 {
			description = strings.Join(os.Args[3:], " ")
		}
		addTodo(todoList, title, description)

	case "list", "l":
		listTodos(todoList)

	case "complete", "c":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo complete <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID. Please provide a number.")
			return
		}
		completeTodo(todoList, id)

	case "delete", "d":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo delete <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID. Please provide a number.")
			return
		}
		deleteTodo(todoList, id)

	case "edit", "e":
		if len(os.Args) < 4 {
			fmt.Println("Usage: todo edit <id> <new_title> [new_description]")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid ID. Please provide a number.")
			return
		}
		title := os.Args[3]
		description := ""
		if len(os.Args) > 4 {
			description = strings.Join(os.Args[4:], " ")
		}
		editTodo(todoList, id, title, description)

	case "save", "s":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo save <server_url> [username] [password]")
			return
		}
		serverURL := os.Args[2]
		username := ""
		password := ""
		if len(os.Args) > 3 {
			username = os.Args[3]
		}
		if len(os.Args) > 4 {
			password = os.Args[4]
		}
		saveToNetwork(todoList, serverURL, username, password)

	case "load", "ld":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo load <server_url> [username] [password]")
			return
		}
		serverURL := os.Args[2]
		username := ""
		password := ""
		if len(os.Args) > 3 {
			username = os.Args[3]
		}
		if len(os.Args) > 4 {
			password = os.Args[4]
		}
		loadFromNetwork(serverURL, username, password)

	case "sync":
		if len(os.Args) < 3 {
			fmt.Println("Usage: todo sync <server_url> [username] [password]")
			return
		}
		serverURL := os.Args[2]
		username := ""
		password := ""
		if len(os.Args) > 3 {
			username = os.Args[3]
		}
		if len(os.Args) > 4 {
			password = os.Args[4]
		}
		syncWithNetwork(todoList, serverURL, username, password)

	case "upload", "up":
		uploadToGoogleDrive(todoList)

	case "download", "down":
		downloadFromGoogleDrive()

	case "help", "h":
		showHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
	}
}

func showHelp() {
	printHeader()

	fmt.Printf("%s%s📋 USAGE%s\n", ColorYellow, ColorBold, ColorReset)
	fmt.Printf("  %stodo%s <command> [arguments]\n\n", ColorCyan, ColorReset)

	fmt.Printf("%s%s🎯 COMMANDS%s\n", ColorYellow, ColorBold, ColorReset)

	// Local operations
	fmt.Printf("  %s%s📝 Local Operations%s\n", ColorBlue, ColorBold, ColorReset)
	fmt.Printf("    %sadd, a%s     %s<title> [description]%s    %sAdd a new todo%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %slist, l%s    %s%s                     %sList all todos%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %scomplete, c%s %s<id>%s                  %sMark a todo as completed%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %sdelete, d%s   %s<id>%s                  %sDelete a todo%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %sedit, e%s     %s<id> <title> [desc]%s   %sEdit a todo%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Println()

	// Network operations
	fmt.Printf("  %s%s🌐 Network Operations%s\n", ColorPurple, ColorBold, ColorReset)
	fmt.Printf("    %ssave, s%s    %s<server_url> [user] [pass]%s %sSave todos to network%s\n", ColorCyan, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %sload, ld%s   %s<server_url> [user] [pass]%s %sLoad todos from network%s\n", ColorCyan, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %ssync%s       %s<server_url> [user] [pass]%s %sSync with network%s\n", ColorCyan, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Println()

	// Cloud operations
	fmt.Printf("  %s%s☁️  Cloud Operations%s\n", ColorGreen, ColorBold, ColorReset)
	fmt.Printf("    %supload, up%s  %s%s                     %sUpload todos to Google Drive%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Printf("    %sdownload, down%s %s%s                   %sDownload todos from Google Drive%s\n", ColorGreen, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Println()

	// Utility
	fmt.Printf("  %s%s🔧 Utility%s\n", ColorYellow, ColorBold, ColorReset)
	fmt.Printf("    %shelp, h%s    %s%s                     %sShow this help message%s\n", ColorWhite, ColorReset, ColorDim, ColorReset, ColorItalic, ColorReset)
	fmt.Println()

	fmt.Printf("%s%s💡 EXAMPLES%s\n", ColorYellow, ColorBold, ColorReset)
	examples := []string{
		"todo add \"Buy groceries\" \"Get milk and bread\"",
		"todo list",
		"todo complete 1",
		"todo save http://localhost:8080",
		"todo load http://api.example.com user123 pass456",
		"todo sync http://api.example.com",
		"todo upload",
		"todo download",
	}

	for _, example := range examples {
		fmt.Printf("  %s$ %s%s%s\n", ColorDim, ColorGreen, example, ColorReset)
	}
	fmt.Println()

	fmt.Printf("%s%s🎨 Features%s\n", ColorYellow, ColorBold, ColorReset)
	features := []string{
		"✨ Beautiful table formatting with colors",
		"💾 Local JSON storage",
		"🌐 Network synchronization",
		"📊 Progress indicators",
		"🎯 Intuitive command structure",
		"⚡ Fast and lightweight",
	}

	for _, feature := range features {
		fmt.Printf("  %s%s\n", ColorDim, feature)
	}
	fmt.Println()
}

func loadTodos() (*TodoList, error) {
	data, err := os.ReadFile(storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &TodoList{Todos: []Todo{}, NextID: 1}, nil
		}
		return nil, err
	}

	var todoList TodoList
	err = json.Unmarshal(data, &todoList)
	if err != nil {
		return nil, err
	}

	return &todoList, nil
}

func saveTodos(todoList *TodoList) error {
	data, err := json.MarshalIndent(todoList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(storageFile, data, 0644)
}

func addTodo(todoList *TodoList, title, description string) {
	todo := Todo{
		ID:          todoList.NextID,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	todoList.Todos = append(todoList.Todos, todo)
	todoList.NextID++

	err := saveTodos(todoList)
	if err != nil {
		fmt.Printf("Error saving todo: %v\n", err)
		return
	}

	printSuccess(fmt.Sprintf("Added todo #%d: %s", todo.ID, todo.Title))
}

func listTodos(todoList *TodoList) {
	if len(todoList.Todos) == 0 {
		fmt.Printf("%s%s📝 Your Todos%s\n", ColorYellow, ColorBold, ColorReset)
		fmt.Println()
		printBoxedText("No todos found. Add one with 'todo add <title>'", ColorYellow)
		fmt.Println()
		return
	}

	fmt.Printf("%s%s📝 Your Todos%s\n", ColorYellow, ColorBold, ColorReset)
	fmt.Println()

	// Table header
	fmt.Printf("%-3s %-2s %-30s %-50s %-15s %-20s\n", "ID", "ST", "TITLE", "DESCRIPTION", "STATUS", "DATE")
	fmt.Println(strings.Repeat("-", 125))

	for _, todo := range todoList.Todos {
		status := "⏳"
		statusText := "Pending"
		if todo.Completed {
			status = "✅"
			statusText = "Completed"
		}

		// Truncate title if too long
		title := todo.Title
		if utf8.RuneCountInString(title) > 30 {
			runes := []rune(title)
			title = string(runes[:27]) + "..."
		}

		// Format date
		timeStr := todo.CreatedAt.Format("2006-01-02 15:04")
		if todo.Completed && todo.CompletedAt != nil {
			timeStr = todo.CompletedAt.Format("2006-01-02 15:04")
		}

		fmt.Printf("%-3d %-2s %-30s %-50s %-15s %-20s\n",
			todo.ID, status, title, todo.Description, statusText, timeStr)

		// Show description if it exists
		// if todo.Description != "" {
		// 	desc := todo.Description
		// 	if len(desc) > 70 {
		// 		desc = desc[:67] + "..."
		// 	}
		// 	fmt.Printf("     %s\n", desc)
		// }
	}

	// Summary
	completed := 0
	for _, todo := range todoList.Todos {
		if todo.Completed {
			completed++
		}
	}

	fmt.Println()
	fmt.Printf("%s%s📊 Summary: %d total, %d completed, %d pending%s\n",
		ColorDim, ColorBold, len(todoList.Todos), completed, len(todoList.Todos)-completed, ColorReset)
	fmt.Println()
}

func completeTodo(todoList *TodoList, id int) {
	for i, todo := range todoList.Todos {
		if todo.ID == id {
			if todo.Completed {
				printWarning(fmt.Sprintf("Todo #%d is already completed", id))
				return
			}

			now := time.Now()
			todoList.Todos[i].Completed = true
			todoList.Todos[i].CompletedAt = &now

			err := saveTodos(todoList)
			if err != nil {
				fmt.Printf("Error saving todo: %v\n", err)
				return
			}

			printSuccess(fmt.Sprintf("Completed todo #%d: %s", id, todo.Title))
			return
		}
	}

	printError(fmt.Sprintf("Todo #%d not found", id))
}

func deleteTodo(todoList *TodoList, id int) {
	for i, todo := range todoList.Todos {
		if todo.ID == id {
			title := todo.Title
			todoList.Todos = append(todoList.Todos[:i], todoList.Todos[i+1:]...)

			err := saveTodos(todoList)
			if err != nil {
				fmt.Printf("Error saving todo: %v\n", err)
				return
			}

			printSuccess(fmt.Sprintf("Deleted todo #%d: %s", id, title))
			return
		}
	}

	printError(fmt.Sprintf("Todo #%d not found", id))
}

func editTodo(todoList *TodoList, id int, title, description string) {
	for i, todo := range todoList.Todos {
		if todo.ID == id {
			oldTitle := todo.Title
			todoList.Todos[i].Title = title
			todoList.Todos[i].Description = description

			err := saveTodos(todoList)
			if err != nil {
				fmt.Printf("Error saving todo: %v\n", err)
				return
			}

			printSuccess(fmt.Sprintf("Updated todo #%d: %s → %s", id, oldTitle, title))
			return
		}
	}

	printError(fmt.Sprintf("Todo #%d not found", id))
}

// Interactive mode for adding todos
func addTodoInteractive(todoList *TodoList) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter todo title: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	if title == "" {
		fmt.Println("Title cannot be empty.")
		return
	}

	fmt.Print("Enter description (optional): ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	addTodo(todoList, title, description)
}

// Network functions

func saveToNetwork(todoList *TodoList, serverURL, username, password string) {
	url := strings.TrimSuffix(serverURL, "/") + apiEndpoint

	// Prepare JSON data
	jsonData, err := json.Marshal(todoList)
	if err != nil {
		fmt.Printf("Error marshaling todos: %v\n", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		printSuccess(fmt.Sprintf("Successfully saved %d todos to %s", len(todoList.Todos), serverURL))
	} else {
		body, _ := io.ReadAll(resp.Body)
		printError(fmt.Sprintf("Error saving todos: %s (Status: %d)", string(body), resp.StatusCode))
	}
}

func loadFromNetwork(serverURL, username, password string) {
	url := strings.TrimSuffix(serverURL, "/") + apiEndpoint

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		printError(fmt.Sprintf("Error loading todos: %s (Status: %d)", string(body), resp.StatusCode))
		return
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Parse JSON
	var todoList TodoList
	err = json.Unmarshal(body, &todoList)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// Save to local file
	err = saveTodos(&todoList)
	if err != nil {
		fmt.Printf("Error saving to local file: %v\n", err)
		return
	}

	printSuccess(fmt.Sprintf("Successfully loaded %d todos from %s", len(todoList.Todos), serverURL))
}

func syncWithNetwork(todoList *TodoList, serverURL, username, password string) {
	printProgress("Syncing with network...")

	// First, try to load from network
	url := strings.TrimSuffix(serverURL, "/") + apiEndpoint

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error loading from network: %v\n", err)
		printInfo("Saving local todos to network instead...")
		saveToNetwork(todoList, serverURL, username, password)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error loading from network: %s (Status: %d)\n", string(body), resp.StatusCode)
		printInfo("Saving local todos to network instead...")
		saveToNetwork(todoList, serverURL, username, password)
		return
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Parse JSON
	var networkTodoList TodoList
	err = json.Unmarshal(body, &networkTodoList)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// Merge local and network todos
	mergedList := mergeTodoLists(todoList, &networkTodoList)

	// Save merged list locally
	err = saveTodos(mergedList)
	if err != nil {
		fmt.Printf("Error saving merged todos: %v\n", err)
		return
	}

	// Save merged list to network
	saveToNetwork(mergedList, serverURL, username, password)

	printSuccess(fmt.Sprintf("Successfully synced %d todos with %s", len(mergedList.Todos), serverURL))
}

func mergeTodoLists(local, network *TodoList) *TodoList {
	merged := &TodoList{
		Todos:  []Todo{},
		NextID: max(local.NextID, network.NextID),
	}

	// Create maps for easier lookup
	localMap := make(map[int]Todo)
	networkMap := make(map[int]Todo)

	for _, todo := range local.Todos {
		localMap[todo.ID] = todo
	}

	for _, todo := range network.Todos {
		networkMap[todo.ID] = todo
	}

	// Add all unique todos
	addedIDs := make(map[int]bool)

	// Add local todos
	for _, todo := range local.Todos {
		if !addedIDs[todo.ID] {
			merged.Todos = append(merged.Todos, todo)
			addedIDs[todo.ID] = true
		}
	}

	// Add network todos that don't exist locally
	for _, todo := range network.Todos {
		if !addedIDs[todo.ID] {
			merged.Todos = append(merged.Todos, todo)
			addedIDs[todo.ID] = true
		}
	}

	return merged
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Google Drive functions

func getGoogleDriveService() (*drive.Service, error) {
	ctx := context.Background()

	// Try to load credentials from file
	credentialsFile := "credentials.json"
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		printError("Google Drive credentials not found. Please place 'credentials.json' in the current directory.")
		printInfo("To get credentials:")
		printInfo("1. Go to Google Cloud Console")
		printInfo("2. Create a new project or select existing one")
		printInfo("3. Enable Google Drive API")
		printInfo("4. Create credentials (OAuth 2.0 Client ID)")
		printInfo("5. Download JSON and save as 'credentials.json'")
		return nil, fmt.Errorf("credentials file not found")
	}

	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return service, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		printError(fmt.Sprintf("Unable to read authorization code: %v", err))
		os.Exit(1)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		printError(fmt.Sprintf("Unable to retrieve token from web: %v", err))
		os.Exit(1)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		printError(fmt.Sprintf("Unable to cache oauth token: %v", err))
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func uploadToGoogleDrive(todoList *TodoList) {
	printProgress("Uploading todos to Google Drive...")

	service, err := getGoogleDriveService()
	if err != nil {
		printError(fmt.Sprintf("Failed to get Google Drive service: %v", err))
		return
	}

	// Convert todos to JSON
	jsonData, err := json.MarshalIndent(todoList, "", "  ")
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal todos: %v", err))
		return
	}

	// Check if file already exists
	fileID, err := findFileInDrive(service, googleDriveFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to search for existing file: %v", err))
		return
	}

	// Create file metadata
	fileMetadata := &drive.File{
		Name: googleDriveFile,
	}

	// Create media content
	mediaContent := bytes.NewReader(jsonData)

	var file *drive.File
	if fileID != "" {
		// Update existing file
		file, err = service.Files.Update(fileID, fileMetadata).Media(mediaContent).Do()
		if err != nil {
			printError(fmt.Sprintf("Failed to update file: %v", err))
			return
		}
		printSuccess(fmt.Sprintf("Updated file '%s' in Google Drive (ID: %s)", googleDriveFile, file.Id))
	} else {
		// Create new file
		file, err = service.Files.Create(fileMetadata).Media(mediaContent).Do()
		if err != nil {
			printError(fmt.Sprintf("Failed to create file: %v", err))
			return
		}
		printSuccess(fmt.Sprintf("Created file '%s' in Google Drive (ID: %s)", googleDriveFile, file.Id))
	}
}

func downloadFromGoogleDrive() {
	printProgress("Downloading todos from Google Drive...")

	service, err := getGoogleDriveService()
	if err != nil {
		printError(fmt.Sprintf("Failed to get Google Drive service: %v", err))
		return
	}

	// Find the file
	fileID, err := findFileInDrive(service, googleDriveFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to search for file: %v", err))
		return
	}

	if fileID == "" {
		printWarning(fmt.Sprintf("File '%s' not found in Google Drive", googleDriveFile))
		return
	}

	// Download the file
	resp, err := service.Files.Get(fileID).Download()
	if err != nil {
		printError(fmt.Sprintf("Failed to download file: %v", err))
		return
	}
	defer resp.Body.Close()

	// Read the content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError(fmt.Sprintf("Failed to read file content: %v", err))
		return
	}

	// Parse JSON
	var todoList TodoList
	err = json.Unmarshal(body, &todoList)
	if err != nil {
		printError(fmt.Sprintf("Failed to parse JSON: %v", err))
		return
	}

	// Save to local file
	err = saveTodos(&todoList)
	if err != nil {
		printError(fmt.Sprintf("Failed to save local file: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Downloaded and saved %d todos from Google Drive", len(todoList.Todos)))
}

func findFileInDrive(service *drive.Service, fileName string) (string, error) {
	// Search for the file
	r, err := service.Files.List().
		Q(fmt.Sprintf("name='%s'", fileName)).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return "", err
	}

	if len(r.Files) == 0 {
		return "", nil // File not found
	}

	return r.Files[0].Id, nil
}
