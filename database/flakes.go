package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) AddFlakes(user_id string, amount int) (bool, error) {
	var flakes model.Flakes

	currentFlakes := db.flakesExist(user_id)

	flakes.ID = uuid.New().String()
	flakes.UserID = user_id
	flakes.Amount = amount
	flakes.CreatedAt = time.Now()

	var am int

	if currentFlakes != nil {
		am = currentFlakes.Amount + amount
	} else {
		am = amount
	}

	_, err := db.client.NewInsert().Model(&flakes).On("CONFLICT (user_id) DO UPDATE").Set("amount = ?", am).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when following user: ", err)
		return false, err
	}

	return true, nil
}

func (db *BUN) updateFlakes(user_id string, amount int) bool {
	var flakes model.Flakes
	res, err := db.client.NewUpdate().Model(&flakes).Set("amount = ?", amount).Where("user_id = ?", user_id).Exec(context.Background())
	if err != nil {
		return false
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return false
	}

	if rows > 0 {
		return true
	}

	return false
}

func (db *BUN) flakesExist(user_id string) *model.Flakes {
	var flakes model.Flakes

	count, err := db.client.NewSelect().Model(&flakes).Where("user_id = ?", user_id).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Could check if flakes exists: ", err)
		return nil
	}

	if count > 0 {
		return &flakes
	}

	return nil
}

func (db *BUN) SendFlakes(channel_id string, user_id string, amount int) (bool, error) {
	currentFlakes := db.flakesExist(user_id)

	if currentFlakes == nil || currentFlakes.Amount == 0 {
		return false, nil
	}

	id := uuid.New().String()
	res, err := db.client.NewRaw(
		"INSERT INTO channel_flakes (id, channel_id, sender_id, amount, created_at) VALUES (?, ?, ?, ?, ?)",
		id, channel_id, user_id, amount, time.Now(),
	).Exec(context.Background())

	if err != nil {
		return false, nil
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		db.updateFlakes(user_id, currentFlakes.Amount-amount)
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetFlakes(user_id string) (int, error) {
	var flakes model.Flakes

	err := db.client.NewSelect().Model(&flakes).Where("user_id = ?", user_id).Scan(context.Background())

	if err != nil {
		return 0, err
	}

	return flakes.Amount, nil
}

func (db *BUN) GetChannelFlakes(channel_id string) ([]*model.ChannelFlakes, error) {
	var channel_flakes []*model.ChannelFlakes

	err := db.client.NewSelect().Model(&channel_flakes).Where("channel_id = ? AND created_at >= date_trunc('month', current_date) ", channel_id).Scan(context.Background())

	if err != nil {
		return nil, err
	}

	return channel_flakes, nil
}
