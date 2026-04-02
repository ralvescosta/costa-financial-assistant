package errors

type AppError struct {
	Message   string
	Retryable bool
	Err       error
}

func New(message string) *AppError {
	return &AppError{
		Message: message,
	}
}

func NewRetryable(message string) *AppError {
	return &AppError{
		Message:   message,
		Retryable: true,
	}
}

func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) String() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
