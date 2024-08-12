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
		"INSERT INTO ? (id, channel_id, tier, name, description, cost, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (channel_id, tier) DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description, cost=EXCLUDED.cost, updated_at=EXCLUDED.updated_at",
		bun.Ident("membership_details"), id, input.ChannelID, input.Tier, input.Name, input.Description, input.Cost, now, now,
	).Exec(context.Background())

	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) CreateMembership(input model.NewMembership) (*model.Membership, error) {
	id := uuid.New().String()
	now := time.Now()

	fmt.Println("isActive: ", input.IsActive)

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, channel_id, user_id, gifter, is_gift, is_active, tier, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bun.Ident("memberships"), id, input.ChannelID, input.UserID, input.GifterID, input.IsGift, input.IsActive, input.Tier, now, now,
	).Exec(context.Background())

	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rows > 0 {
		membership, err := db.GetMembershipById(id)
		if err != nil {
			return nil, err
		}
		return membership, nil
	}

	return nil, nil
}

func (db *BUN) UpdateMembership(id string, input model.NewMembership) (bool, error) {
	now := time.Now()

	row, err := db.client.NewRaw(
		"UPDATE memberships SET channel_id = ?, user_id = ?, gifter = ?, is_gift = ?, is_active = ?, tier = ?, updated_at = ? WHERE id = ?",
		input.ChannelID, input.UserID, input.GifterID, input.IsGift, input.IsActive, input.Tier, now, id,
	).Exec(context.Background())
	if err != nil {
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) UpdateMembershipStatus(id string, is_active bool) (bool, error) {
	now := time.Now()

	row, err := db.client.NewRaw(
		"UPDATE memberships SET is_active = ?, updated_at = ? WHERE id = ?",
		is_active, now, id,
	).Exec(context.Background())
	if err != nil {
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetMembershipById(id string) (*model.Membership, error) {
	var result model.Membership

	err := db.client.NewRaw("SELECT * FROM memberships WHERE id = ?", id).Scan(context.Background(), &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (db *BUN) GetUserMembership(user_id string, channel_id string) ([]*model.Membership, error) {
	var result []*model.Membership

	err := db.client.NewRaw("SELECT * FROM memberships WHERE user_id = ? AND channel_id = ? AND is_active = 'true'", user_id, channel_id).Scan(context.Background(), &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *BUN) GetChannelMemberships(channel_id string) ([]*model.Membership, error) {
	var result []*model.Membership

	err := db.client.NewRaw("SELECT * FROM memberships WHERE channel_id = ? AND is_active = 'true'", channel_id).Scan(context.Background(), &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *BUN) GetChannelMembershipDetails(channelID string) ([]*model.MembershipDetails, error) {
	var result []*model.MembershipDetails

	err := db.client.NewRaw("SELECT * FROM membership_details WHERE channel_id = ?", channelID).Scan(context.Background(), &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *BUN) DeleteMembership(id string) (bool, error) {
	var membership model.Membership
	row, err := db.client.NewDelete().Model(&membership).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}
