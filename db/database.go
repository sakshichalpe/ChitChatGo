package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	dsn := "host=localhost port=5433 user=postgres password=1234 dbname=postgres sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Database Connection Failed:", err)
		return
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Database is not reachable: %v", err) //verfify the credencial
		return
	}

	log.Println("Connected to PostgreSQL successfully!")
}
