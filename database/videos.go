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

func (db *BUN) CreateVideo(input model.NewVideo) (*model.Video, error) {
	var video model.Video

	video.ID = uuid.New().String()
	video.Title = input.Title
	video.ChannelID = input.ChannelID
	video.JobID = input.JobID
	video.CreatedAt = time.Now()
	video.UpdatedAt = time.Now()

	_, err := db.client.NewInsert().Model(&video).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when creating video: ", err)
		return nil, err
	}

	return &video, nil
}

func (db *BUN) CreateVideoView(input model.NewVideoView) (int, error) {

	var now = time.Now()
	id := uuid.New().String()

	_, err := db.client.NewRaw(
		"INSERT INTO ? (id, channel_id, video_id, user_id, created_at) VALUES (?, ?, ?, ?, ?)",
		bun.Ident("video_views"), id, input.ChannelID, input.VideoID, input.UserID, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when inserting video view: ", err)
		return 0, err
	}

	count, _ := db.GetVideoViews(input.VideoID)

	return count, nil
}

func (db *BUN) GetVideoViews(video_id string) (int, error) {
	var video_views model.VideoView
	count, err := db.client.NewSelect().Model(&video_views).Where("video_id = ?", video_id).Count(context.Background())
	if err != nil {
		fmt.Println("Could not get video views: ", err)
		return 0, err
	}

	return count, nil
}

func (db *BUN) GetChannelViews(channel_id string) (int, error) {
	var video_views model.VideoView
	count, err := db.client.NewSelect().Model(&video_views).Where("channel_id = ?", channel_id).Count(context.Background())
	if err != nil {
		fmt.Println("Could not get channel views: ", err)
		return 0, err
	}

	return count, nil
}

func (db *BUN) CountChannelVideos(channel_id string) (int, error) {
	var videos model.Video
	count, err := db.client.NewSelect().Model(&videos).Where("channel_id = ?", channel_id).Count(context.Background())
	if err != nil {
		fmt.Println("Could not get channel views: ", err)
		return 0, err
	}

	return count, nil
}

func (db *BUN) UpdateVideo(id string, input model.UpdateVideo) (*model.Video, error) {
	var video model.Video

	row, err := db.client.NewUpdate().Model(&video).Set("title = ?", input.Title).Set("caption = ?", input.Caption).Set("media = ?", input.Media).Set("tier = ?", input.Tier).Set("is_visible = ?", input.IsVisible).Set("thumbnail = ?", input.Thumbnail).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not update video: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in update video: ", err)
		return nil, err
	}

	if rows > 0 {
		return &video, nil
	}

	return nil, nil
}

func (db *BUN) DeleteVideo(id string) (*model.Video, error) {
	var video model.Video

	row, err := db.client.NewDelete().Model(&video).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not delete video: ", err)
		return nil, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in delete video: ", err)
		return nil, err
	}

	if rows > 0 {
		return &video, nil
	}

	return nil, nil
}

func (db *BUN) GetVideos(channelID string, first int, after string) (*model.VideosResult, error) {
	var videos []*model.Video
	if after == "" {
		err := db.client.NewSelect().Model(&videos).Where("channel_id = ?", channelID).Limit(first).Scan(context.Background())
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewSelect().Model(&videos).Where("channel_id = ?", channelID).Where("created_at > ?", after).Limit(first).Scan(context.Background())
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	}

	var result *model.VideosResult
	var edges []*model.VideosEdge
	var pageInfo *model.PageInfo

	for _, v := range videos {
		edges = append(edges, &model.VideosEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.ID)),
			Node:   v,
		})
	}

	var endCursor = videos[len(videos)-1].CreatedAt.String()
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&videos).Where("channel_id = ?", channelID).Where("created_at > ?", videos[len(videos)-1].CreatedAt).Count(context.Background())

	if err != nil {
		fmt.Println("Could not count remaining video rows for pagination: ", err)
		return nil, err
	}

	if count > 0 {
		hasNextPage = true
	}

	pageInfo = &model.PageInfo{
		EndCursor:   endCursor,
		HasNextPage: hasNextPage,
	}

	result = &model.VideosResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) GetVideoByID(id string) (*model.Video, error) {
	var video model.Video

	err := db.client.NewSelect().Model(&video).Where("id = ?", id).Scan(context.Background())

	if err != nil {
		fmt.Println("Could not fetch video: ", err)
		return nil, err
	}

	return &video, nil
}
