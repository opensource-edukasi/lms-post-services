// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v5.27.2
// source: posts/post_service.proto

package posts

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Posts_CreatePost_FullMethodName = "/posts.Posts/CreatePost"
)

// PostsClient is the client API for Posts service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PostsClient interface {
	CreatePost(ctx context.Context, in *CreatePostRequest, opts ...grpc.CallOption) (*Post, error)
}

type postsClient struct {
	cc grpc.ClientConnInterface
}

func NewPostsClient(cc grpc.ClientConnInterface) PostsClient {
	return &postsClient{cc}
}

func (c *postsClient) CreatePost(ctx context.Context, in *CreatePostRequest, opts ...grpc.CallOption) (*Post, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Post)
	err := c.cc.Invoke(ctx, Posts_CreatePost_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PostsServer is the server API for Posts service.
// All implementations should embed UnimplementedPostsServer
// for forward compatibility
type PostsServer interface {
	CreatePost(context.Context, *CreatePostRequest) (*Post, error)
}

// UnimplementedPostsServer should be embedded to have forward compatible implementations.
type UnimplementedPostsServer struct {
}

func (UnimplementedPostsServer) CreatePost(context.Context, *CreatePostRequest) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePost not implemented")
}

// UnsafePostsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PostsServer will
// result in compilation errors.
type UnsafePostsServer interface {
	mustEmbedUnimplementedPostsServer()
}

func RegisterPostsServer(s grpc.ServiceRegistrar, srv PostsServer) {
	s.RegisterService(&Posts_ServiceDesc, srv)
}

func _Posts_CreatePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PostsServer).CreatePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Posts_CreatePost_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PostsServer).CreatePost(ctx, req.(*CreatePostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Posts_ServiceDesc is the grpc.ServiceDesc for Posts service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Posts_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "posts.Posts",
	HandlerType: (*PostsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreatePost",
			Handler:    _Posts_CreatePost_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "posts/post_service.proto",
}
