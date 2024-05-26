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
	post.Media = input.Media
	post.Type = input.Type
	post.ChannelID = input.UserID
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

func (db *BUN) GetPosts() ([]*model.Post, error) {
	var posts []*model.Post

	err := db.client.NewSelect().Model(&posts).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch posts: ", err)
		return nil, err
	}

	return posts, nil
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

func (db *BUN) LikePost(user_id string, post_id string) (bool, error) {
	var like model.Like

	_, err := db.client.NewInsert().Model(&like).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when liking post: ", err)
		return false, err
	}

	return true, nil

}

func (db *BUN) UnlikePost(user_id string, post_id string) (bool, error) {
	var like model.Like

	row, err := db.client.NewDelete().Model(&like).Where("user_id = ?", user_id).Where("post_id = ?", post_id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not unlike post: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows to unlike post: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}
