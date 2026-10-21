package models

import "errors"

var (
	ErrNoRecord            = errors.New("models: not matching record found")
	ErrDuplicateWallet     = errors.New("models: duplicate wallet name")
	ErrInsufficientBalance = errors.New("model: insufficient balance")
	ErrNoRowsAffected      = errors.New("model: no rows affected")
	ErrInvalidAmount       = errors.New("model: amount must be greater than zero")
)
