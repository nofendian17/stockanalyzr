package grpcerr

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorMap maps domain-specific sentinel errors to gRPC status codes.
type ErrorMap map[error]codes.Code

// Handle translates domain errors into standardized gRPC status errors.
// It iterates through the provided errMap to find a matching error using errors.Is.
// If no match is found, it returns a codes.Internal error with a generic message
// to prevent sensitive internal details from leaking to the client.
func Handle(err error, errMap ErrorMap) error {
	if err == nil {
		return nil
	}

	for targetErr, code := range errMap {
		if errors.Is(err, targetErr) {
			return status.Error(code, err.Error())
		}
	}

	// Hide internal error details from the client
	return status.Error(codes.Internal, "internal error")
}
