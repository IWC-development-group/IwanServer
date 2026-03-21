package main

import (
	"os"
//	"fmt"
	"path/filepath"
	"github.com/spf13/cobra"

	"database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

type ModuleFunc func(db *sql.DB, argOffset int)
type CobraCallbackErr func(cmd *cobra.Command, args []string) error
type CobraCallback func(cmd *cobra.Command, args []string)

func main() {
	rootCmd := &cobra.Command{
		Use: "iwans",
		Short: "documentation server",
	}

	execPath, _ := os.Executable()
	dbPath := filepath.Join(filepath.Dir(execPath), "page_registry.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil { panic(err) }
	defer db.Close()

	var port int
	serveCmd := &cobra.Command{
		Use: "serve",
		Short: "Host the documentation server",
		Run: func(cmd *cobra.Command, args []string){
			ServerMain(db, port)
		},
	}
	serveCmd.Flags().IntVarP(&port, "port", "p", 8085, "Set the server port")

	var namespace string
	var forced bool
	indexCmd := &cobra.Command{
		Use: "index",
		Short: "Generate indexes for pages in the specified directory",
		Run: func(cmd *cobra.Command, args []string){
			var path string
			if len(args) > 0 {
				path = args[0]
			} else {
				path = "."
			}

			IndexerMain(db, path, namespace, forced)
		},
	}
	indexCmd.Flags().StringVarP(&namespace, "namespace", "ns", "global", "Documentation namespace")
	indexCmd.Flags().BoolVarP(&forced, "force-ns", "fns", false, "Force the indexer to set specified namespace")

	rootCmd.AddCommand(serveCmd, indexCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
	db.Close()
}