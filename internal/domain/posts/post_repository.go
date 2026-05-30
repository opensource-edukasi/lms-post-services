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
	Log *log.Logger
	tx  *sql.Tx
	pb  postPb.Post
}

type CommentRepository struct {
	db  *sql.DB
	Log *log.Logger
	tx  *sql.Tx
	pb  postPb.Comment
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

	err = stmt.QueryRowContext(ctx,
		a.pb.SubjectClassId,
		a.pb.TopicSubjectId,
		a.pb.Type,
		sql.NullString{String: a.pb.TypeId, Valid: a.pb.TypeId != ""},
		a.pb.Title,
		a.pb.Description,
		a.pb.FileType,
		sql.NullString{String: a.pb.StorageId, Valid: a.pb.StorageId != ""},
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

func (a *PostRepository) UpdatePost(ctx context.Context) error {
	query := `
		UPDATE posts SET
			subject_class_id = $1,
			topic_subject_id = $2,
			type_id = $3,
			title = $4,
			description = $5,
			file_type = $6,
			storage_id = $7,
			source = $8,
			is_allow_to_comment = $9,
			is_published = $10,
			updated_by = $11,
			updated_at = timezone('utc', NOW())
		WHERE id = $12
		RETURNING id, subject_class_id, topic_subject_id, type, type_id, title, description, file_type, storage_id, source, is_allow_to_comment, is_published, updated_by, created_at, updated_at
	`

	stmt, err := a.tx.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return status.Errorf(codes.Internal, "Prepare statement update post: %v", err)
	}
	defer stmt.Close()

	var typeId, storageId sql.NullString
	var postType int32

	err = stmt.QueryRowContext(ctx,
		a.pb.SubjectClassId,
		a.pb.TopicSubjectId,
		sql.NullString{String: a.pb.TypeId, Valid: a.pb.TypeId != ""},
		a.pb.Title,
		a.pb.Description,
		a.pb.FileType,
		sql.NullString{String: a.pb.StorageId, Valid: a.pb.StorageId != ""},
		a.pb.Source,
		a.pb.IsAllowToComment,
		a.pb.IsPublished,
		a.pb.UpdatedBy,
		a.pb.Id,
	).Scan(
		&a.pb.Id,
		&a.pb.SubjectClassId,
		&a.pb.TopicSubjectId,
		&postType,
		&typeId,
		&a.pb.Title,
		&a.pb.Description,
		&a.pb.FileType,
		&storageId,
		&a.pb.Source,
		&a.pb.IsAllowToComment,
		&a.pb.IsPublished,
		&a.pb.UpdatedBy,
		&a.pb.CreatedAt,
		&a.pb.UpdatedAt,
	)

	if err != nil {
		a.Log.Println("Error updating post: ", err)
		return status.Errorf(codes.Internal, "Exec update post: %v", err)
	}

	a.pb.Type = postPb.PostType(postType)
	if typeId.Valid {
		a.pb.TypeId = typeId.String
	}
	if storageId.Valid {
		a.pb.StorageId = storageId.String
	}

	return nil
}

func (a *PostRepository) GetPost(ctx context.Context) error {
	query := `
		SELECT id, subject_class_id, topic_subject_id, type, type_id, title, description, file_type, storage_id, source, is_allow_to_comment, is_published, updated_by, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	stmt, err := a.db.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return status.Errorf(codes.Internal, "Prepare statement get post: %v", err)
	}
	defer stmt.Close()

	var typeId, storageId sql.NullString
	var postType int32

	err = stmt.QueryRowContext(ctx, a.pb.Id).Scan(
		&a.pb.Id,
		&a.pb.SubjectClassId,
		&a.pb.TopicSubjectId,
		&postType,
		&typeId,
		&a.pb.Title,
		&a.pb.Description,
		&a.pb.FileType,
		&storageId,
		&a.pb.Source,
		&a.pb.IsAllowToComment,
		&a.pb.IsPublished,
		&a.pb.UpdatedBy,
		&a.pb.CreatedAt,
		&a.pb.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return status.Errorf(codes.NotFound, "Post not found")
	}

	if err != nil {
		a.Log.Println("Error getting post: ", err)
		return status.Errorf(codes.Internal, "Exec get post: %v", err)
	}

	a.pb.Type = postPb.PostType(postType)
	if typeId.Valid {
		a.pb.TypeId = typeId.String
	}
	if storageId.Valid {
		a.pb.StorageId = storageId.String
	}

	return nil
}

func (a *PostRepository) UnpublishPost(ctx context.Context) error {
	query := `
		UPDATE posts SET
			is_published = NOT is_published,
			updated_at = timezone('utc', NOW())
		WHERE id = $1
		RETURNING is_published
	`

	stmt, err := a.tx.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return status.Errorf(codes.Internal, "Prepare statement unpublish post: %v", err)
	}
	defer stmt.Close()

	var isPublished bool
	err = stmt.QueryRowContext(ctx, a.pb.Id).Scan(&isPublished)

	if err == sql.ErrNoRows {
		return status.Errorf(codes.NotFound, "Post not found")
	}

	if err != nil {
		a.Log.Println("Error unpublishing post: ", err)
		return status.Errorf(codes.Internal, "Exec unpublish post: %v", err)
	}

	a.pb.IsPublished = isPublished

	return nil
}

func (a *PostRepository) GetPostAllowComment(ctx context.Context, postId string) (bool, error) {
	query := `SELECT is_allow_to_comment FROM posts WHERE id = $1`

	stmt, err := a.tx.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return false, status.Errorf(codes.Internal, "Prepare statement get post allow comment: %v", err)
	}
	defer stmt.Close()

	var isAllowToComment bool
	err = stmt.QueryRowContext(ctx, postId).Scan(&isAllowToComment)

	if err == sql.ErrNoRows {
		return false, status.Errorf(codes.NotFound, "Post not found")
	}

	if err != nil {
		a.Log.Println("Error getting post allow comment: ", err)
		return false, status.Errorf(codes.Internal, "Exec get post allow comment: %v", err)
	}

	return isAllowToComment, nil
}

func (a *CommentRepository) CreateComment(ctx context.Context) error {
	query := `
		INSERT INTO student_posts
		(post_id, student_id, student_name, comment)
		VALUES
		($1, $2, $3, $4)
		RETURNING id, created_at
	`

	stmt, err := a.tx.PrepareContext(ctx, query)
	if err != nil {
		a.Log.Println("error prepare context: ", err)
		return status.Errorf(codes.Internal, "Prepare statement create comment: %v", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx,
		a.pb.PostId,
		a.pb.StudentId,
		a.pb.StudentName,
		a.pb.Comment,
	).Scan(
		&a.pb.Id,
		&a.pb.CreatedAt,
	)

	if err != nil {
		a.Log.Println("Error inserting comment: ", err)
		return status.Errorf(codes.Internal, "Exec create comment: %v", err)
	}

	return nil
}
