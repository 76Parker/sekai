package dto

type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}
