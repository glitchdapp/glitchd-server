package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) AddFollower(input model.FollowInput) (*model.Follower, error) {
	var follow model.Follower

	follow.ID = uuid.New().String()
	follow.UserID = input.UserID
	follow.FollowerID = input.FollowerID
	follow.CreatedAt = time.Now()

	_, err := db.client.NewInsert().Model(&follow).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when following user: ", err)
		return nil, err
	}

	return &follow, nil
}

func (db *BUN) RemoveFollower(user_id string, follower_id string) (bool, error) {
	var follower model.Follower
	row, err := db.client.NewDelete().Model(&follower).Where("user_id = ? AND follower_id = ?", user_id, follower_id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not unfollow user: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in followers: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetFollowers(user_id string) ([]*model.User, error) {
	var users []*model.User
	err := db.client.
		NewRaw(
			"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers) as f ON f.follower_id = text(u.id) WHERE f.user_id = ?",
			bun.Ident("users"), user_id).
		Scan(context.Background(), &users)

	if err != nil {
		fmt.Println("Could not get followers: ", err)
	}

	return users, nil
}

func (db *BUN) GetFollowing(follower_id string) ([]*model.User, error) {
	var users []*model.User
	err := db.client.
		NewRaw(
			"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers) as f ON f.user_id = text(u.id) WHERE f.follower_id = ?",
			bun.Ident("users"), follower_id).
		Scan(context.Background(), &users)

	if err != nil {
		fmt.Println("Could not get following: ", err)
	}

	return users, nil
}

func (db *BUN) CountFollowers(user_id string) (int, error) {
	var follower []*model.Follower
	count, err := db.client.NewSelect().Model(&follower).Where("user_id = ?", user_id).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Could not count followers: ", err)
		return 0, err
	}

	return count, nil
}

func (db *BUN) CountFollowing(follower_id string) (int, error) {
	var follower []*model.Follower

	count, err := db.client.NewSelect().Model(&follower).Where("follower_id = ?", follower_id).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Could not count users your following: ", err)
		return 0, err
	}

	return count, nil
}

func (db *BUN) IsFollowing(user_id string) (bool, error) {
	var follower model.Follower

	count, err := db.client.NewSelect().Model(&follower).Where("user_id = ?", user_id).ScanAndCount(context.Background())

	if err != nil {
		fmt.Println("Could not check if user is following: ", err)
		return false, nil
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
