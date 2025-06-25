// This file is licensed under the terms of the MIT License (see LICENSE file)
// Copyright (c) 2025 Pavel Tsayukov p.tsayukov@gmail.com

package optparams

import (
	"errors"
	"fmt"
	"reflect"
)

// Func should initialize some fields in the passed receiver.
//
// The typical usage is to use the final incoming parameter in a function
// signature as a type ...Func.
type Func[T any] func(receiver *T) error

// Default creates [Func] that sets the passed field to the specified default
// value if the field does not equal its zero value.
//
// Note that a [Func] call does not check that the field belongs to the [Func]
// receiver, but it returns an error if the pointer to the field is nil.
func Default[T any, V any](field *V, default_ V) Func[T] {
	return func(receiver *T) error {
		if field == nil {
			return fmt.Errorf("pointer %T to field in receiver %T is nil", field, *receiver)
		}

		var zero V
		if reflect.DeepEqual(*field, zero) {
			*field = default_
		}

		return nil
	}
}

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
