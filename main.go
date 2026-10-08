package main

import (
	"fmt"
	"os"
)

func main() {
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbPassword == "" {
		fmt.Println("Error: DB_PASSWORD not set")
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	fmt.Println("DSN:", dsn)
	// TODO: gunakan DSN untuk connect ke database

}
