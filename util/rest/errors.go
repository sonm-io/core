package rest

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPStatusFromError(err error) int {
	switch err.(type) {
	case nil:
		return http.StatusOK
	default:
		s, ok := status.FromError(err)
		if !ok {
			return http.StatusInternalServerError
		}

		return HTTPStatusFromCode(s.Code())
	}
}

func HTTPStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unknown, codes.DataLoss, codes.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
