package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/wallets", app.walletHandler)
	mux.HandleFunc("/wallet", app.GetWalletHandler)
	mux.HandleFunc("/wallet/create", app.createWalletHandler)
	mux.HandleFunc("/expenses", app.listExpensesHandler)
	mux.HandleFunc("/expense/create", app.createExpenseHandler)
	mux.HandleFunc("/transfer", app.transferHandler)
	mux.HandleFunc("/healthz", app.healthz)


	return mux
}
