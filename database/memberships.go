package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) CreateMembershipDetails(input model.MembershipDetailsInput) (bool, error) {

	id := uuid.New().String()
	now := time.Now()

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, channel_id, tier, name, description, cost, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		bun.Ident("membership_details"), id, input.ChannelID, input.Tier, input.Name, input.Description, input.Cost, now, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not create membership details. Error: ", err)
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not retreive rows affected by membership details. ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}
