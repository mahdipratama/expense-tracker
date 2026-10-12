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

func (m *WalletModel) Insert(name string, balance int) (int, error) {
	stmt := `INSERT INTO wallets (name, balance, created) 
	VALUES (?,?, UTC_TIMESTAMP())`

	result, err := m.DB.Exec(stmt, name, balance)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
