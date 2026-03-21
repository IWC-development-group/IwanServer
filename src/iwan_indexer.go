package main
import (
    "fmt"
    "strings"
    "io/fs"
    "bufio"
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

type IwanIndexHint struct {
	Dir string
	Namespace string
}

func IsMarkdown(filename string) bool {
	extension := strings.ToLower(filepath.Ext(filename))
	return extension == ".md" || extension == ".markdown" || extension == ".mdown"
}

func GetIndexHint(dir string) IwanIndexHint {
	path := dir + "/hint.iwan"
	_, err := os.Stat(path)

	if err != nil { return IwanIndexHint{"", "",} }

	file, _ := os.Open(path)

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() { return IwanIndexHint{"", "",} }

	words := strings.Fields(scanner.Text())
	file.Close()

	var hint IwanIndexHint
	hint.Dir = dir
	hint.Namespace = words[0]
	return hint
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

func ProcessPages(db *sql.DB, root string, namespace string, forced bool) (int, int, error) {
	processedCount := 0
	createdCount := 0

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil { return nil }

		if !forced && info.IsDir() {
			hint := GetIndexHint(path)
			if hint.Namespace != "" { return nil }
		}

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

	if err != nil { return 0, 0, err }
	return processedCount, createdCount, nil
}

func CollectHints(root string) ([]IwanIndexHint, error) {
	var hints []IwanIndexHint

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() || err != nil { return nil }

		hint := GetIndexHint(path)
		if hint.Namespace != "" {
			hints = append(hints, hint)
		}

		return nil
	})

	return hints, err
}

func RunIndexing(db *sql.DB, root string, namespace string) (int, int, error) {
	hints, err := CollectHints(root)
	if err != nil { return 0, 0, err }
	fmt.Printf("Hints collected: %d\n", len(hints))

	totalProcessed := 0
	totalCreated := 0

	for _, hint := range hints {
		processedCount, createdCount, err := ProcessPages(db, hint.Dir, namespace, false)
		if err != nil { return 0, 0, err }

		totalProcessed += processedCount
		totalCreated += createdCount
	}

	return totalProcessed, totalCreated, nil
}

func IndexerMain(db *sql.DB, path string, namespace string, forced bool) {
	root, absErr := filepath.Abs(path)
	if absErr != nil { panic(absErr) }

	InitPagesTable(db)
	processedCount, createdCount, err := ProcessPages(db, root, namespace, true)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Printf("Processed: %d\nCreated: %d\n", processedCount, createdCount)
}