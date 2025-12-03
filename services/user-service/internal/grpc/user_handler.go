package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/yourorg/boilerplate/shared/proto/gen/common/v1"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/services/user-service/internal/domain"
	"github.com/yourorg/boilerplate/services/user-service/internal/service"
	"github.com/yourorg/boilerplate/shared/errors"
)

// UserHandler implements the gRPC UserService
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service *service.UserService
}

// NewUserHandler creates a new gRPC user handler
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		service: svc,
	}
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.service.GetUser(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetUserResponse{
		User: toProtoUser(user),
	}, nil
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := h.service.CreateUser(ctx, req.Email, req.Name, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateUserResponse{
		User: toProtoUser(user),
	}, nil
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err := h.service.UpdateUser(ctx, req.Id, req.Name, req.AvatarUrl)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdateUserResponse{
		User: toProtoUser(user),
	}, nil
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := h.service.DeleteUser(ctx, req.Id); err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers lists users with pagination
func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page := int(req.Pagination.Page)
	pageSize := int(req.Pagination.PageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, total, err := h.service.ListUsers(ctx, pageSize, offset, req.Search)
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = toProtoUser(user)
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))

	return &pb.ListUsersResponse{
		Users: protoUsers,
		Pagination: &commonv1.PaginationResponse{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalCount: total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetUserByEmail retrieves a user by email
func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserByEmailResponse, error) {
	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetUserByEmailResponse{
		User: toProtoUser(user),
	}, nil
}

// Helper functions

func toProtoUser(user *domain.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarUrl: user.AvatarURL,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.IsNotFound(err):
		return status.Error(codes.NotFound, err.Error())
	case errors.IsAlreadyExists(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.IsValidation(err):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.IsUnauthorized(err):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.IsForbidden(err):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
