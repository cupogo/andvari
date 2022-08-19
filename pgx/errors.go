package pgx

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrEmptyPK   = errors.New("empty pk")
	ErrEmptyKey  = errors.New("empty key")
	ErrDuplicate = errors.New("duplicate")
	ErrInternal  = errors.New("internal error")
)
