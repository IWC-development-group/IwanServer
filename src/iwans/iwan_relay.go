package main

import (
    "fmt"
    "net/url"
    "net/http"
    "errors"
    "encoding/json"

    "database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

var ErrCantAccess = errors.New("Can't access relayed page!")
var ErrFailedToRead = errors.New("Failed to read response body!")
var ErrNotIwanApi = errors.New("Relayed response does not correspond the Iwan API format!")

type PosJsonResponse struct {
	Content string 		`json:"content"`
}

func IsUrl(line string) bool {
	u, err := url.Parse(line)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

func IsJson(contentType string) {
    return strings.Contains(contentType, "json")
}

/* Extracts content from JSON-formated string */
func ExcractContent(content string) (string, error) {
	var result PosJsonResponse
	err := json.Unmarshal(content, &result)
	if err != nil {
		return "", ErrNotIwanApi
	}

	return result.Content
}

func readFromUrl(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, ErrCantAccess
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, ErrCantAccess
    }

    content, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, ErrFailedToRead
    }

    strContent := string(content)
    if IsJson(resp.Header.Get("Content-Type")) {
    	return ExcractContent(strContent)
    }

    return strContent, nil
}

func AddRelay(db *sql.DB, urlLine string, pageFullName string) {
	if !IsUrl(urlLine) {
		fmt.Println("Not a valid url!")
		return
	}

	var page IwanPage
	page.SetupInfoFromFullName(urlLine, pageFullName)
	_, err := db.Exec("INSERT INTO Pages (name, namespace, path) VALUES(?, ?, ?)",
		page.Name, page.Namespace, page.Path)

	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Relay \"%s\" added!", urlLine)
}