package errors

var (
	ErrGenericError       = New("an error occurred")
	ErrUnformattedRequest = New("unformatted request body")
	ErrResourceNotFound   = New("resource not found")
	ErrDatabaseError      = NewRetryable("database error")
)
