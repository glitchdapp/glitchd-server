package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) LikePost(post_id string, user_id string) (bool, error) {
	var id = uuid.New().String()
	var now = time.Now()

	res, err := db.client.NewRaw(
		"INSERT INTO likes (id, post_id, user_id, created_at) VALUES (?, ?, ?, ?)",
		id, post_id, user_id, now,
	).Exec(context.Background())
	if err != nil {
		fmt.Println("Error liking post: ", err)
		return false, err
	}

	affected, _ := res.RowsAffected()

	if affected > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) UnlikePost(post_id string, user_id string) (bool, error) {
	res, err := db.client.NewRaw(
		"DELETE FROM likes WHERE post_id = ? AND user_id = ?",
		post_id, user_id,
	).Exec(context.Background())
	if err != nil {
		fmt.Println("Error liking post: ", err)
		return false, err
	}

	affected, _ := res.RowsAffected()

	if affected > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) GetLikes(post_id string) (int, error) {
	var likes model.Like
	count, err := db.client.NewSelect().Model(&likes).Where("post_id = ?", post_id).Count(context.Background())

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *BUN) GetLikedByUser(post_id string, user_id string) (bool, error) {
	var likes model.Like

	count, err := db.client.NewSelect().Model(&likes).Where("post_id = ?", post_id, "user_id = ?", user_id).Count(context.Background())

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
