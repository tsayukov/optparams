package optparams

import (
	"errors"
	"fmt"
)

// Func should initialize some fields in the passed receiver.
type Func[T any] func(receiver *T) error

// ErrFailFast indicates that a failure in the [Func] call causes the [Apply]
// call to terminate early.
var ErrFailFast = errors.New("fail fast")

// Apply applies all the passed [Func] to the non-nil receiver.
//   - Passing the nil receiver causes the error.
//   - If some [Func] calls return a non-nil error, that does not lead to early
//     termination, until the first [Func] call returns wrapped [ErrFailFast].
func Apply[T any](receiver *T, opts ...Func[T]) error {
	if receiver == nil {
		var zero T
		return fmt.Errorf("receiver %T is nil", zero)
	}

	var errs []error

	for _, o := range opts {
		if err := o(receiver); err != nil {
			errs = append(errs, err)

			if errors.Is(err, ErrFailFast) {
				break
			}
		}
	}

	return errors.Join(errs...)
}
