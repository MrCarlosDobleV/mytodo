package main

type Task struct {
	ID int `json:"id"` 
	Text string `json:"text"`
	Done bool `json:"done"`
}
