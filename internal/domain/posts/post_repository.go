package posts

import (
	"context"
	"database/sql"
	postPb "lms-post-service/pb/posts"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostRepository struct {
	db  *sql.DB
	Log log.Logger
	tx  *sql.Tx
	pb  postPb.Post
}

func (a *PostRepository) CreatePost(ctx context.Context) error {
	query := `
		INSERT INTO posts
		(subject_class_id, topic_subject_id, type, type_id, title, description, file_type, storage_id, source, is_allow_to_comment, is_published, updated_by)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, updated_by, created_at, updated_at
	`

	stmt, err := a.tx.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return status.Errorf(codes.Internal, "Prepare statement create post: %v", err)
	}
	defer stmt.Close()

	a.pb.UpdatedBy = ctx.Value("user_id").(string)

	err = stmt.QueryRowContext(ctx,
		a.pb.SubjectClassId,
		a.pb.TopicSubjectId,
		a.pb.Type,
		a.pb.TypeId,
		a.pb.Title,
		a.pb.Description,
		a.pb.FileType,
		a.pb.StorageId,
		a.pb.Source,
		a.pb.IsAllowToComment,
		a.pb.IsPublished,
		a.pb.UpdatedBy,
	).Scan(
		&a.pb.Id,
		&a.pb.UpdatedBy,
		&a.pb.CreatedAt,
		&a.pb.UpdatedAt,
	)

	if err != nil {
		a.Log.Println("Error inserting post: ", err)
		return status.Errorf(codes.Internal, "Exec create post: %v", err)
	}

	return nil
}
