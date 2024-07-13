package posts

import (
	"context"
	"database/sql"
	"lms-post-service/internal/pkg/app"
	"lms-post-service/internal/pkg/db/redis"
	postPb "lms-post-service/pb/posts"
	"log"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService struct {
	Db    *sql.DB
	Cache *redis.Cache
	Log   *log.Logger
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func isValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	return err == nil
}

func isYouTubeURL(u string) bool {
	re := regexp.MustCompile(`^(https?\:\/\/)?(www\.youtube\.com|youtu\.?be)\/.+$`)
	return re.MatchString(u)
}

func isValidFileType(fileType string) (string, bool) {
	validFileTypes := map[string]string{
		".jpg":  "J",
		".png":  "P",
		".pdf":  "F",
		".docx": "D",
		".mp4":  "M",
	}
	char, exists := validFileTypes[fileType]
	return char, exists
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
		if !isValidUUID(postRepo.pb.StorageId) {
			return nil, status.Errorf(codes.InvalidArgument, "StorageId harus UUID yang valid")
		}
	}

	if len(postRepo.pb.Source) > 0 {
		if !isValidURL(postRepo.pb.Source) {
			return nil, status.Errorf(codes.InvalidArgument, "Source harus URL yang valid")
		}

		if isYouTubeURL(postRepo.pb.Source) {
			postRepo.pb.FileType = "Y"
		}
	}

	if len(postRepo.pb.FileType) > 0 {
		if char, valid := isValidFileType(postRepo.pb.FileType); !valid {
			if postRepo.pb.FileType != "Y" {
				return nil, status.Errorf(codes.InvalidArgument, "FileType harus berisi data yang valid")
			}
		} else {
			postRepo.pb.FileType = char
		}
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
