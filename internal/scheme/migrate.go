package scheme

import (
	"database/sql"

	"github.com/GuiaBolso/darwin"
)

var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Create uuid extension",
		Script:      `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
	},
	{
		Version:     2,
		Description: "Create posts table",
		Script: `
			CREATE TABLE posts (
				id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
				subject_class_id uuid NOT NULL,
				topic_subject_id uuid NOT NULL,
				type char(5) NOT Null,
				type_id uuid,
				title varchar(45) NOT NULL,
				description varchar(255),
				file_type char(1),
				storage_id uuid,
				source varchar(128),
				is_allow_to_comment bool NOT NULL,
				updated_at timestamptz NOT NULL DEFAULT timezone('utc', NOW()),
				updated_by uuid,
				created_at timestamptz NOT NULL DEFAULT timezone('utc', NOW())
			);
			CREATE INDEX idx_subject_class_id ON posts(subject_class_id);
			CREATE INDEX idx_topic_subject_id ON posts(topic_subject_id);
			CREATE INDEX idx_type_id ON posts(type_id);
			CREATE INDEX idx_storage_id ON posts(storage_id);
		`,
	},
	{
		Version:     3,
		Description: "Create student_posts table",
		Script: `
			CREATE TABLE student_posts (
				id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
				post_id uuid NOT NULL,
				student_id uuid NOT NULL, 
				student_name varchar(45) NOT NULL, 
				comment text NOT NULL,
				created_at timestamptz NOT NULL DEFAULT timezone('utc', NOW()),
				CONSTRAINT fk_student_posts_to_posts FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE 
			);
			CREATE INDEX idx_student_id ON student_posts(student_id);
		`,
	},
}

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sql.DB) error {
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	return d.Migrate()
}
