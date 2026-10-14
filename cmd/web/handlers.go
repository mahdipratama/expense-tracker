package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(wallets)
	if err != nil {
		app.serverError(w, err)
		return
	}
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

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(wallet)
	if err != nil {
		app.serverError(w, err)
		return
	}
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := map[string]string{
		"ID":      strconv.Itoa(id),
		"message": "Wallet created",
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		app.serverError(w, err)
		return
	}

}
