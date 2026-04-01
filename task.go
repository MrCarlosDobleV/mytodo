package main

type Task struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func nextTaskID(tasks []Task) int {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

func addTask(tasks []Task, text string) []Task {
	task := Task{
		ID:   nextTaskID(tasks),
		Text: text,
		Done: false,
	}
	return append(tasks, task)
}

func deleteTask(tasks []Task, index int) []Task {
	if index < 0 || index >= len(tasks) {
		return tasks
	}
	return append(tasks[:index], tasks[index+1:]...)
}

func toggleTask(tasks []Task, index int) []Task {
	if index < 0 || index >= len(tasks) {
		return tasks
	}
	tasks[index].Done = !tasks[index].Done
	return tasks
}

func editTask(tasks []Task, index int, newText string) []Task {
	if index < 0 || index >= len(tasks) {
		return tasks
	}
	tasks[index].Text = newText
	return tasks
}
