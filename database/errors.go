package database

import (
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func prepareError(err error) error {
	if err == nil {
		return nil
	}
	_, ok := status.FromError(err)
	if !ok {
		if err == ErrNotFound {
			return status.Error(codes.NotFound, err.Error())
		}
		if err == ErrAlreadyExists {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if err == ErrUnsupportedAlgorithm {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if err == ErrFailedToGetUser {
			return status.Error(codes.Unauthenticated, err.Error())
		}
		if err == bbolt.ErrTimeout {
			return status.Error(codes.Internal, "internal error")
		}
		if strings.Contains(err.Error(), "does not exist") {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if strings.Contains(err.Error(), "trigger: ") {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		return status.Error(codes.Internal, err.Error())
	}
	return err
}
