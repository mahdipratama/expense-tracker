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

func (m *WalletModel) Transfer(fromID, toID, amount int) error {
	if fromID == toID {
		return ErrDuplicateWallet
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Urutan ID (min dan max) untuk menghindari deadlock
	/*
			Tujuan utamanya hanya satu: Memaksa database untuk selalu mengunci (lock) baris data dengan urutan yang sama persis, dari ID terkecil ke ID terbesar, tidak peduli arah transfernya dari mana ke mana.

		teknik standar untuk mencegah Deadlock (situasi di mana database macet atau hang karena dua proses saling mengunci satu sama lain).
	*/
	w1ID, w2ID := fromID, toID

	if w1ID > w2ID {
		w1ID, w2ID = w2ID, w1ID
	}

	balances := make(map[int]int, 2)
	for _, id := range []int{w1ID, w2ID} {
		var bal int
		err := tx.QueryRow("SELECT balance FROM wallets WHERE id = ? FOR UPDATE", id).Scan(&bal)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNoRecord
			}
			return err
		}
		balances[id] = bal
	}

	if balances[fromID] < amount {
		return ErrInsufficientBalance
	}

	// UPDATE kedua Wallet
	res, err := tx.Exec("UPDATE wallets SET balance = balance - ? WHERE id = ? AND BALANCE >= ?", amount, fromID, amount)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNoRowsAffected
	}

	res, err = tx.Exec("UPDATE wallets SET balance = balance + ? WHERE id = ?", amount, toID)
	if err != nil {
		return err
	}

	n, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNoRowsAffected
	}

	return tx.Commit()
}
