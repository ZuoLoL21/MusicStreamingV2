package helpers

type ErrorResult struct {
	Message string
	Status  int
}

func (e *ErrorResult) Error() string {
	return e.Message
}
