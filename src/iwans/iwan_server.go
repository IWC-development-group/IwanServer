package main

import (
//	"os"
	"fmt"
	"strconv"
//	"bytes"
	"encoding/json"
	"net/http"
//	"net/url"

    "database/sql"
)

func IwanSerialize(response any) []byte {
	jsonReq, err := json.Marshal(response)
	if err != nil { panic(err) }
	return jsonReq
}

type IwanPageResponse struct {
	Status string 		`json:"status"`
	Name string 		`json:"name"`
	Namespace string 	`json:"namespace"`
	Content string 		`json:"content"`
}

func (response *IwanPageResponse) SetErrorDescription(message string) []byte {
	response.Content = message
	return IwanSerialize(*response)
}

type IwanPageListResponse struct {
	Status string		`json:"status"`
	Namespace string	`json:"namespace"`
	Pages []string		`json:"pages"`
}

func (response *IwanPageListResponse) SetErrorDescription(message string) []byte {
	response.Pages[0] = message
	return IwanSerialize(*response)
}

type IwanNamespaceListResponse struct {
	Status string 		`json:"status"`
	Namespaces []string	`json:"namespaces"`
}

func (response *IwanNamespaceListResponse) SetErrorDescription(message string) []byte {
	response.Namespaces[0] = message
	return IwanSerialize(*response)
}

func GetPagePath(db *sql.DB, page *IwanPage) (string, int, error) {
	var path string
	var id int
	err := db.QueryRow("SELECT id, path FROM Pages WHERE name = ? AND namespace = ?",
		page.Name, page.Namespace).Scan(&id, &path)

	//fmt.Printf("Using name %s and namespace %s. Found: (%d) %s\n", page.Name, page.Namespace, id, path)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, err
	}

	return path, id, nil
}

func RemovePage(db *sql.DB, id int) {
	fmt.Println("Removing index for non-existent page")
	_, err := db.Exec(`DELETE FROM Pages WHERE id = ?`, id)
	if err != nil {
		panic(err)
	}
}

func GetPagesByNamespace(db *sql.DB, namespace string) ([]string, error) {
	rows, err := db.Query("SELECT name FROM Pages WHERE namespace = ?", namespace)
	if err != nil {
		defer rows.Close()
		return nil, err
	}

	var pages []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		pages = append(pages, name)
	}

	return pages, nil
}

func GetNamespaces(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT namespace FROM Pages")
	if err != nil {
		defer rows.Close()
		return nil, err
	}

	var namespaces []string
	for rows.Next() {
		var namespace string
		rows.Scan(&namespace)
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func PageHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := IwanPageResponse {
		Status: "ERR",
		Name: "none",
		Namespace: "global",
		Content: "none",
	}

	pageFullName := r.URL.Query().Get("name")
	fmt.Println("Client requested " + pageFullName)

	if pageFullName == "" {
		jsonReq := response.SetErrorDescription("Page is unspecified!")
		w.Write(jsonReq)
		return
	}

	var page IwanPage
	page.SetupInfoFromFullName("", pageFullName)
	path, id, pathErr := GetPagePath(db, &page)
	page.Path = path

	response.Name = page.Name
	response.Namespace = page.Namespace

	if pathErr != nil {
		jsonReq := response.SetErrorDescription("Something went wrong when searching for the page.")
		w.Write(jsonReq)
		panic(pathErr)
		return
	} else if page.Path == "" {
		jsonReq := response.SetErrorDescription("Page not found!")
		w.Write(jsonReq)
		return
	}

	content, err := page.GetContent()
	if err != nil {
		RemovePage(db, id)
		jsonReq := response.SetErrorDescription("Page indexed but not exists!")
		w.Write(jsonReq)
		return
	}

	response.Status = "OK"
	response.Content = string(content)

	jsonReq := IwanSerialize(response)
	fmt.Fprintf(w, string(jsonReq))
}

func PageListHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := IwanPageListResponse{
		Status: "ERR",
		Namespace: "global",
		Pages: []string{"none"},
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" { namespace = "global" }
	response.Namespace = namespace

	fmt.Printf("Client requested page list in %s\n", namespace)

	pages, err := GetPagesByNamespace(db, namespace)
	if err != nil {
		jsonReq := response.SetErrorDescription(fmt.Sprintf("Something went wrong when searching for pages in \"%s\".", namespace))
		w.Write(jsonReq)
		return
	}

	if len(pages) == 0 {
		jsonReq := response.SetErrorDescription(fmt.Sprintf("No pages found in namespace \"%s\"!", namespace))
		w.Write(jsonReq)
		return
	}

	response.Status = "OK"
	response.Pages = pages

	jsonReq := IwanSerialize(response)
	fmt.Fprintf(w, string(jsonReq))
}

func NamespaceListHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := IwanNamespaceListResponse{
		Status: "ERR",
		Namespaces: []string{"none"},
	}

	fmt.Println("Client requested namespace list")

	namespaces, err := GetNamespaces(db)
	if err != nil {
		jsonReq := response.SetErrorDescription("Something went wrong when searching for namespaces.")
		w.Write(jsonReq)
		return
	}

	if len(namespaces) == 0 {
		jsonReq := response.SetErrorDescription("No namespaces found on the server!")
		w.Write(jsonReq)
		return
	}

	response.Status = "OK"
	response.Namespaces = namespaces

	jsonReq := IwanSerialize(response)
	fmt.Fprintf(w, string(jsonReq))
}

func ServerMain(db *sql.DB, port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		PageHandler(db, w, r)
	})
	http.HandleFunc("/pages", func(w http.ResponseWriter, r *http.Request) {
		PageListHandler(db, w, r)
	})
	http.HandleFunc("/namespaces", func(w http.ResponseWriter, r *http.Request) {
		NamespaceListHandler(db, w, r)
	})

	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Serving on %s!\n", addr)
	http.ListenAndServe(addr, nil)
}