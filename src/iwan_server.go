package main

import (
	"os"
	"fmt"
//	"bytes"
	"encoding/json"
	"net/http"
//	"net/url"

    "database/sql"
)

type IwanResponse struct {
	Status string 		`json:"status"`
	Name string 		`json:"name"`
	Namespace string 	`json:"namespace"`
	Content string 		`json:"content"`
}

func SetErrorDescription(response *IwanResponse, message string) []byte {
	response.Content = message
	jsonReq, err := json.Marshal(response)
	if err != nil { panic(err.Error()) }
	return jsonReq
}

func GetPagePath(db *sql.DB, page *IwanPage) (string, error) {
	var path string
	err := db.QueryRow("SELECT path FROM Pages WHERE name = ? AND namespace = ?",
		page.Name, page.Namespace).Scan(&path)

	fmt.Printf("Using name: %s and namespace %s. Found: %s\n", page.Name, page.Namespace, path)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return path, nil
}

func PageHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := IwanResponse {
		Status: "ERR",
		Name: "none",
		Namespace: "global",
		Content: "none",
	}

	pageFullName := r.URL.Query().Get("name")
	fmt.Println("Client requested " + pageFullName)

	if pageFullName == "" {
		jsonReq := SetErrorDescription(&response, "Page is unspecified!")
		w.Write(jsonReq)
		return
	}

	var page IwanPage
	page.SetupInfoFromFullName("", pageFullName)
	path, pathErr := GetPagePath(db, &page)
	page.Path = path

	response.Name = page.Name
	response.Namespace = page.Namespace

	if pathErr != nil {
		jsonReq := SetErrorDescription(&response, "Something went wrong when searching for the page.")
		w.Write(jsonReq)
		panic(pathErr)
		return
	} else if page.Path == "" {
		jsonReq := SetErrorDescription(&response, "Page not found!")
		w.Write(jsonReq)
		return
	}

	content, err := page.GetContent()
	if err != nil {
		jsonReq := SetErrorDescription(&response, "Page indexed but not exists!")
		w.Write(jsonReq)
		return
	}

	response.Status = "OK"
	response.Content = string(content)

	jsonReq, err := json.Marshal(response)
	if err != nil { panic(err.Error()) }
	fmt.Fprintf(w, string(jsonReq))
}

func ServerMain(db *sql.DB, argOffset int) {
	port := os.Args[argOffset + 1]

	fmt.Println("Started!")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		PageHandler(db, w, r)
	})
	http.ListenAndServe(":" + port, nil)
}