package posts

import (
	"context"
	"database/sql"
	"lms-post-service/internal/pkg/app"
	"lms-post-service/internal/pkg/db/redis"
	postPb "lms-post-service/pb/posts"
	"log"

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
	postRepo.Log = a.Log

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return &postRepo.pb, status.Errorf(codes.Internal, "Error beginning transaction: %v", err)
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
		UpdatedBy:        ctx.Value(app.Ctx("user_id")).(string),
		TypeId:           in.TypeId,
		StorageId:        in.StorageId,
	}

	if len(postRepo.pb.StorageId) > 0 {
		// validasi postRepo.pb.StorageId harus uuid yang valid
	}

	if len(postRepo.pb.Source) > 0 {
		// validasi postRepo.pb.Source harus url yang valid
	}

	if len(postRepo.pb.FileType) > 0 {
		// validasi postRepo.pb.FileType harus berisi data yang valid. type fiel yang diijinkan ada apa saja?
		// misal .jpg, .pdf,  karena disimpan di DB dalam bentuk CHAR(1) berarti ada mapping dari tipe file .jpg ke CHAR(1)
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
