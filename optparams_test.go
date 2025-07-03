// This file is licensed under the terms of the MIT License (see LICENSE file)
// Copyright (c) 2025 Pavel Tsayukov p.tsayukov@gmail.com

package optparams

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Apply(t *testing.T) {
	t.Parallel()

	type mockReceiver struct {
		values []rune
	}

	newMockReceiver := func(count ...int) *mockReceiver {
		if len(count) == 0 {
			return &mockReceiver{}
		}

		return &mockReceiver{values: make([]rune, count[0])}
	}

	newMockReceiverOf := func(values ...rune) *mockReceiver {
		return &mockReceiver{values: values}
	}

	newOpt := func(pos int, value rune, err ...error) Func[mockReceiver] {
		return func(mr *mockReceiver) error {
			if len(err) > 0 && err[0] != nil {
				return err[0]
			}

			mr.values[pos] = value

			return nil
		}
	}

	errMocks := func() []error {
		const size = 5

		errs := make([]error, 0, size)
		for i := 0; i < size; i++ {
			errs = append(errs, errors.New(strconv.Itoa(i)))
		}

		return errs
	}()

	type testCase struct {
		name     string
		receiver *mockReceiver
		want     *mockReceiver
		opts     []Func[mockReceiver]
		wantErr  error
	}

	tests := []testCase{
		{
			name:    "nil receiver",
			wantErr: fmt.Errorf("receiver %T is nil", mockReceiver{}),
		},
		{
			name:     "no opts",
			receiver: newMockReceiver(),
			want:     newMockReceiver(),
		},
		{
			name:     "one successful opt",
			receiver: newMockReceiver(1),
			want:     newMockReceiverOf('a'),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
			},
		},
		{
			name:     "one failed opt",
			receiver: newMockReceiver(),
			want:     newMockReceiver(),
			opts: []Func[mockReceiver]{
				newOpt(0, 'x', errMocks[0]),
			},
			wantErr: errMocks[0],
		},
		{
			name:     "multiple opts",
			receiver: newMockReceiver(7),
			want:     newMockReceiverOf('a', 'b', 'c', 'd', 'e', 'f', 'g'),
			opts: func() []Func[mockReceiver] {
				opts := []Func[mockReceiver]{
					newOpt(0, 'a'),
					newOpt(1, 'b'),
					newOpt(2, 'c'),
					newOpt(3, 'd'),
					newOpt(4, 'e'),
					newOpt(5, 'f'),
					newOpt(6, 'g'),
				}

				rand.Shuffle(len(opts), func(i, j int) {
					opts[i], opts[j] = opts[j], opts[i]
				})

				return opts
			}(),
		},
		{
			name:     "multiple opts with first fast fail error",
			receiver: newMockReceiver(3),
			want:     newMockReceiverOf(0, 0, 0),
			opts: []Func[mockReceiver]{
				newOpt(2, 'x', ErrFailFast),
				newOpt(1, 'b'),
				newOpt(0, 'a'),
			},
			wantErr: ErrFailFast,
		},
		{
			name:     "partial init with fast fail error",
			receiver: newMockReceiver(3),
			want:     newMockReceiverOf(0, 'b', 0),
			opts: []Func[mockReceiver]{
				newOpt(1, 'b'),
				newOpt(2, 'x', ErrFailFast),
				newOpt(0, 'a'),
			},
			wantErr: ErrFailFast,
		},
		{
			name:     "wrapped fail fast error",
			receiver: newMockReceiver(3),
			want:     newMockReceiverOf(0, 0, 0),
			opts: []Func[mockReceiver]{
				newOpt(0, 'x', fmt.Errorf("err: %w", ErrFailFast)),
				newOpt(1, 'b'),
				newOpt(1, 'c'),
			},
			wantErr: fmt.Errorf("err: %w", ErrFailFast),
		},
		{
			name:     "multiple opts with one error",
			receiver: newMockReceiver(3),
			want:     newMockReceiverOf('a', 0, 'c'),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
				newOpt(1, 'x', errMocks[0]),
				newOpt(2, 'c'),
			},
			wantErr: errMocks[0],
		},
		{
			name:     "multiple opts with a few errors",
			receiver: newMockReceiver(6),
			want:     newMockReceiverOf('a', 0, 'c', 0, 0, 'f'),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
				newOpt(1, 'x', errMocks[0]),
				newOpt(2, 'c'),
				newOpt(3, 'x', errMocks[1]),
				newOpt(4, 'x', errMocks[2]),
				newOpt(5, 'f'),
			},
			wantErr: errors.Join(errMocks[0], errMocks[1], errMocks[2]),
		},
		{
			name:     "multiple opts with all errors",
			receiver: newMockReceiver(len(errMocks)),
			want:     newMockReceiverOf(slices.Repeat([]rune{0}, len(errMocks))...),
			opts: func() []Func[mockReceiver] {
				opts := make([]Func[mockReceiver], 0, len(errMocks))
				for i := range errMocks {
					opts = append(opts, newOpt(i, 'x', errMocks[i]))
				}

				return opts
			}(),
			wantErr: errors.Join(errMocks...),
		},
		{
			name:     "nil field",
			receiver: newMockReceiver(),
			want:     newMockReceiver(),
			opts: []Func[mockReceiver]{
				Default[mockReceiver](nil, 'a'),
			},
			wantErr: fmt.Errorf(
				"pointer %T to field in receiver %T is nil",
				new(rune), mockReceiver{},
			),
		},
		func() testCase {
			mr := newMockReceiver(1)
			return testCase{
				name:     "no opt + default",
				receiver: mr,
				want:     newMockReceiverOf('a'),
				opts: []Func[mockReceiver]{
					Default[mockReceiver](&mr.values[0], 'a'),
				},
			}
		}(),
		func() testCase {
			mr := newMockReceiver(1)
			return testCase{
				name:     "opt + default",
				receiver: mr,
				want:     newMockReceiverOf('a'),
				opts: []Func[mockReceiver]{
					newOpt(0, 'a'),
					Default[mockReceiver](&mr.values[0], 'a'),
				},
			}
		}(),
		{
			name:     "empty Join",
			receiver: newMockReceiver(),
			want:     newMockReceiver(),
			opts: []Func[mockReceiver]{
				Join[mockReceiver](),
			},
		},
		{
			name:     "Join with one opt",
			receiver: newMockReceiver(1),
			want:     newMockReceiverOf('a'),
			opts: []Func[mockReceiver]{
				Join[mockReceiver](newOpt(0, 'a')),
			},
		},
		{
			name:     "Join with a few opts",
			receiver: newMockReceiver(5),
			want:     newMockReceiverOf('a', 'b', 'c', 'd', 'e'),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
				Join[mockReceiver](
					newOpt(1, 'b'),
					newOpt(2, 'c'),
					newOpt(3, 'd'),
				),
				newOpt(4, 'e'),
			},
		},
		{
			name:     "Join with a few opts + fast fail error",
			receiver: newMockReceiver(5),
			want:     newMockReceiverOf('a', 'b', 0, 0, 0),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
				Join[mockReceiver](
					newOpt(1, 'b'),
					newOpt(2, 'c', ErrFailFast),
					newOpt(3, 'd'),
				),
				newOpt(4, 'e'),
			},
			wantErr: ErrFailFast,
		},
		{
			name:     "Join with a few opts + error",
			receiver: newMockReceiver(5),
			want:     newMockReceiverOf('a', 'b', 0, 'd', 'e'),
			opts: []Func[mockReceiver]{
				newOpt(0, 'a'),
				Join[mockReceiver](
					newOpt(1, 'b'),
					newOpt(2, 'c', errMocks[0]),
					newOpt(3, 'd'),
				),
				newOpt(4, 'e'),
			},
			wantErr: errMocks[0],
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Apply(tt.receiver, tt.opts...)

			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, tt.receiver)
		})
	}
}
