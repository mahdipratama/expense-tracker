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
