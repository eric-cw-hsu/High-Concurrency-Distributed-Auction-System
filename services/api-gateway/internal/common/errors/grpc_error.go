package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGRPCError converts gRPC error to HTTP error response
func HandleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	httpStatus, errorResponse := mapGRPCCodeToHTTP(st.Code(), st.Message())
	c.JSON(httpStatus, errorResponse)
}

// mapGRPCCodeToHTTP maps gRPC status code to HTTP status code and response
func mapGRPCCodeToHTTP(code codes.Code, message string) (int, gin.H) {
	switch code {
	// 400 Bad Request
	case codes.InvalidArgument:
		return http.StatusBadRequest, gin.H{
			"error": message,
			"code":  "INVALID_ARGUMENT",
		}

	case codes.FailedPrecondition:
		return http.StatusBadRequest, gin.H{
			"error": message,
			"code":  "FAILED_PRECONDITION",
		}

	case codes.OutOfRange:
		return http.StatusBadRequest, gin.H{
			"error": message,
			"code":  "OUT_OF_RANGE",
		}

	// 401 Unauthorized
	case codes.Unauthenticated:
		return http.StatusUnauthorized, gin.H{
			"error": message,
			"code":  "UNAUTHENTICATED",
		}

	// 403 Forbidden
	case codes.PermissionDenied:
		return http.StatusForbidden, gin.H{
			"error": message,
			"code":  "PERMISSION_DENIED",
		}

	// 404 Not Found
	case codes.NotFound:
		return http.StatusNotFound, gin.H{
			"error": message,
			"code":  "NOT_FOUND",
		}

	// 409 Conflict
	case codes.AlreadyExists:
		return http.StatusConflict, gin.H{
			"error": message,
			"code":  "ALREADY_EXISTS",
		}

	case codes.Aborted:
		return http.StatusConflict, gin.H{
			"error": message,
			"code":  "ABORTED",
		}

	// 429 Too Many Requests
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, gin.H{
			"error": message,
			"code":  "RESOURCE_EXHAUSTED",
		}

	// 499 Client Closed Request
	case codes.Canceled:
		return 499, gin.H{
			"error": message,
			"code":  "CANCELED",
		}

	// 500 Internal Server Error
	case codes.Internal:
		return http.StatusInternalServerError, gin.H{
			"error": "internal server error",
			"code":  "INTERNAL_ERROR",
		}

	case codes.DataLoss:
		return http.StatusInternalServerError, gin.H{
			"error": "data loss",
			"code":  "DATA_LOSS",
		}

	case codes.Unknown:
		return http.StatusInternalServerError, gin.H{
			"error": "unknown error",
			"code":  "UNKNOWN",
		}

	// 501 Not Implemented
	case codes.Unimplemented:
		return http.StatusNotImplemented, gin.H{
			"error": message,
			"code":  "UNIMPLEMENTED",
		}

	// 503 Service Unavailable
	case codes.Unavailable:
		return http.StatusServiceUnavailable, gin.H{
			"error": "service temporarily unavailable",
			"code":  "UNAVAILABLE",
		}

	// 504 Gateway Timeout
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, gin.H{
			"error": "request timeout",
			"code":  "DEADLINE_EXCEEDED",
		}

	// Default: 500
	default:
		return http.StatusInternalServerError, gin.H{
			"error": "internal server error",
			"code":  "UNKNOWN",
		}
	}
}
