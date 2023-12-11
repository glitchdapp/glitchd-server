package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) CreatePost(input model.NewPost) (*model.Post, error) {
	var post model.Post

	post.ID = uuid.New().String()
	post.Title = input.Title
	post.Caption = input.Caption
	post.Media = input.Media
	post.IsPremium = input.IsPremium
	post.IsVisible = input.IsVisible
	post.Thumbnail = input.Thumbnail
	post.Type = input.Type
	post.UserID = input.UserID
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	_, err := db.client.NewInsert().Model(&post).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when creating post: ", err)
		return nil, err
	}

	return &post, nil
}

func (db *BUN) UpdatePost(id string, input model.UpdatePost) (*model.Post, error) {
	var post model.Post

	row, err := db.client.NewUpdate().Model(&post).Set("title = ?", input.Title).Set("caption = ?", input.Caption).Set("media = ?", input.Media).Set("is_premium = ?", input.IsPremium).Set("is_visible = ?", input.IsVisible).Set("thumbnail = ?", input.Thumbnail).Set("type = ?", input.Type).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not update post: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in update post: ", err)
		return nil, err
	}

	if rows > 0 {
		return &post, nil
	}

	return nil, nil
}

func (db *BUN) DeletePost(id string) (*model.Post, error) {
	var post model.Post

	row, err := db.client.NewDelete().Model(&post).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not delete post: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in delete post: ", err)
		return nil, err
	}

	if rows > 0 {
		return &post, nil
	}

	return nil, nil
}

func (db *BUN) GetPostByID(id string) (*model.Post, error) {
	var post model.Post

	err := db.client.NewSelect().Model(&post).Where("id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch post: ", err)
		return nil, err
	}

	return &post, nil
}
