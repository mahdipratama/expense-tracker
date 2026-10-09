package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	mux := http.NewServeMux()

	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	appPort := os.Getenv("APP_PORT")

	if dbPassword == "" {
		fmt.Println("Error: DB_PASSWORD not set")
		return
	}

	addr := ":" + appPort

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)
	fmt.Print(dsn)

	db, err := OpenDB(dsn)
	if err != nil {
		log.Fatalf("Gagal koneksi: %v", err)
	}

	defer db.Close()

	mux.HandleFunc("/", home)

	log.Printf("Starting new server on %s", appPort)
	err = http.ListenAndServe(addr, mux)
	log.Fatal(err)

}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
