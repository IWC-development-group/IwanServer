package main
import (
    "strings"
    "io/fs"
    "os"
)

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

func (page *IwanPage) GetContent() ([]byte, error) {
    content, err := os.ReadFile(page.Path)
    if err != nil {
        return nil, err
    }

    return content, nil
}