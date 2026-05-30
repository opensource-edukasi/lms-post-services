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
	postPb.UnimplementedPostsServer
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

func (a *PostService) UpdatePost(ctx context.Context, in *postPb.UpdatePostRequest) (*postPb.Post, error) {
	var postRepo PostRepository
	var err error
	postRepo.Log = a.Log

	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus diisi")
	}

	if !isValidUUID(in.Id) {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus UUID yang valid")
	}

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return &postRepo.pb, status.Errorf(codes.Internal, "Error beginning transaction: %v", err)
	}

	postRepo.pb = postPb.Post{
		Id:               in.Id,
		SubjectClassId:   in.SubjectClassId,
		TopicSubjectId:   in.TopicSubjectId,
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
			postRepo.tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "StorageId harus UUID yang valid")
		}
	}

	if len(postRepo.pb.Source) > 0 {
		if !isValidURL(postRepo.pb.Source) {
			postRepo.tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "Source harus URL yang valid")
		}

		if isYouTubeURL(postRepo.pb.Source) {
			postRepo.pb.FileType = "Y"
		}
	}

	if len(postRepo.pb.FileType) > 0 {
		if char, valid := isValidFileType(postRepo.pb.FileType); !valid {
			if postRepo.pb.FileType != "Y" {
				postRepo.tx.Rollback()
				return nil, status.Errorf(codes.InvalidArgument, "FileType harus berisi data yang valid")
			}
		} else {
			postRepo.pb.FileType = char
		}
	}

	if postRepo.pb.FileType == "" {
		if postRepo.pb.StorageId != "" && postRepo.pb.Source != "" {
			postRepo.tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "StorageId dan Source harus kosong jika FileType kosong")
		} else if postRepo.pb.StorageId != "" {
			postRepo.tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "FileType harus diisi jika StorageId diisi")
		} else if postRepo.pb.Source != "" {
			postRepo.tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "FileType harus diisi jika Source diisi")
		}
	}

	err = postRepo.UpdatePost(ctx)
	if err != nil {
		postRepo.tx.Rollback()
		return &postRepo.pb, err
	}
	postRepo.tx.Commit()
	return &postRepo.pb, nil
}

func (a *PostService) GetPost(ctx context.Context, in *postPb.Id) (*postPb.Post, error) {
	var postRepo PostRepository
	postRepo.Log = a.Log
	postRepo.db = a.Db

	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus diisi")
	}

	if !isValidUUID(in.Id) {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus UUID yang valid")
	}

	postRepo.pb = postPb.Post{Id: in.Id}

	err := postRepo.GetPost(ctx)
	if err != nil {
		return &postRepo.pb, err
	}

	// Log intent to call other services based on post type
	switch postRepo.pb.Type {
	case postPb.PostType_CONFERENCE:
		a.Log.Printf("TODO: call lms-conference-service for type_id: %s", postRepo.pb.TypeId)
	case postPb.PostType_MATERIAL:
		a.Log.Printf("TODO: call lms-material-service for type_id: %s", postRepo.pb.TypeId)
	case postPb.PostType_QUIZ:
		a.Log.Printf("TODO: call lms-quiz-service for type_id: %s", postRepo.pb.TypeId)
	case postPb.PostType_TASK:
		a.Log.Printf("TODO: call lms-task-service for type_id: %s", postRepo.pb.TypeId)
	}

	return &postRepo.pb, nil
}

func (a *PostService) UnpublishPost(ctx context.Context, in *postPb.Id) (*postPb.BoolMessage, error) {
	var postRepo PostRepository
	var err error
	postRepo.Log = a.Log

	if in.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus diisi")
	}

	if !isValidUUID(in.Id) {
		return nil, status.Errorf(codes.InvalidArgument, "Id harus UUID yang valid")
	}

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return nil, status.Errorf(codes.Internal, "Error beginning transaction: %v", err)
	}

	postRepo.pb = postPb.Post{Id: in.Id}

	err = postRepo.UnpublishPost(ctx)
	if err != nil {
		postRepo.tx.Rollback()
		return nil, err
	}
	postRepo.tx.Commit()

	return &postPb.BoolMessage{IsTrue: !postRepo.pb.IsPublished}, nil
}

func (a *PostService) CreateComment(ctx context.Context, in *postPb.CreateCommentRequest) (*postPb.Comment, error) {
	var postRepo PostRepository
	var commentRepo CommentRepository
	var err error
	postRepo.Log = a.Log
	commentRepo.Log = a.Log

	if in.PostId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "PostId harus diisi")
	}

	if !isValidUUID(in.PostId) {
		return nil, status.Errorf(codes.InvalidArgument, "PostId harus UUID yang valid")
	}

	if in.StudentId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "StudentId harus diisi")
	}

	if in.StudentName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "StudentName harus diisi")
	}

	if in.Comment == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Comment harus diisi")
	}

	postRepo.tx, err = a.Db.BeginTx(ctx, nil)
	if err != nil {
		a.Log.Println("Error beginning transaction: ", err)
		return nil, status.Errorf(codes.Internal, "Error beginning transaction: %v", err)
	}

	commentRepo.tx = postRepo.tx

	// Check if post allows comments
	isAllowToComment, err := postRepo.GetPostAllowComment(ctx, in.PostId)
	if err != nil {
		postRepo.tx.Rollback()
		return nil, err
	}

	if !isAllowToComment {
		postRepo.tx.Rollback()
		return nil, status.Errorf(codes.PermissionDenied, "Post tidak mengizinkan komentar")
	}

	commentRepo.pb = postPb.Comment{
		PostId:      in.PostId,
		StudentId:   in.StudentId,
		StudentName: in.StudentName,
		Comment:     in.Comment,
	}

	err = commentRepo.CreateComment(ctx)
	if err != nil {
		postRepo.tx.Rollback()
		return &commentRepo.pb, err
	}
	postRepo.tx.Commit()
	return &commentRepo.pb, nil
}
