package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
