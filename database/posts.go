package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/glitchd/glitchd-server/graph/model"
	"github.com/google/uuid"
)

func (db *BUN) CreatePost(input model.NewPostInput) (bool, error) {

	id := uuid.New().String()

	res, err := db.client.NewRaw(
		"INSERT INTO posts (id, author, message, media, media_type, reply_to) VALUES (?, ?, ?, ?, ?, ?)",
		id, input.Author, input.Message, input.Media, input.MediaType, input.ReplyTo,
	).Exec(context.Background())

	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		fmt.Println("RowsAffected when creating post: ", err)
		return false, nil
	}

	if rows > 0 {
		return true, nil
	}

	return true, nil
}

func (db *BUN) GetPostByID(post_id string) (*model.Post, error) {
	var post model.Post

	err := db.client.NewRaw("SELECT * from posts WHERE id = ?", post_id).Scan(context.Background(), &post)

	if err != nil {
		return nil, err
	}

	user, _ := db.GetUser(post.Author)
	likes, _ := db.GetPostLikes(post.ID)
	post.User = user
	post.Likes = likes

	return &post, nil
}

func (db *BUN) GetPosts(first int, after string) (*model.PostsResult, error) {
	var posts []*model.Post
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
			"SELECT * FROM posts WHERE reply_to='' ORDER BY created_at DESC LIMIT ?",
			first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM posts WHERE created_at < ? AND reply_to='' ORDER BY created_at DESC LIMIT ?",
			t, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	}

	var result *model.PostsResult
	var edges []*model.PostsEdge
	var pageInfo *model.PageInfo

	if len(posts) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.PostsResult{
			PageInfo: pageInfo,
			Edges:    []*model.PostsEdge{},
		}
		return result, nil
	}

	for _, v := range posts {

		user, _ := db.GetUser(v.Author)
		likes, _ := db.GetPostLikes(v.ID)
		v.User = user
		v.Likes = likes

		edges = append(edges, &model.PostsEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(posts[len(posts)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&posts).Where("created_at < ?", posts[len(posts)-1].CreatedAt).Count(context.Background())

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

	result = &model.PostsResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) GetUserPosts(channel_id string, first int, after string) (*model.PostsResult, error) {
	var posts []*model.Post
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
			"SELECT * FROM posts WHERE author = ? AND reply_to='' ORDER BY created_at DESC LIMIT ?",
			channel_id, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM posts WHERE author = ? AND reply_to='' AND created_at < ? ORDER BY created_at DESC LIMIT ?",
			channel_id, t, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	}

	var result *model.PostsResult
	var edges []*model.PostsEdge
	var pageInfo *model.PageInfo

	if len(posts) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.PostsResult{
			PageInfo: pageInfo,
			Edges:    []*model.PostsEdge{},
		}
		return result, nil
	}

	for _, v := range posts {
		user, _ := db.GetUser(v.Author)
		likes, _ := db.GetPostLikes(v.ID)
		v.User = user
		v.Likes = likes
		edges = append(edges, &model.PostsEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(posts[len(posts)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&posts).Where("author = ?", channel_id).Where("created_at < ?", posts[len(posts)-1].CreatedAt).Count(context.Background())

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

	result = &model.PostsResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) GetPostsByQuery(query string, first int, after string) (*model.PostsResult, error) {
	var posts []*model.Post
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
			"SELECT * FROM posts WHERE LOWER(message) LIKE LOWER(?) ORDER BY created_at DESC LIMIT ?",
			"%"+query+"%", first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM posts WHERE LOWER(message) LIKE LOWER(?) AND created_at < ? ORDER BY created_at DESC LIMIT ?",
			"%"+query+"%", t, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	}

	var result *model.PostsResult
	var edges []*model.PostsEdge
	var pageInfo *model.PageInfo

	if len(posts) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.PostsResult{
			PageInfo: pageInfo,
			Edges:    []*model.PostsEdge{},
		}
		return result, nil
	}

	for _, v := range posts {
		user, _ := db.GetUser(v.Author)
		likes, _ := db.GetPostLikes(v.ID)
		v.User = user
		v.Likes = likes
		edges = append(edges, &model.PostsEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(posts[len(posts)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&posts).Where("LOWER(message) LIKE LOWER(?)", "%"+query+"%").Where("created_at < ?", posts[len(posts)-1].CreatedAt).Count(context.Background())

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

	result = &model.PostsResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) GetPostReplies(post_id string, first int, after string) (*model.PostsResult, error) {
	var posts []*model.Post
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
			"SELECT * FROM posts WHERE reply_to = ? ORDER BY created_at DESC LIMIT ?",
			post_id, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	} else {
		err := db.client.NewRaw(
			"SELECT * FROM posts WHERE reply_id = ? AND created_at < ? ORDER BY created_at DESC LIMIT ?",
			post_id, t, first,
		).Scan(context.Background(), &posts)
		if err != nil {
			fmt.Println("Could not fetch posts: ", err)
			return nil, err
		}
	}

	var result *model.PostsResult
	var edges []*model.PostsEdge
	var pageInfo *model.PageInfo

	if len(posts) == 0 {
		pageInfo = &model.PageInfo{
			EndCursor:   "",
			HasNextPage: false,
		}

		result = &model.PostsResult{
			PageInfo: pageInfo,
			Edges:    []*model.PostsEdge{},
		}
		return result, nil
	}

	for _, v := range posts {
		user, _ := db.GetUser(v.Author)
		likes, _ := db.GetPostLikes(v.ID)
		v.User = user
		v.Likes = likes
		edges = append(edges, &model.PostsEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(v.CreatedAt.String())),
			Node:   v,
		})
	}

	var endCursor = base64.StdEncoding.EncodeToString([]byte(posts[len(posts)-1].CreatedAt.String()))
	var hasNextPage bool

	count, err := db.client.NewSelect().Model(&posts).Where("reply_to = ?", post_id).Where("created_at < ?", posts[len(posts)-1].CreatedAt).Count(context.Background())

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

	result = &model.PostsResult{
		PageInfo: pageInfo,
		Edges:    edges,
	}

	return result, nil
}

func (db *BUN) CountPostReplies(post_id string) (int, error) {
	var posts model.Post
	count, err := db.client.NewSelect().Model(&posts).Where("reply_to = ?", post_id).Count(context.Background())

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *BUN) DeletePost(post_id string) (bool, error) {
	res, err := db.client.NewRaw("DELETE FROM posts where id = ?", post_id).Exec(context.Background())

	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()

	if err != nil {
		return false, err
	}

	if affected > 0 {
		return true, nil
	}

	return false, nil
}
