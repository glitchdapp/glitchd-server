package database

import (
	"context"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) CreateChannel(user_id string, input model.ChannelInput) (bool, error) {
	id := uuid.New().String()
	now := time.Now()

	res, err := db.client.NewRaw(
		"INSERT INTO channels (id, user_id, title, notification, category, streamkey, playback_id, tags, is_branded, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT (user_id) DO UPDATE SET title=EXCLUDED.title, notification=EXCLUDED.notification, tags=EXCLUDED.tags",
		id, input.BroadcasterID, input.Title, input.Notification, input.Category, input.Streamkey, input.PlaybackID, input.Tags, input.IsBranded, now,
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

func (db *BUN) UpdateStreamkey(user_id string, streamkey string, playback_id string) (bool, error) {
	res, err := db.client.NewRaw(
		"UPDATE channels SET streamkey = ?, playback_id = ? WHERE user_id = ?",
		streamkey, user_id,
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

func (db *BUN) GetChannelInfo(user_id string) (*model.Channel, error) {
	var channel model.Channel
	err := db.client.NewRaw("SELECT * FROM channels WHERE user_id = ?", user_id).Scan(context.Background(), &channel)

	if err != nil {
		return nil, err
	}

	user, _ := db.GetUser(user_id)
	channel.Broadcaster = user

	return &channel, nil
}
