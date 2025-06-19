package optparams

import (
	"errors"
)

// Func should initialize some fields in the passed receiver.
type Func[T any] func(receiver *T) error

// ErrFailFast indicates that a failure in the [Func] call causes the [Apply]
// call to terminate early.
var ErrFailFast = errors.New("fail fast")
