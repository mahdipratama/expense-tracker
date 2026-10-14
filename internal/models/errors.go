package models

import "errors"

var (
	ErrNoRecord        = errors.New("models: not matching record found")
	ErrDuplicateWallet = errors.New("models: duplicate wallet name")
)
