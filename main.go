package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

func (p Priority) String() string {
	switch p {
	case Low:
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	default:
		return "Unknown"
	}
}

type ToDo struct {
	ID          int
	Title       string
	CreatedAt   time.Time
	CompletedAt time.Time `json:",omitempty"`
	DueDate     time.Time `json:",omitempty"`
	Priority    Priority  `json:",omitempty"`
	Completed   bool
}

var todos []ToDo

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <add|delete|list|complete>")
		return
	}

	const filename = "todos.json"

	initialise(filename)

	switch os.Args[1] {
	case "add":
		addToDo()
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go delete <task_id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}
		deleteToDo(id)
	case "list":
		listToDo()
	case "complete":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go complete <task_id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}
		completeToDo(id)
	}

	saveToDoToJson(filename)
}

func initialise(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		saveToDoToJson(filename)
	}
	loadToDoFromJson(filename)
}

func loadToDoFromJson(filename string) ([]ToDo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func saveToDoToJson(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(todos); err != nil {
		return err
	}

	return nil
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func findUniqueID() int {
	uniqueID := 0
	for _, todo := range todos {
		if todo.ID > uniqueID {
			uniqueID = todo.ID
		}
	}
	return uniqueID + 1
}

func addToDo() {
	fmt.Println("Add")
	fmt.Println("Enter the task:")
	title := readLine()

	fmt.Println("Enter the due date (YYYY-MM-DD):")
	dueDate := readLine()

	parsedDueDate, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		fmt.Println("Invalid due date, defaulting to no due date")
		parsedDueDate = time.Time{}
	}
	if parsedDueDate.Before(time.Now().AddDate(0, 0, -1)) {
		fmt.Println("Due date must be in the future, defaulting to no due date")
		parsedDueDate = time.Time{}
	}

	fmt.Println("Enter the priority (1-3) (1: Low, 2: Medium, 3: High) (default: 1)")
	priorityStr := readLine()
	priorityNum, _ := strconv.Atoi(priorityStr)
	priorityNum--

	if priorityNum < 0 || priorityNum > 2 {
		fmt.Println("Invalid priority, defaulting to Low")
		priorityNum = 0
	}

	todo := ToDo{
		ID:        findUniqueID(),
		Title:     title,
		CreatedAt: time.Now(),
		DueDate:   parsedDueDate,
		Priority:  Priority(priorityNum),
		Completed: false,
	}
	todos = append(todos, todo)
	fmt.Println("Task added successfully")
}

func deleteToDo(id int) {
	for i, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			fmt.Println("Task deleted successfully")
			return
		}
	}
	fmt.Println("Task not found")
}

func listToDo() {
	fmt.Println("--------------------------------")
	fmt.Println("Your ToDo List:")
	if len(todos) == 0 {
		fmt.Println("No tasks found")
		fmt.Println("--------------------------------")
		return
	}
	for _, todo := range todos {
		if !todo.Completed {
			fmt.Printf("%d. %s\n", todo.ID, todo.Title)
			indent := strings.Repeat(" ", len(strconv.Itoa(todo.ID))+2)
			if !todo.DueDate.IsZero() {
				fmt.Printf("%sDue: %s\n", indent, todo.DueDate.Format("2006-01-02"))
			} else {
				fmt.Printf("%sDue: %s\n", indent, "No due date")
			}
			fmt.Printf("%sPriority: %s\n", indent, todo.Priority)
		}
	}
	fmt.Println("--------------------------------")

	fmt.Println("Completed tasks:")
	hasCompleted := false
	for _, todo := range todos {
		if todo.Completed {
			hasCompleted = true
			fmt.Printf("%d. %s\n", todo.ID, todo.Title)
			indent := strings.Repeat(" ", len(strconv.Itoa(todo.ID))+2)
			fmt.Printf("%sCompleted: %s\n", indent, todo.CompletedAt.Format("2006-01-02"))
		}
	}
	if !hasCompleted {
		fmt.Println("No completed tasks")
	}

	fmt.Println("--------------------------------")
}

func completeToDo(id int) {
	for i, todo := range todos {
		if todo.ID == id {
			todos[i].Completed = true
			todos[i].CompletedAt = time.Now()
			fmt.Println("Task completed successfully")
			return
		}
	}
	fmt.Println("Task not found")
}
