package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError converts domain errors to gRPC status errors
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case IsNotFound(err):
		return status.Error(codes.NotFound, err.Error())
	case IsAlreadyExists(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case IsValidation(err):
		return status.Error(codes.InvalidArgument, err.Error())
	case IsUnauthorized(err):
		return status.Error(codes.Unauthenticated, err.Error())
	case IsForbidden(err):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// FromGRPCError converts gRPC status errors to domain errors
func FromGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return NewNotFoundError("resource", st.Message())
	case codes.AlreadyExists:
		return NewAlreadyExistsError("resource", "field", st.Message())
	case codes.InvalidArgument:
		return NewValidationError(st.Message())
	case codes.Unauthenticated:
		return NewUnauthorizedError(st.Message())
	case codes.PermissionDenied:
		return NewForbiddenError(st.Message())
	default:
		return New("INTERNAL", st.Message())
	}
}
