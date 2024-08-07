package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (db *BUN) CreateVideo(input model.NewVideo) (bool, error) {
	var id = uuid.New().String()
	var now = time.Now()

	_, err := db.client.NewRaw(
		"INSERT INTO videos (id, title, channel_id, job_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, input.Title, input.ChannelID, input.JobID, now, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when creating video: ", err)
		return false, err
	}

	return true, nil
}
func (db *BUN) CreateChannelViewer(channelID string, userID string) (int, error) {

	var now = time.Now()
	id := uuid.New().String()

	_, err := db.client.NewRaw(
		"INSERT INTO ? (id, channel_id, user_id, created_at) VALUES (?, ?, ?, ?)",
		bun.Ident("channel_viewers"), id, channelID, userID, now,
	).Exec(context.Background())

	if err != nil {
		fmt.Println("Error found when inserting Channel view: ", err)
		return 0, err
	}

	count, _ := db.GetChannelViewers(channelID)

	return count, nil
}

func (db *BUN) DeleteChannelView(channel_id string, user_id string) (int, error) {
	var channel_view model.ChannelViewer

	row, err := db.client.NewDelete().Model(&channel_view).Where("channel_id = ? AND user_id = ?", channel_id, user_id).Returning("*").Exec(context.Background())
	count, _ := db.GetChannelViewers(channel_id)

	if err != nil {
		return count, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		return count, err
	}

	if rows > 0 {
		return count, nil
	}

	return count, nil

}

func (db *BUN) GetChannelViewers(channel_id string) (int, error) {
	var channel_viewer model.ChannelViewer
	count, err := db.client.NewSelect().Model(&channel_viewer).Where("channel_id = ?", channel_id).Count(context.Background())
	if err != nil {
		fmt.Println("Could not get channel views: ", err)
		return 0, err
	}

	return count, nil
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

func (db *BUN) deleteVideoViews(id string) (bool, error) {
	var video_view model.VideoView

	row, err := db.client.NewDelete().Model(&video_view).Where("video_id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not delete video views: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in delete video views: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil

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

func (db *BUN) UpdateVideo(id string, input model.UpdateVideo) (bool, error) {

	row, err := db.client.NewRaw(
		"UPDATE videos SET title = ?, caption = ?,  media = ?, tier = ?, is_visible = ?, thumbnail = ?, category = ? WHERE id = ?",
		input.Title, input.Caption, input.Media, input.Tier, input.IsVisible, input.Thumbnail, input.Category, id,
	).Exec(context.Background())
	if err != nil {
		fmt.Println("Could not update video: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in update video: ", err)
		return false, err
	}

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (db *BUN) DeleteVideo(id string) (bool, error) {
	var video model.Video

	row, err := db.client.NewDelete().Model(&video).Where("id = ?", id).Returning("*").Exec(context.Background())

	if err != nil {
		fmt.Println("Could not delete video: ", err)
		return false, err
	}

	rows, err := row.RowsAffected()

	if err != nil {
		fmt.Println("Could not fetch rows in delete video: ", err)
		return false, err
	}

	if rows > 0 {
		deleted, _ := db.deleteVideoViews(id)

		if deleted {
			fmt.Println("Deleted video views")
			return true, nil
		}

		return true, nil
	}

	return false, nil
}

func (db *BUN) GetVideos(channelID string, first int, after string) (*model.VideosResult, error) {
	var videos []*model.Video
	var decodedCursor string
	b, err := base64.StdEncoding.DecodeString(after)

	if err != nil {
		fmt.Println("Could not decode cursor: ", err)
		return nil, err
	}

	decodedCursor = string(b)
	t := strings.Trim(decodedCursor, " +0000")

	if after == "" {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE channel_id = ? ORDER BY created_at ASC LIMIT ?",
			channelID, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE channel_id = ? AND created_at > ? ORDER BY created_at ASC LIMIT ?",
			channelID, t, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	}

	var result *model.VideosResult
	var edges []*model.VideosEdge
	var pageInfo *model.PageInfo

	if len(videos) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.VideosResult{
			PageInfo: pageInfo,
			Edges:    []*model.VideosEdge{},
		}
		return result, nil
	}

	for _, v := range videos {
		v.Views, _ = db.GetVideoViews(v.ID)
		edges = append(edges, &model.VideosEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(videos[len(videos)-1].CreatedAt.String()))
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

func (db *BUN) GetAllVideos(first int, after string) (*model.VideosResult, error) {
	var videos []*model.Video
	var decodedCursor string
	b, err := base64.StdEncoding.DecodeString(after)

	if err != nil {
		fmt.Println("Could not decode cursor: ", err)
		return nil, err
	}

	decodedCursor = string(b)
	t := strings.Trim(decodedCursor, " +0000")

	if after == "" {
		err := db.client.NewRaw(
			"SELECT * FROM videos ORDER BY created_at ASC LIMIT ?",
			first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE created_at > ? ORDER BY created_at ASC LIMIT ?",
			t, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	}

	var result *model.VideosResult
	var edges []*model.VideosEdge
	var pageInfo *model.PageInfo

	if len(videos) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.VideosResult{
			PageInfo: pageInfo,
			Edges:    []*model.VideosEdge{},
		}
		return result, nil
	}

	for _, v := range videos {
		v.Views, _ = db.GetVideoViews(v.ID)
		edges = append(edges, &model.VideosEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(videos[len(videos)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&videos).Where("created_at > ?", videos[len(videos)-1].CreatedAt).Count(context.Background())

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

func (db *BUN) GetVideosByCategory(category string, first int, after string) (*model.VideosResult, error) {
	var videos []*model.Video
	var decodedCursor string
	b, err := base64.StdEncoding.DecodeString(after)

	if err != nil {
		fmt.Println("Could not decode cursor: ", err)
		return nil, err
	}

	decodedCursor = string(b)
	t := strings.Trim(decodedCursor, " +0000")

	if after == "" {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE category = ? ORDER BY created_at ASC LIMIT ?",
			category, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE channel_id = ? AND created_at > ? ORDER BY created_at ASC LIMIT ?",
			category, t, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	}

	var result *model.VideosResult
	var edges []*model.VideosEdge
	var pageInfo *model.PageInfo

	if len(videos) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.VideosResult{
			PageInfo: pageInfo,
			Edges:    []*model.VideosEdge{},
		}
		return result, nil
	}

	for _, v := range videos {
		v.Views, _ = db.GetVideoViews(v.ID)
		edges = append(edges, &model.VideosEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(videos[len(videos)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&videos).Where("category = ?", category).Where("created_at > ?", videos[len(videos)-1].CreatedAt).Count(context.Background())

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

	err := db.client.NewRaw("select * from videos where id = ?", id).Scan(context.Background(), &video)

	if err != nil {
		fmt.Println("Could not fetch video: ", err)
		return nil, err
	}

	return &video, nil
}

func (db *BUN) SearchVideos(query string, first int, after string) (*model.VideosResult, error) {
	var videos []*model.Video
	var decodedCursor string
	b, err := base64.StdEncoding.DecodeString(after)

	if err != nil {
		fmt.Println("Could not decode cursor: ", err)
		return nil, err
	}

	decodedCursor = string(b)
	t := strings.Trim(decodedCursor, " +0000")

	if after == "" {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE LOWER(title) LIKE LOWER(?) OR LOWER(caption) LIKE LOWER(?) ORDER BY created_at ASC LIMIT ?",
			"%"+query+"%", "%"+query+"%", first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch videos: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM videos WHERE LOWER(title) LIKE LOWER(?) OR LOWER(caption) LIKE LOWER(?) AND created_at > ? ORDER BY created_at ASC LIMIT ?",
			"%"+query+"%", "%"+query+"%", t, first,
		).Scan(context.Background(), &videos)
		if err != nil {
			fmt.Println("Could not fetch sesrched videos: ", err)
			return nil, err
		}
	}

	var result *model.VideosResult
	var edges []*model.VideosEdge
	var pageInfo *model.PageInfo

	if len(videos) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.VideosResult{
			PageInfo: pageInfo,
			Edges:    []*model.VideosEdge{},
		}
		return result, nil
	}

	for _, v := range videos {
		v.Views, _ = db.GetVideoViews(v.ID)
		edges = append(edges, &model.VideosEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(videos[len(videos)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&videos).Where("LOWER(title) LIKE LOWER(?) OR LOWER(caption) LIKE LOWER(?)", query, query).Where("created_at > ?", videos[len(videos)-1].CreatedAt).Count(context.Background())

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
