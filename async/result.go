package async

import "github.com/epochtimeout/baselibrary/status"

// Result is a generic future result interface.
type Result[T any] interface {
	// Wait awaits the result.
	Wait() <-chan struct{}

	// Result returns a value and a status or zero.
	Result() (T, status.Status)

	// Status returns a status.
	Status() status.Status
}
