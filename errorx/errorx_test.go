package errorx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testErr     = errors.New("test error")
	testCause   = errors.New("test cause")
	testPkg     = "testpkg"
	testOp      = "testOp"
	testDetails = "test details"
	testErr2    = errors.New("test error 2")
	testErr3    = errors.New("test error 3")
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "full error with all fields",
			err: &Error{
				Pkg:     testPkg,
				Err:     testErr,
				Op:      testOp,
				Details: testDetails,
				Cause:   testCause,
			},
			want: "testpkg.testOp: test error | details=test details | cause=test cause",
		},
		{
			name: "error without op",
			err: &Error{
				Pkg: testPkg,
				Err: testErr,
			},
			want: "testpkg: test error",
		},
		{
			name: "error with only pkg and err",
			err: &Error{
				Pkg: testPkg,
				Err: testErr,
			},
			want: "testpkg: test error",
		},
		{
			name: "error with cause but no err",
			err: &Error{
				Pkg:   testPkg,
				Err:   nil,
				Cause: testCause,
			},
			want: "testpkg: unknown error (caused by: test cause)",
		},
		{
			name: "error with nil err and nil cause",
			err: &Error{
				Pkg: testPkg,
				Err: nil,
			},
			want: "testpkg: unknown error",
		},
		{
			name: "error with details only",
			err: &Error{
				Pkg:     testPkg,
				Err:     testErr,
				Details: testDetails,
			},
			want: "testpkg: test error | details=test details",
		},
		{
			name: "error with cause only",
			err: &Error{
				Pkg:   testPkg,
				Err:   testErr,
				Cause: testCause,
			},
			want: "testpkg: test error | cause=test cause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	err := &Error{
		Pkg:   testPkg,
		Err:   testErr,
		Cause: testCause,
	}

	assert.Equal(t, testCause, err.Unwrap())
	assert.True(t, errors.Is(err, testCause))
}

func TestError_Is(t *testing.T) {
	err := &Error{
		Pkg: testPkg,
		Err: testErr,
	}

	assert.True(t, err.Is(testErr))
	assert.False(t, err.Is(testErr2))
	assert.False(t, err.Is(testCause))
}

func TestError_In(t *testing.T) {
	err := &Error{
		Pkg: testPkg,
		Err: testErr,
	}

	assert.True(t, err.In(testErr))
	assert.True(t, err.In(testErr, testErr2))
	assert.False(t, err.In(testErr2, testErr3))
	assert.False(t, err.In())
}

func TestIn(t *testing.T) {
	assert.True(t, In(testErr, testErr))
	assert.True(t, In(testErr, testErr, testErr2))
	assert.False(t, In(testErr, testErr2, testErr3))
	assert.False(t, In(testErr))
	assert.False(t, In(nil, testErr))
}

func TestNew(t *testing.T) {
	err := New(testErr, testPkg, testOp, testCause)

	assert.NotNil(t, err)
	errorx, ok := err.(*Error)
	assert.True(t, ok)
	assert.Equal(t, testPkg, errorx.Pkg)
	assert.Equal(t, testErr, errorx.Err)
	assert.Equal(t, testOp, errorx.Op)
	assert.Equal(t, testCause, errorx.Cause)
	assert.Empty(t, errorx.Details)
}

func TestNewWithDetails(t *testing.T) {
	err := NewWithDetails(testErr, testPkg, testOp, testDetails, testCause)

	assert.NotNil(t, err)
	errorx, ok := err.(*Error)
	assert.True(t, ok)
	assert.Equal(t, testPkg, errorx.Pkg)
	assert.Equal(t, testErr, errorx.Err)
	assert.Equal(t, testOp, errorx.Op)
	assert.Equal(t, testDetails, errorx.Details)
	assert.Equal(t, testCause, errorx.Cause)
}

func TestErrorsIsWithWrappedError(t *testing.T) {
	wrapped := New(testErr, testPkg, testOp, nil)

	assert.True(t, errors.Is(wrapped, testErr))
	assert.False(t, errors.Is(wrapped, testErr2))
}
