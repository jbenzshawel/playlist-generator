package cerror

import (
	"log/slog"
)

type ValidationError struct {
	message     string
	fieldErrors map[string]string
}

func (r ValidationError) Error() string {
	return r.message
}

func (r ValidationError) LogValue() slog.Value {
	attr := []slog.Attr{
		slog.String("error", r.message),
	}
	for key, value := range r.fieldErrors {
		attr = append(attr, slog.String(key, value))
	}

	return slog.GroupValue()
}

func NewValidationError(errorMsg string, fieldErrors map[string]string) ValidationError {
	return ValidationError{message: errorMsg, fieldErrors: fieldErrors}
}
