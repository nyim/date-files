package main

import (
    "crypto/sha1"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "strings"
)

type env_setting struct {
    root string
}

type file_info struct {
    path string
    os_info os.FileInfo
    month string
    date string
    hash string
    ext string
}

func (f* file_info) Hash() string {
    fd, err := os.Open(f.path)
    if err != nil {
        log.Fatal(err)
        return "";
    }
    defer fd.Close()
    h := sha1.New()
    _, err = io.Copy(h, fd)
    if err != nil {
        log.Fatal(err)
    }
    f.hash = fmt.Sprintf("%x", h.Sum(nil))
    return f.hash
}

func NewFileInfo(info os.FileInfo, path string) *file_info {

    f := file_info{os_info: info, path: path}

    f.month = info.ModTime().Format("2006-01")
    f.date = info.ModTime().Format("02")
    f.ext = strings.ToLower(filepath.Ext(path))

    log.Println("hashing", path)

    f.Hash()

    return &f
}

func (s* env_setting) init () error {
    path, err := os.Getwd()
    if err != nil {
        log.Println(err)
        return err
    }
    s.root = path
    return nil
}

func FindTodoFiles(setting env_setting) ([]file_info, error) {

    list := make([]file_info, 0)

    err := filepath.Walk(setting.root, func (path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Println(err)
            return err
        }
        if !info.IsDir() {
            list = append(list, *NewFileInfo(info, path))
        }
        return nil
    })

    if err != nil {
        log.Println(err)
        return nil, err
    }

    return list, nil
}

func MoveFile(setting env_setting, n *file_info, done_files map[string]*file_info) {

    old, found:= done_files[n.hash]
    if found {
        if n.os_info.ModTime().After(old.os_info.ModTime()) || (n.month == old.month && n.date == old.date){
            log.Println("removing newer file:", n.path)
            if err := os.Remove(n.path); err != nil {
                log.Fatal("failed, ", err)
            }
            return
        }
    }

    dir := setting.root + "/" + n.month
    dst := dir + "/" + n.date + "-" + n.hash + n.ext

    if dst == n.path {
        log.Println("skip:", n.path)
        done_files[n.hash] = n
        return
    }

    err := os.MkdirAll(dir, 0755)
    if err != nil {
        log.Fatal("fail create dir: ", dir, "error:", err)
        return
    }

    log.Println("moving %s to %s", n.path, dst)
    err = os.Rename(n.path, dst)
    if err != nil {
        log.Fatal("failed", err)
        return
    }
    n.path = dst
    done_files[n.hash] = n

    if found {
        log.Println("removing exist file:", old.path)
        if err := os.Remove(old.path); err != nil {
            log.Fatal("failed, ", err)
        }
    }
}

func main() {

    setting := env_setting{}
    err := setting.init()
    if err != nil {
        return
    }

    log.Println("listing files in:", setting.root)

    todo_files, err := FindTodoFiles(setting)
    if err != nil {
        return
    }

    done_files := make(map[string]*file_info)

    for i := 0; i < len(todo_files); i++ {
        f := &todo_files[i]
        MoveFile(setting, f, done_files)
    }

}
