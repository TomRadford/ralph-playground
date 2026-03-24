package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"ralph-playground/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./data/foosball.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiServer, err := api.NewServer(db)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	api.RegisterHandlers(e, apiServer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
