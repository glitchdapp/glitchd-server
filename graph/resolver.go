package graph

import (
	"math/rand"
	"sync"

	"github.com/glitchd/glitchd-server/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Rooms sync.Map
}

type ChatResolver struct {
	ChatMessages *model.Message
	ChatChannel  string
	// All active subscriptions
	ChatObservers map[string]chan *model.Message
	mu            sync.Mutex
}

type Observer struct {
	ChannelID string
	Message   chan *model.Message
}

type Chatroom struct {
	ChannelID string
	Message   *model.Message
	Observers sync.Map
}

func (r *Resolver) getRoom(channelID string) *Chatroom {
	room, _ := r.Rooms.LoadOrStore(channelID, &Chatroom{
		ChannelID: channelID,
		Observers: sync.Map{},
	})
	return room.(*Chatroom)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
