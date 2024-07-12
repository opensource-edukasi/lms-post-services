package posts

import (
	"context"
	"database/sql"
	postPb "lms-post-service/pb/posts"
	"log"
)

type PostService struct {
	Db  *sql.DB
	Log log.Logger
}

func (a *PostService) CreatePost(ctx context.Context, in *postPb.CreatePostRequest) (*postPb.Post, error) {
	var postRepo PostRepository
	var err error
	postRepo.Log = a.Log

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return &postRepo.pb, err
	}

	postRepo.pb = postPb.Post{
		SubjectClassId:   in.SubjectClassId,
		TopicSubjectId:   in.TopicSubjectId,
		Type:             in.Type,
		Title:            in.Title,
		Description:      in.Description,
		FileType:         in.FileType,
		Source:           in.Source,
		IsAllowToComment: in.IsAllowToComment,
		IsPublished:      in.IsPublished,
	}

	err = postRepo.CreatePost(ctx)
	if err != nil {
		return &postRepo.pb, err
	}
	postRepo.tx.Commit()
	return &postRepo.pb, nil
}
