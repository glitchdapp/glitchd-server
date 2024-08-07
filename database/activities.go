package database

import (
	"context"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) CreateActivity(sender_id string, target_id string, activity_type string, message string) (*model.Activity, error) {
	id := uuid.New().String()
	now := time.Now()

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, sender_id, target_id, type, message, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		bun.Ident("activities"), id, sender_id, target_id, activity_type, message, now,
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rows > 0 {
		activity, _ := db.GetActivity(id)
		return activity, nil
	}

	return nil, nil
}

func (db *BUN) GetActivity(id string) (*model.Activity, error) {
	var result model.Activity

	err := db.client.NewRaw("SELECT * FROM activities WHERE id = ?", id).Scan(context.Background(), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (db *BUN) GetRecentActivity(channelID string) ([]*model.Activity, error) {
	var activity []*model.Activity
	err := db.client.NewRaw(
		"SELECT msg.* FROM (SELECT * FROM ? WHERE target_id = ? ORDER BY created_at DESC LIMIT 40) msg ORDER BY created_at ASC",
		bun.Ident("activities"), channelID).Scan(context.Background(), &activity)

	for index, act := range activity {
		user, _ := db.GetUser(act.SenderID)
		target, _ := db.GetUser(act.TargetID)
		activity[index].Sender = user
		activity[index].Target = target
	}

	if err != nil {
		return nil, err
	}
	return activity, nil
}
