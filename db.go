package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "Sandeep@123"
	dbname := "eventplanner"

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("DB ping failed:", err)
	}

	log.Println("✅ Connected to PostgreSQL")
}
