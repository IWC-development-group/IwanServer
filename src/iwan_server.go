package main

import (
	"os"
	"fmt"
//	"bytes"
	"encoding/json"
	"net/http"
//	"net/url"
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

func PageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response := IwanResponse {
		Status: "ERR",
		Name: "none",
		Namespace: "global",
		Content: "none",
	}

	pageName := r.URL.Query().Get("name")
	fmt.Println("Client requested " + pageName)

	if pageName == "" {
		jsonReq := SetErrorDescription(&response, "Page is unspecified!")
		w.Write(jsonReq)
		return
	}

	response.Name = pageName

	content, err := os.ReadFile(pageName)
	if err != nil {
		jsonReq := SetErrorDescription(&response, "Page not found!")
		w.Write(jsonReq)
		return
	}

	response.Status = "OK"
	response.Content = string(content)

	jsonReq, err := json.Marshal(response)
	if err != nil { panic(err.Error()) }
	fmt.Fprintf(w, string(jsonReq))
}

func ServerMain(argOffset int) {
	port := os.Args[argOffset + 1]
	http.HandleFunc("/", PageHandler)
	http.ListenAndServe(":" + port, nil)
}