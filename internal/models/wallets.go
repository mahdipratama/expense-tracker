package models

import (
	"database/sql"
	"time"
)

type WalletModel struct {
	DB *sql.DB
}

type Wallet struct {
	ID      int
	Name    string
	Balance int
	Created time.Time
}
