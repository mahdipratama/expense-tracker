package main

import (
	"encoding/json"
	"net/http"
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
