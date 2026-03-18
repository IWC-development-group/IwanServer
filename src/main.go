package main

import (
	"os"
	"fmt"

	"database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

type ModuleFunc func(db *sql.DB, argOffset int)

func main() {
	commands := map[string]ModuleFunc{
		"serve": ServerMain,
		"index": IndexerMain,
	}

	moduleName := os.Args[1]
	fn, exists := commands[moduleName]
	if !exists {
		fmt.Printf("Module not found: %s!\n", moduleName)
		os.Exit(0)
	}

	db, err := sql.Open("sqlite3", "page_registry.db")
	if err != nil { panic(err) }
	defer db.Close()

	fn(db, 1)
	db.Close()
}