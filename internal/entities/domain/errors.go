package domain

type Error struct {
	Model   string
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func NewError(model, message string) Error {
	return Error{
		Model:   model,
		Message: message,
	}
}
