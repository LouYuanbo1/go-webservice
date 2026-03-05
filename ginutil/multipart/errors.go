package multipart

import (
	"errors"
)

var (
	ErrInvalidObject         = errors.New("ginutil/multipart: invalid object")
	ErrParseMultipartForm    = errors.New("ginutil/multipart: parse multipart form error")
	ErrDecodeForm            = errors.New("ginutil/multipart: decode form error")
	ErrFillFiles             = errors.New("ginutil/multipart: fill files error")
	ErrDuplicateBracketIndex = errors.New("ginutil/multipart: duplicate bracket index")
	ErrDuplicateDotIndex     = errors.New("ginutil/multipart: duplicate dot index")
	ErrInvalidIndex          = errors.New("ginutil/multipart: invalid index")
)
