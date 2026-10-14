package models

import (
	"database/sql"
	"time"
)

type ExpenseModel struct {
	DB *sql.DB
}

type Expense struct {
	ID       int
	WalletID int
	Title    string
	Amount   int
	Category string
	SpentAt  time.Time
}
