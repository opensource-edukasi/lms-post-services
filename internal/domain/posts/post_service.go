package posts

import (
	"context"
	"database/sql"
	"lms-post-service/internal/pkg/db/redis"
	postPb "lms-post-service/pb/posts"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService struct {
	Db    *sql.DB
	Cache *redis.Cache
	Log   *log.Logger
}

func (a *PostService) CreatePost(ctx context.Context, in *postPb.CreatePostRequest) (*postPb.Post, error) {
	var postRepo PostRepository
	var err error
	postRepo.Log = *a.Log

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return &postRepo.pb, err
	}

	if ctx.Value("user_id") == nil {
		ctx = context.WithValue(ctx, "user_id", uuid.New().String())
	}
	userID := ctx.Value("user_id").(string)

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
		UpdatedBy:        userID,
		TypeId:           in.TypeId,
		StorageId:        in.StorageId,
	}

	if postRepo.pb.FileType == "" {
		if postRepo.pb.StorageId != "" && postRepo.pb.Source != "" {
			return nil, status.Errorf(codes.InvalidArgument, "StorageId dan Source harus kosong jika FileType kosong")
		} else if postRepo.pb.StorageId != "" {
			return nil, status.Errorf(codes.InvalidArgument, "FileType harus diisi jika StorageId diisi")
		} else if postRepo.pb.Source != "" {
			return nil, status.Errorf(codes.InvalidArgument, "FileType harus diisi jika Source diisi")
		}
	}

	if (postRepo.pb.Type != postPb.PostType_DISKUSI && postRepo.pb.Type != postPb.PostType_INFO) && postRepo.pb.TypeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "TypeId cannot be empty for this post type")
	}

	err = postRepo.CreatePost(ctx)
	if err != nil {
		return &postRepo.pb, err
	}
	postRepo.tx.Commit()
	return &postRepo.pb, nil
}
