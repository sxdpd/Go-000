package dao

import (
	"github.com/pkg/errors"
)

var (
	ErrNoRows = errors.New("No rows")
)

func Query() error {
	return ErrNoRows
}
