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

func (hint *IwanIndexHint) ToString() string {
	return "[" + hint.Namespace + ": " + hint.Dir + "]"
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

func CreateMultipleIndex(db *sql.DB, pages *[]IwanPage) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO Pages (name, namespace, path)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		panic(err)
		return
	}
	defer stmt.Close()

	for _, page := range *pages {
		_, err := stmt.Exec(page.Name, page.Namespace, page.Path)
		if err != nil {
			panic(err)
			return
		}
	}

	txErr := tx.Commit()
	if txErr != nil {
		panic(txErr)
	}
}

func ProcessPages(db *sql.DB, root string, namespace string, forced bool) (int, int, error) {
	processedCount := 0
	createdCount := 0
	firstRoot := true
	var pages []IwanPage

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil { return nil }

		if !forced && info.IsDir() {
			hint := GetIndexHint(path)

			if firstRoot {
				firstRoot = false
			} else if hint.Namespace != "" {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() && IsMarkdown(path) {
			page := &IwanPage{}
			page.SetupInfo(path, info, namespace)
			
			exists, err := IsPageExists(db, page.Path)
			if err != nil {
				panic(err) 
			}
			if !exists { 
				pages = append(pages, *page)
				createdCount++
				fmt.Printf("\rCollecting pages: %d", createdCount)
			}

			processedCount++
		}

		return nil
	})
	fmt.Println("")

	if createdCount != 0 {
		CreateMultipleIndex(db, &pages)
	}

	if err != nil { return 0, 0, err }
	return processedCount, createdCount, nil
}

func CollectHints(hints *[]IwanIndexHint, root string) error {
	isRoot := true
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() || err != nil { return nil }

		hint := GetIndexHint(path)
		if hint.Namespace != "" {
			if isRoot {
				(*hints)[0] = hint
				isRoot = false
			} else {
				*hints = append(*hints, hint)
			}
		}

		return nil
	})

	return err
}

func RunIndexing(db *sql.DB, root string, namespace string) (int, int, error) {
	var hints []IwanIndexHint
	hints = append(hints, IwanIndexHint{root, namespace})
	err := CollectHints(&hints, root)
	
	if err != nil { return 0, 0, err }
	fmt.Printf("Hints collected: %d\n", len(hints))

	if len(hints) <= 1 {
		processedCount, createdCount, err := ProcessPages(db, root, hints[0].Namespace, true)
		if err != nil { return 0, 0, err }
		return processedCount, createdCount, nil
	}

	totalProcessed := 0
	totalCreated := 0

	for _, hint := range hints {
		fmt.Printf("Using hint: %s\n", hint.ToString())
		processedCount, createdCount, err := ProcessPages(db, hint.Dir, hint.Namespace, false)
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

	var processedCount int
	var createdCount int
	var err error

	if !forced {
		processedCount, createdCount, err = RunIndexing(db, root, namespace)
	} else {
		processedCount, createdCount, err = ProcessPages(db, root, namespace, true)
	}

	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Printf("Indexing completed. Processed: %d, created: %d\n", processedCount, createdCount)
}