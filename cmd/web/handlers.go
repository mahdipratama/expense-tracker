package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mahdipratama/expense-tracker/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from expense_tracker"))
}

func (app *application) walletHandler(w http.ResponseWriter, r *http.Request) {
	wallets, err := app.wallets.All()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.writeJSONResponseBody(w, http.StatusOK, wallets)
}

func (app *application) GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	wallet, err := app.wallets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.writeJSONResponseBody(w, http.StatusOK, wallet)
}

func (app *application) createWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, "Failed to parse form")

		return
	}

	name := r.PostForm.Get("name")
	balanceStr := r.PostForm.Get("balance")
	// name := r.URL.Query().Get("name")
	// balanceStr := r.URL.Query().Get("balance")

	if strings.TrimSpace(name) == "" {
		app.badRequest(w, "Wallet name can not be empty")
		return
	}

	balance, err := strconv.Atoi(balanceStr)
	if err != nil {
		app.badRequest(w, "balance must be a valid number")
		return
	}

	if balance <= 0 {
		app.badRequest(w, "balance can not be negative or zero")
		return
	}

	id, err := app.wallets.Insert(name, balance)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateWallet) {
			app.badRequest(w, "Wallet with this name already exists")
			return
		}

		app.serverError(w, err)
		return
	}

	resp := map[string]string{
		"ID":      strconv.Itoa(id),
		"message": "Wallet created",
	}

	app.writeJSONResponseBody(w, http.StatusCreated, resp)
}

func (app *application) listExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	var walletID *int = nil
	walletIDStr := r.URL.Query().Get("wallet")

	if walletIDStr != "" {
		var err error
		id, err := strconv.Atoi(walletIDStr)
		if err != nil || id < 1 {
			app.badRequest(w, "invalid wallet_id parameter")
			return
		}
		walletID = &id
	}

	expenses, err := app.expense.All(walletID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.writeJSONResponseBody(w, http.StatusOK, expenses)

}

func (app *application) createExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, "Failed to parse form")
		return
	}

	walletIDStr := r.PostForm.Get("wallet")
	title := strings.TrimSpace(r.PostForm.Get("title"))
	amountStr := r.PostForm.Get("amount")
	category := strings.TrimSpace(r.PostForm.Get("category"))

	walletID, err := strconv.Atoi(walletIDStr)
	if err != nil || walletID < 1 {
		app.badRequest(w, "Invalid wallet ID")
		return
	}

	wallet, err := app.wallets.Get(walletID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	if title == "" || category == "" {
		app.badRequest(w, "Wallet name and category can not be empty")
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		app.badRequest(w, "amount must be a valid number")
		return
	}

	if amount <= 0 {
		app.badRequest(w, "amount can not be negative or zero")
		return
	}

	expense := &models.Expense{
		WalletID: wallet.ID,
		Title:    title,
		Amount:   amount,
		Category: category,
		SpentAt:  time.Now(),
	}

	err = app.expense.Spend(expense)
	if err != nil {
		if errors.Is(err, models.ErrInsufficientBalance) {
			app.badRequest(w, "Insufficient balance")
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.writeJSONResponseBody(w, http.StatusCreated, expense)

}

func (app *application) transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, "Failed to parsed form")
		return
	}

	fromIDStr := r.PostForm.Get("from")
	toIDStr := r.PostForm.Get("to")
	amountStr := strings.TrimSpace(r.PostForm.Get("amount"))

	fromID, err := strconv.Atoi(fromIDStr)
	if err != nil || fromID < 1 {
		app.badRequest(w, "Invalid wallet ID")
		return
	}

	toID, err := strconv.Atoi(toIDStr)
	if err != nil || toID < 1 {
		app.badRequest(w, "Invalid wallet ID")
		return
	}

	if fromID == toID {
		app.badRequest(w, "The source and target wallets cannot be identical.")
		return
	}

	walletsToCheck := []int{fromID, toID}
	for _, id := range walletsToCheck {
		_, err := app.wallets.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				app.notFound(w)
				return
			} else {
				app.serverError(w, err)
				return
			}
		}
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		app.badRequest(w, "amount must be a valid number")
		return
	}

	if amount <= 0 {
		app.badRequest(w, "amount can not be negative or zero")
		return
	}

	err = app.wallets.Transfer(fromID, toID, amount)
	if err != nil {
		if errors.Is(err, models.ErrInsufficientBalance) {
			app.badRequest(w, "Insufficient balance")
		} else if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	type transfer struct {
		Message string
		FromID  int
		ToID    int
		Amount  int
	}

	transferResponse := &transfer{
		Message: "Transfer successful",
		FromID:  fromID,
		ToID:    toID,
		Amount:  amount,
	}

	app.writeJSONResponseBody(w, http.StatusOK, transferResponse)

}

func (app *application) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err := app.db.Ping()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.writeJSONResponseBody(w, http.StatusOK, "Server: OK!")

}
