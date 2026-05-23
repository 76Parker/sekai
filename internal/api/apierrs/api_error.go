package apierrs


type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}

var (
	ErrNotFound = APIError{Status: 404, Code: "NOT_FOUND", Message: "resource not found"}
	ErrInternalServerError = APIError{Status: 500, Code: "INTERNAL_SERVER_ERROR", Message: "internal server error"}
	ErrBadRequest = APIError{Status: 400, Code: "BAD_REQUEST", Message: "invalid request"}
	ErrUnauthorized = APIError{Status: 401, Code: "UNAUTHORIZED", Message: "authentication required"}

)
