package pgx

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrEmptyPK   = errors.New("empty pk")
	ErrEmptyKey  = errors.New("empty key")
	ErrDuplicate = errors.New("duplicate")
	ErrInternal  = errors.New("internal error")
)

type errID struct {
	ID  string
	msg string
}

func (e *errID) Error() string {
	return fmt.Sprintf("DB error, id: '%s' is %s", e.ID, e.msg)
}

func NewErrNotFoundID(id string) error {
	return &errID{ID: id, msg: "not found"}
}

func NewErrInvalidID(id string) error {
	return &errID{ID: id, msg: "invalid"}
}

func NewErrExistedID(id string) error {
	return &errID{ID: id, msg: "existed"}
}
