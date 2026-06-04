package main
import (
    "strings"
    "io/fs"
    "os"
    "errors"
)

var ErrCantReadFromFile = errors.New("Can't read content from file!")

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

func (page *IwanPage) SetupInfoFromFullName(path string, fullName string) {
    components := strings.Split(fullName, "/")
    if len(components) < 2 {
        page.Name = fullName
        page.Namespace = "global"
        return
    }

    page.Name = components[1]
    page.Namespace = components[0]
    page.Path = path
}

func (page *IwanPage) GetFullName() string {
    return page.Namespace + "/" + page.Name
}

func readFromFile(path string) (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return "", ErrCantReadFromFile
    }
    return string(content), nil
}

func (page *IwanPage) GetContent() (string, error) {
    if IsUrl(page.Path) {
        return readFromUrl(page.Path)
    }
    return readFromFile(page.Path)
}