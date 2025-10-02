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
	Rooms          sync.Map
	Viewers        sync.Map
	ChannelViewers sync.Map
	Activity       sync.Map
	Job            sync.Map
	Post           sync.Map
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

type VideoResolver struct {
	VideoID        string
	VideoViewers   *int
	VideoObservers map[string]chan *int
	mu             sync.Mutex
}

type VideoObserver struct {
	VideoID string
	Count   chan int
}

type VideoPage struct {
	VideoID   string
	Count     int
	Observers sync.Map
}

type ChannelObserver struct {
	ChannelID string
	Count     chan int
}

type ChannelPage struct {
	ChannelID string
	Count     int
	Observers sync.Map
}

type JobObserver struct {
	JobID  string
	Status chan string
}

type JobPage struct {
	JobID     string
	Status    string
	Observers sync.Map
}

type ActivityResolver struct {
	ChannelID         string
	Activity          *model.Activity
	ActivityObservers sync.Map
}

type ActivityObserver struct {
	ChannelID string
	Activity  chan *model.Activity
}

type ActivityPage struct {
	ChannelID string
	Activity  *model.Activity
	Observers sync.Map
}

type PostsResolver struct {
	ChannelID      string
	Post           *model.Post
	PostsObservers sync.Map
}

type PostObserver struct {
	ChannelID string
	Post      chan *model.Post
}

type PostPage struct {
	ChannelID string
	Post      *model.Post
	Observers sync.Map
}

func (r *Resolver) getRoom(channelID string) *Chatroom {
	room, _ := r.Rooms.LoadOrStore(channelID, &Chatroom{
		ChannelID: channelID,
		Observers: sync.Map{},
	})
	return room.(*Chatroom)
}

func (r *Resolver) getVideoViewers(videoID string) *VideoPage {
	page, _ := r.Viewers.LoadOrStore(videoID, &VideoPage{
		VideoID: videoID,
		Count:   0,
	})

	return page.(*VideoPage)
}

func (r *Resolver) getChannelViewers(channelID string) *ChannelPage {
	page, _ := r.ChannelViewers.LoadOrStore(channelID, &ChannelPage{
		ChannelID: channelID,
		Count:     0,
	})

	return page.(*ChannelPage)
}

func (r *Resolver) getJobStatus(jobID string) *JobPage {
	page, _ := r.Job.LoadOrStore(jobID, &JobPage{
		JobID:     jobID,
		Observers: sync.Map{},
	})

	return page.(*JobPage)
}

func (r *Resolver) getChannelActivity(channelID string) *ActivityPage {
	page, _ := r.Activity.LoadOrStore(channelID, &ActivityPage{
		ChannelID: channelID,
		Observers: sync.Map{},
	})

	return page.(*ActivityPage)
}

func (r *Resolver) getPosts(channelID string) *PostPage {
	page, _ := r.Post.LoadOrStore(channelID, &PostPage{
		ChannelID: channelID,
		Observers: sync.Map{},
	})

	return page.(*PostPage)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
