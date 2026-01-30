package errfmt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

type ClassifiedError struct {
	Message string
	Hint    string
}

func Classify(err error) ClassifiedError {
	if err == nil {
		return ClassifiedError{}
	}

	var retryable *asc.RetryableError
	if errors.As(err, &retryable) {
		hints := make([]string, 0, 2)
		if retryable.RetryAfter > 0 {
			hints = append(hints, fmt.Sprintf("Retry after %s.", retryable.RetryAfter.Truncate(time.Second)))
		}
		if retryable.RateLimit != nil {
			if summary := retryable.RateLimit.Summary(); summary != "" {
				hints = append(hints, fmt.Sprintf("Quota: %s.", summary))
			}
		}
		if len(hints) == 0 {
			hints = append(hints, "Reduce request volume and try again.")
		}
		return ClassifiedError{
			Message: err.Error(),
			Hint:    strings.Join(hints, " "),
		}
	}

	if errors.Is(err, shared.ErrMissingAuth) {
		return ClassifiedError{
			Message: err.Error(),
			Hint:    "Run `asc auth login` or `asc auth init` (or set ASC_KEY_ID/ASC_ISSUER_ID/ASC_PRIVATE_KEY_PATH). Try `asc auth doctor` if you're unsure what's misconfigured.",
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return ClassifiedError{
			Message: err.Error(),
			Hint:    "Increase the request timeout (e.g. set `ASC_TIMEOUT=90s`).",
		}
	}

	if errors.Is(err, asc.ErrForbidden) {
		return ClassifiedError{
			Message: err.Error(),
			Hint:    "Check that your API key has the right role/permissions for this operation in App Store Connect.",
		}
	}

	if errors.Is(err, asc.ErrUnauthorized) {
		return ClassifiedError{
			Message: err.Error(),
			Hint:    "Your credentials may be invalid or expired. Try `asc auth status` and re-login if needed.",
		}
	}

	return ClassifiedError{
		Message: err.Error(),
		Hint:    "",
	}
}

func FormatStderr(err error) string {
	ce := Classify(err)
	if ce.Message == "" {
		return ""
	}
	if ce.Hint == "" {
		return fmt.Sprintf("Error: %s\n", ce.Message)
	}
	return fmt.Sprintf("Error: %s\nHint: %s\n", ce.Message, ce.Hint)
}

