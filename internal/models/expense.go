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

func (m ExpenseModel) All(WalletID *int) ([]*Expense, error) {
	stmt := `
	SELECT id, wallet_id, title, amount, category, spent_at
	FROM expense 
	`

	var args []any

	if WalletID != nil {
		stmt += `	WHERE wallet_id = ? `
		args = append(args, *WalletID)
	}

	stmt += `ORDER BY spent_at DESC;`

	rows, err := m.DB.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	expenses := make([]*Expense, 0)

	for rows.Next() {
		e := &Expense{}
		err = rows.Scan(&e.ID, &e.WalletID, &e.Title, &e.Amount, &e.Category, &e.SpentAt)

		if err != nil {
			return nil, err
		}

		expenses = append(expenses, e)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return expenses, nil

}

func (m *ExpenseModel) Spend(e *Expense) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// CHECK BALANCE FIRST
	var currentBalance int
	err = tx.QueryRow("SELECT balance FROM wallets WHERE id = ? FOR UPDATE ", e.WalletID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	if currentBalance < e.Amount {
		return ErrInsufficientBalance
	}

	// UPDATE WALLETS
	updateStmt := `
	UPDATE wallets 
  SET balance = balance - ? 
  WHERE id = ? AND balance >= ? ;
	`
	res, err := tx.Exec(updateStmt, e.Amount, e.WalletID, e.Amount)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if count == 0 {
		return ErrNoRowsAffected
	}

	err = tx.Commit()

	// INSERT EXPENSES
	insertStmt := `
	INSERT INTO expense (wallet_id, title, amount, category, spent_at)
	VALUES (?, ?, ?, ?, ?)
	`
	res, err = tx.Exec(insertStmt, e.WalletID, e.Title, e.Amount, e.Category, e.SpentAt)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = int(id)

	return err
}
