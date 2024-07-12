package route

import (
	"database/sql"
	"log"

	postDomain "lms-post-service/internal/domain/posts"
	"lms-post-service/internal/pkg/db/redis"
	postPb "lms-post-service/pb/posts"

	"google.golang.org/grpc"
)

// GrpcRoute func
func GrpcRoute(grpcServer *grpc.Server, db *sql.DB, log *log.Logger, cache *redis.Cache) {
	postServer := postDomain.PostService{Db: db, Cache: cache, Log: log}
	postPb.RegisterPostsServer(grpcServer, &postServer)
}
