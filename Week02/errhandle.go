package main

import (
	"errhandle/dao"
	"fmt"

	"github.com/pkg/errors"
)

var (
	errNoRecord    = errors.New("No record")
	errDbException = errors.New("Database exception")
)

func main() {
	err := service()
	fmt.Printf("main: %+v\n", err)
}

func biz() error {
	err := dao.Query()
	if err != nil {
		var bizErr error
		if errors.Is(err, dao.ErrNoRows) {
			bizErr = errNoRecord
		} else {
			bizErr = errDbException
		}
		return errors.Wrap(bizErr, "dao")
	}
	return nil
}

func service() error {
	return biz()
}
