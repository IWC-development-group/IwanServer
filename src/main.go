package main

import (
	"os"
	"fmt"
)

func main() {
	var commands map[string]func() {
		"serve": ServerMain
		"index": IndexerMain
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
}