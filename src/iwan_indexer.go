package main
import (
    "fmt"
    "strings"
    "io/fs"
    "path/filepath"
    "os"

    "database/sql"
    _ "github.com/ncruces/go-sqlite3/embed"
    _ "github.com/ncruces/go-sqlite3/driver"
)

/* Do not pass a full path here! */
func GetPureName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

type IwanPage struct {
	Name string
	Namespace string
	Path string
}

func (page *IwanPage) SetupInfo(path string, fileInfo fs.FileInfo, namespace string) {
	page.Name = GetPureName(fileInfo.Name())
	page.Namespace = namespace
	page.Path = path
}

func (page *IwanPage) GetFullName() string {
	return page.Namespace + "/" + page.Name
}

func IsMarkdown(filename string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	return extension == ".md" || extension == ".markdown" || extension == ".mdown"
}

func IsIndexHint(filename string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	return extension == ".iwan"
}

func InitPagesTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS Pages(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			namespace TEXT NOT NULL,
			path TEXT NOT NULL
		)`)

	if err != nil {
		panic(err)
	}
}

func IsPageExists(db *sql.DB, path string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Pages WHERE path = ?", path).Scan(&count)

	if err != nil {
		return false, err
	} else {
		return count > 0, nil
	}
}

func CreateIndex(db *sql.DB, pageInfo *IwanPage) {
	_, err := db.Exec(`
		INSERT INTO Pages (name, namespace, path)
		VALUES ($1, $2, $3)
	`, pageInfo.Name, pageInfo.Namespace, pageInfo.Path)

	if err != nil {
		fmt.Printf("Can't create index for page %s\n", pageInfo.Path)
		panic(err)
	}

	fmt.Printf("Page \"%s\" added!\n", pageInfo.GetFullName())
}

func ProcessPages(db *sql.DB, root string, namespace string) (int, int, error) {
	processedCount := 0
	createdCount := 0

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil { return nil }

		if !info.IsDir() && IsMarkdown(path) {
			page := &IwanPage{}
			page.SetupInfo(path, info, namespace)
			
			exists, err := IsPageExists(db, page.Path)
			if err != nil {
				panic(err) 
			}
			if !exists { 
				CreateIndex(db, page)
				createdCount++
			}

			processedCount++
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	} else {
		return processedCount, createdCount, nil
	}
}

func IndexerMain(db *sql.DB, argOffset int) {
	if len(os.Args) - 1 < argOffset + 1 {
		fmt.Println("No such args!")
		os.Exit(0)
	}

	root, absErr := filepath.Abs(os.Args[argOffset + 1])
	if absErr != nil { panic(absErr) }

	var namespace string
	if len(os.Args) - 1 >= argOffset + 2 {
		namespace = os.Args[argOffset + 2]
	} else {
		namespace = "global"
		fmt.Println("No namespace specified so set it to global.")
	}

	InitPagesTable(db)
	processedCount, createdCount, err := ProcessPages(db, root, namespace)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Printf("Processed: %d\nCreated: %d\n", processedCount, createdCount)
}