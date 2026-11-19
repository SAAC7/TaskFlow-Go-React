package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// A “Task” represents an individual task in TaskFlow Web 📋
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// Temporary list for storing tasks in memory🗄️
var tasks = []Task{
	{ID: "1", Title: "Configure Devcontainer", Description: "Ensure that Go and Node work", Completed: true},
	{ID: "2", Title: "Creating APIs in Go", Description: "Implement Basic REST Endpoints", Completed: false},
}

// getTask manage the get to return  the task list in JSON format
func getTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)

}

// createTask manage the post to add a new task
func createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTask Task
	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tasks = append(tasks, newTask)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(newTask)
}

func main() {
	const port = 8080
	portStr := strconv.Itoa(port)

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getTask(w, r)
		case http.MethodPost:
			createTask(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Running Go Server on http://localhost:" + portStr)
	http.ListenAndServe(":"+portStr, nil)
}
