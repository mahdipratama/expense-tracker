package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
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
		mysqlErr, ok := err.(*mysql.MySQLError)

		if ok && mysqlErr.Number == 1062 {
			return 0, ErrDuplicateWallet
		} else {
			return 0, err
		}
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *WalletModel) Get(id int) (*Wallet, error) {
	stmt := `SELECT id, name, balance, created FROM wallets WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	w := &Wallet{}

	err := row.Scan(&w.ID, &w.Name, &w.Balance, &w.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return w, nil
}

func (m *WalletModel) All() ([]*Wallet, error) {
	stmt := `SELECT id, name, balance, created FROM wallets `
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	wallets := make([]*Wallet, 0)

	for rows.Next() {
		w := &Wallet{}
		err = rows.Scan(&w.ID, &w.Name, &w.Balance, &w.Created)

		if err != nil {
			return nil, err
		}

		wallets = append(wallets, w)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return wallets, nil
}
