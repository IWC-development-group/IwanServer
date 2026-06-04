package main

import (
    "fmt"
    "net/url"

    "database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

func IsUrl(line string) bool {
	u, err := url.Parse(line)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

func AddRelay(db *sql.DB, urlLine string, pageFullName string) {
	if !IsUrl(urlLine) {
		fmt.Prinln("Not a valid url!")
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

	fmt.Println("Relay \"%s\" added!", urlLine)
}