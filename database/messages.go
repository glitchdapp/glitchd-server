package database

import (
	"context"
	"fmt"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) CreateMessage(input *model.NewMessage) (*model.Message, error) {
	var message model.Message

	var now = time.Now()
	id := uuid.New().String()

	user, err := db.GetUser(input.SenderID)

	if err != nil {
		fmt.Println("user could not be fetched in messages: ", err)
	}

	message.ID = id
	message.Sender = user
	message.ChannelID = input.ChannelID
	message.IsSent = input.IsSent
	message.Message = input.Message
	message.MessageType = input.MessageType
	message.Amount = input.Amount
	message.ReplyParentMessageID = input.ReplyParentMessageID
	message.CreatedAt = now
	message.UpdatedAt = now

	res, err := db.client.NewRaw(
		"INSERT INTO ? (id, sender_id, channel_id, is_sent, message, message_type, amount, drop_code, drop_message, reply_parent_message_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bun.Ident("messages"), id, input.SenderID, input.ChannelID, input.IsSent, input.Message, input.MessageType, input.Amount, "", "", input.ReplyParentMessageID,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when following user: ", err)
		return nil, err
	}

	affected, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Error when detecting affected rows: ", err)
	}

	if affected > 0 {
		return &message, nil
	}

	return &message, nil
}

func (db *BUN) GetRecentMessages(channelID string) ([]*model.Message, error) {
	var messages []*model.Message
	err := db.client.NewRaw(
		"SELECT msg.* FROM (SELECT * FROM ? WHERE channel_id = ? ORDER BY created_at DESC LIMIT 20) msg ORDER BY created_at ASC",
		bun.Ident("messages"), channelID).Scan(context.Background(), &messages)

	for index, message := range messages {
		user, _ := db.GetUser(message.SenderID)
		messages[index].Sender = user
	}

	if err != nil {
		fmt.Println("Could not fetch post: ", err)
		return nil, err
	}
	return messages, nil
}

func (db *BUN) GetNewMessages(userID string) (<-chan []*model.Message, error) {
	msg := make(chan []*model.Message, 1)

	return msg, nil
}

func (db *BUN) GetChatIdentity(user_id string) (*model.ChatIdentity, error) {
	var chat_identity model.ChatIdentity

	err := db.client.NewSelect().Model(&chat_identity).Where("user_id = ?", user_id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not select chat identity. Something went wrong. ", err)
		return nil, err
	}

	return &chat_identity, nil
}

func (db *BUN) UpdateChatIdentity(user_id string, input model.ChatIdentityInput) (bool, error) {
	id := uuid.New().String()
	data := model.ChatIdentity{
		ID:     id,
		UserID: user_id,
		Color:  input.Color,
		Badge:  input.Badge,
	}

	res, err := db.client.NewInsert().
		Model(&data).
		On("CONFLICT (user_id) DO UPDATE").
		Exec(context.Background())
	if err != nil {
		fmt.Println("Could not update chat identity")
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not get affected rows in chat identity update")
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return true, nil
}

func (db *BUN) IsUserInChat(channel_id string, user_id string) (bool, error) {
	var users model.UsersInChat

	err := db.client.NewRaw(
		"SELECT * FROM chat_users WHERE channel_id = ? AND user_id = ?",
		channel_id, user_id,
	).Scan(context.Background(), &users)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *BUN) AddUserInChat(channel_id string, user_id string) (bool, error) {
	id := uuid.New().String()

	isIn, _ := db.IsUserInChat(channel_id, user_id)

	if isIn {
		return true, nil
	}

	res, err := db.client.NewRaw(
		"INSERT INTO chat_users (id, user_id, channel_id) VALUES (?, ?, ?)",
		id, user_id, channel_id,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not insert users in chat: ", err)
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not get affected rows in users in chat")
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return true, nil
}

func (db *BUN) DeleteUserInChat(channel_id string, user_id string) (bool, error) {

	res, err := db.client.NewRaw(
		"DELETE FROM chat_users WHERE channel_id = ? AND user_id = ?",
		channel_id, user_id,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Could not delete users in chat")
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("Could not get affected rows when deleting users in chat")
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return true, nil
}

func (db *BUN) GetUsersInChat(channel_id string) ([]*model.User, error) {
	var users []*model.User

	err := db.client.NewRaw(
		"SELECT u.* FROM users u JOIN (SELECT * FROM chat_users WHERE channel_id = ? LIMIT 50) as cu ON text(u.id) = cu.user_id",
		channel_id,
	).Scan(context.Background(), &users)

	if err != nil {
		fmt.Println("Could not fetch users in chat: ", err)
		return nil, err
	}

	return users, nil
}
