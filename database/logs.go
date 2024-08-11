package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (db *BUN) CreateLog(data string) (bool, error) {
	var id = uuid.New().String()
	var now = time.Now()

	_, err := db.client.NewRaw(
		"INSERT INTO logs (id, data, created_at) VALUES (?, ?, ?)",
		id, data, now,
	).Exec(context.Background())
	if err != nil {
		fmt.Println("Error creating log: ", err)
		return false, err
	}

	return true, nil
}
