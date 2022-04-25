package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/hherman1/gorgina/db/persist"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	dbUrlKey = "DATABASE_URL"
	portKey  = "PORT"
)

// A series of schema setup queries to be run during startup.
//go:embed db/schema.sql
var initDB string

func main() {
	if err := run(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Parse args
	port, err := strconv.Atoi(os.Getenv(portKey))
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	// setup DB
	db, err := sql.Open("pgx", os.Getenv(dbUrlKey))
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	_, err = db.ExecContext(ctx, initDB)
	if err != nil {
		return fmt.Errorf("initialize DB tables: %w", err)
	}
	_ = persist.New(db)

	// Setup routes
	http.DefaultServeMux.Handle("/numbers", http.HandlerFunc(handleHelloWorld))

	// Start server
	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

// Hello world handler
func handleHelloWorld(response http.ResponseWriter, req *http.Request) {
	_, _ = response.Write([]byte("hello world"))
}
