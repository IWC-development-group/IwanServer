package main

import (
	"os"
	"fmt"
	"path/filepath"
	"github.com/spf13/cobra"

	"database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

type ModuleFunc func(db *sql.DB, argOffset int)

var version = ""

func main() {
	rootCmd := &cobra.Command{
		Use: "iwans",
		Version: version,
		Short: "Native server for IwanClient to store Markdown manuals.",
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
	indexCmd.Flags().StringVarP(&namespace, "namespace", "n", "global", "Documentation namespace")
	indexCmd.Flags().BoolVarP(&forced, "force-ns", "f", false, "Force the indexer to ignore all hints")

	var namespaceDeletion string
	deleteCmd := &cobra.Command{
		Use: "delete",
		Short: "Delete the specified page or namespace",
		Run: func(cmd* cobra.Command, args []string){
			if namespaceDeletion != "" {
				DeleteNamespace(db, namespaceDeletion)
				return
			}
			
			if len(args) == 0 {
				fmt.Println("No page specified!")
				return
			}

			DeletePage(db, args[0])
		},
	}
	deleteCmd.Flags().StringVarP(&namespaceDeletion, "namespace", "n", "", "Delete the specified namespace. If set, page name will be ignored"

	relayCmd := &cobra.Command{
		Use: "relay <url> <namespace/name>",
		Short: "Add a page that will be relayed from HTTP-server or another IwanServer",
		Run: func(cmd* cobra.Command, args []srting){
			if len(args) < 2 {
				fmt.Println("No such args!")
				return
			}

			AddRelay(db, args[0], args[1])
		}
	}

	rootCmd.AddCommand(serveCmd, indexCmd, deleteCmd, relayCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
	db.Close()
}