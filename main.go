package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type ToDo struct {
	ID          int
	Title       string
	CreatedAt   time.Time
	CompletedAt time.Time `json:",omitempty"`
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

func addToDo() {
	fmt.Println("Add")
	fmt.Println("Enter the task:")
	title := readLine()

	todo := ToDo{
		ID:        len(todos) + 1,
		Title:     title,
		CreatedAt: time.Now(),
		Completed: false,
	}
	todos = append(todos, todo)
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
		return
	}
	for _, todo := range todos {
		if !todo.Completed {
			fmt.Printf("%d. %s\n", todo.ID, todo.Title)
		}
	}
	fmt.Println("--------------------------------")

	fmt.Println("Completed tasks:")
	hasCompleted := false
	for _, todo := range todos {
		if todo.Completed {
			hasCompleted = true
			fmt.Printf("%d. %s\n", todo.ID, todo.Title)
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
