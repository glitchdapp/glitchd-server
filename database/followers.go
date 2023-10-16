package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/prizmsol/prizmsol-server/graph/model"
)

func (db *BUN) getUser(id string) *model.User {
	var user model.User

	err := db.client.NewSelect().Model(&user).Where("id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Error found when fetching users for followers getUser(): ", err)
		return nil
	}

	return &user
}

func (db *BUN) NewFollow(input model.NewFollower) (*model.Follower, error) {
	var follower model.Follower
	var user model.User
	var target model.User

	id := uuid.New()

	// check to see if follower exists.
	r := db.client.NewSelect().Model(&follower).Where("user_id = ? AND target_id = ?", input.UserID, input.TargetID).Scan(context.Background())

	if r != nil {
		fmt.Println("Error found when fetching existing followers: ", r)
		return nil, r
	}

	if follower.ID != "" {
		user = *db.getUser(input.UserID)
		target = *db.getUser(input.TargetID)
		// insert into db.
		data := model.Follower{
			ID:       id.String(),
			UserID:   input.UserID,
			User:     &user,
			TargetID: input.TargetID,
			Target:   &target,
		}

		res, err := db.client.NewInsert().Model(&data).Exec(context.Background())

		if err != nil {
			fmt.Println("Could not add follower. ", err)
			return nil, err
		}

		row, err := res.RowsAffected()

		if err != nil {
			fmt.Println("Could not insert follower")
			return nil, err
		}

		if row > 0 {
			fmt.Println("Follow went through")
			return &follower, nil
		}
	}

	return nil, nil
}

func (db *BUN) GetFollowers(id string) ([]*model.Follower, error) {
	var follower []*model.Follower

	err := db.client.NewSelect().Model(&follower).Where("target_id = ?", follower).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not get followers, Error: ", err)
		return nil, err
	}

	return follower, nil
}
