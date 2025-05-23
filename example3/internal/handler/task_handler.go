package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
)

type Task struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

func HandleTask(ctx context.Context, body []byte) error {
	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		return errors.New("invalid message format")
	}

	if task.UserID == "" {
		return errors.New("missing user_id")
	}

	log.Printf("Processed Task: %+v\n", task)
	return nil
}
