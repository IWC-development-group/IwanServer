package main

import (
    "fmt"
	"os"
    "io/fs"
	"strings"
    "regexp"
    "path/filepath"

	"github.com/PuerkitoBio/goquery"
    "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
    "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

func FixTags(htmlContent *string) {
    re := regexp.MustCompile(`<script([^>]*?)/>`)
    *htmlContent = re.ReplaceAllString(*htmlContent, `<script$1></script>`)
    
    re = regexp.MustCompile(`<style([^>]*?)/>`)
    *htmlContent = re.ReplaceAllString(*htmlContent, `<style$1></style>`)
}

func CleanHtml(htmlContent string) (string, error) {
    FixTags(&htmlContent)
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
    if err != nil {
        return "", err
    }
    
    doc.Find("head").Remove()
    
    cleaned, err := doc.Html()
    if err != nil {
        return "", err
    }
    
    return cleaned, nil
}

func HtmlToMarkdown(conv *converter.Converter, source string, destination string) error {
    content, err := os.ReadFile(source)
    if err != nil {
        return err
    }

    cleaned, err := CleanHtml(string(content))
    //fmt.Println(cleaned)
    if err != nil {
        return err
    }

    markdown, err := conv.ConvertString(cleaned)
    if err != nil {
        return err
    }

    //fmt.Println(markdown)
    return os.WriteFile(destination, []byte(markdown), 0644)
}

func IsHtml(filename string) bool {
    extension := strings.ToLower(filepath.Ext(filename))
    return extension == ".html" || extension == ".xhtml"
}

func ProcessPages(srcRoot string, destRoot string) (int, error) {
    conv := converter.NewConverter(
        converter.WithPlugins(
            base.NewBasePlugin(),
            commonmark.NewCommonmarkPlugin(),
            table.NewTablePlugin(),
        ),
    )
    processedCount := 0

    err := filepath.Walk(srcRoot, func(path string, info fs.FileInfo, err error) error {
        if err != nil { return nil }
        if info.IsDir() || !IsHtml(path) { return nil }

        relPath, err := filepath.Rel(srcRoot, path)
        if err != nil {
            return err
        }

        relPath = strings.TrimSuffix(relPath, filepath.Ext(relPath)) + ".md"
        destPath := filepath.Join(destRoot, relPath)

        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            return err
        }

        fmt.Printf("Converting: %s to %s\n", path, destPath)
        if err := HtmlToMarkdown(conv, path, destPath); err != nil {
            return err
        }

        processedCount++
        return nil
    })

    if err != nil {
        return 0, err
    }

    return processedCount, nil
}

func main() {
    if len(os.Args) > 3 {
        fmt.Println("No such args!")
        os.Exit(0)
    }

    source := os.Args[1]
    destination := os.Args[2]
    fmt.Println(source + " -> " + destination)

    processedCount, err := ProcessPages(source, destination)
    if err != nil {
        panic(err)
        os.Exit(0)
    }

    fmt.Printf("Converted! Processed: %d\n", processedCount)
}