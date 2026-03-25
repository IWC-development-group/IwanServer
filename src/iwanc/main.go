package main

import (
    "fmt"
	"os"
    "io/fs"
	"strings"
    "regexp"
    "path/filepath"

    "github.com/spf13/cobra"
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
    if err != nil {
        return err
    }

    markdown, err := conv.ConvertString(cleaned)
    if err != nil {
        return err
    }

    return os.WriteFile(destination, []byte(markdown), 0644)
}

func IsHtml(filename string) bool {
    extension := strings.ToLower(filepath.Ext(filename))
    return extension == ".html" || extension == ".xhtml"
}

func ProcessPages(srcRoot string, destRoot string, verboseFlag bool) (int, error) {
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

        if verboseFlag {
            fmt.Printf("Converting: %s to %s\n", path, destPath)
        } else {
            fmt.Printf("\rProcessing: %d", processedCount + 1)
        }

        if err := HtmlToMarkdown(conv, path, destPath); err != nil {
            return err
        }

        processedCount++
        return nil
    })
    if !verboseFlag { fmt.Println("") }

    if err != nil {
        return 0, err
    }

    return processedCount, nil
}

func RunConverter(source string, destination string, verboseFlag bool) {
    fmt.Println(source + " -> " + destination)

    processedCount, err := ProcessPages(source, destination, verboseFlag)
    if err != nil {
        panic(err)
        os.Exit(0)
    }

    fmt.Printf("Convertation completed! Processed: %d\n", processedCount)
}

func main() {
    var verbose bool
    rootCmd := &cobra.Command{
        Use: "iwanc <source> <destination>",
        Short: "Converter utility for converting HTML pages to Markdown",
        Args: cobra.ExactArgs(2),

        Run: func(cmd *cobra.Command, args []string){
            RunConverter(args[0], args[1], verbose)
        },
    }
    rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

    if err := rootCmd.Execute(); err != nil {
        panic(err)
    }
}