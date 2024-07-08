package database

import (
	"context"
	"encoding/base64"
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

func (db *BUN) GetFollowers(user_id string, first int, after string) (*model.FollowersResult, error) {
	var users []*model.User
	var followers []*model.Follower
	if after != "" {
		err := db.client.
			NewRaw(
				"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers WHERE created_at > ? limit ?) as f ON f.follower_id = text(u.id) WHERE f.user_id = ?",
				bun.Ident("users"), after, first, user_id).
			Scan(context.Background(), &users)
		if err != nil {
			fmt.Println("Could not get followers: ", err)
		}
	} else {
		err := db.client.
			NewRaw(
				"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers limit ?) as f ON f.follower_id = text(u.id) WHERE f.user_id = ?",
				bun.Ident("users"), first, user_id).
			Scan(context.Background(), &users)
		if err != nil {
			fmt.Println("Could not get followers: ", err)
		}
	}

	var result *model.FollowersResult
	var edges []*model.FollowersEdge
	var pageInfo *model.PageInfo

	if len(users) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.FollowersResult{
			PageInfo: pageInfo,
			Edges:    []*model.FollowersEdge{},
		}
		return result, nil
	}

	for _, u := range users {
		edges = append(edges, &model.FollowersEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(u.ID)),
			Node:   u,
		})
	}

	var endCursor = users[len(users)-1].CreatedAt.String()
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&followers).Where("user_id = ?", user_id).Where("created_at > ?", users[len(users)-1].CreatedAt).Count(context.Background())

	if err != nil {
		fmt.Println("Could not count remaining followers rows for pagination: ", err)
		return nil, err
	}

	if count > 0 {
		hasNextPage = true
	}

	pageInfo = &model.PageInfo{
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}

	result = &model.FollowersResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) GetFollowing(follower_id string, first int, after string) (*model.FollowersResult, error) {
	var users []*model.User
	var followers []*model.Follower
	if after != "" {
		err := db.client.
			NewRaw(
				"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers WHERE created_at > ? limit ?) as f ON f.user_id = text(u.id) WHERE f.follower_id = ?",
				bun.Ident("users"), after, first, follower_id).
			Scan(context.Background(), &users)
		if err != nil {
			fmt.Println("Could not get followers: ", err)
		}
	} else {
		err := db.client.
			NewRaw(
				"SELECT u.* FROM ? u JOIN (SELECT user_id, follower_id from followers limit ?) as f ON f.user_id = text(u.id) WHERE f.follower_id = ?",
				bun.Ident("users"), first, follower_id).
			Scan(context.Background(), &users)
		if err != nil {
			fmt.Println("Could not get followers: ", err)
		}
	}

	var result *model.FollowersResult
	var edges []*model.FollowersEdge
	var pageInfo *model.PageInfo

	if len(users) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.FollowersResult{
			PageInfo: pageInfo,
			Edges:    []*model.FollowersEdge{},
		}
		return result, nil
	}

	for _, u := range users {
		edges = append(edges, &model.FollowersEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(u.ID)),
			Node:   u,
		})
	}

	var endCursor = users[len(users)-1].CreatedAt.String()
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&followers).Where("follower_id = ?", follower_id).Where("created_at > ?", users[len(users)-1].CreatedAt).Count(context.Background())

	if err != nil {
		fmt.Println("Could not count remaining followers rows for pagination: ", err)
		return nil, err
	}

	if count > 0 {
		hasNextPage = true
	}

	pageInfo = &model.PageInfo{
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}

	result = &model.FollowersResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil

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
