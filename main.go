package main

import (
    "crypto/sha1"
    "fmt"
    "io"
    "io/ioutil"
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

func (f* file_info) DstDir(setting env_setting) string {
    return setting.root + "/" + f.month
}

func (f* file_info) DstPath(setting env_setting) string {
    return f.DstDir(setting) + "/" + f.date + "-" + f.hash + f.ext
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

func (f* file_info) MoveFile(setting env_setting, done_files map[string]*file_info) {

    done, found:= done_files[f.hash]

    if found {
        if f.os_info.ModTime().Before(done.os_info.ModTime()) {
            log.Println("removing done file:", done.path, "newer than todo file:", f.path)
            if err := os.Remove(f.path); err != nil {
                log.Fatal("failed, ", err)
            }
            //todo file to be moved

        } else {
            log.Println("removing todo file:", f.path, "newer than done file:", done.path)
            if err := os.Remove(f.path); err != nil {
                log.Fatal("failed, ", err)
            }
            //file in place
            return
        }
    }

    dir := f.DstDir(setting)
    dst := f.DstPath(setting)

    if dst == f.path {
        log.Println("skip:", f.path)
        done_files[f.hash] = f
        return
    }

    _, err := os.Lstat(dst)
    //do not overwrite file, move to tmp name
    if err == nil {
        tmpfile, err := ioutil.TempFile(setting.root, "*"+f.ext)
        if err != nil {
            log.Fatal("fail creating temp file, error", err)
            return
        }
        dst = tmpfile.Name()
        tmpfile.Close()
        log.Println("moving", f.path, "to", dst)
        err = os.Rename(f.path, dst)
        if err != nil {
            log.Fatal("failed", err)
            return
        }
        return
    }

    err = os.MkdirAll(dir, 0755)
    if err != nil {
        log.Fatal("fail create dir: ", dir, "error:", err)
        return
    }

    log.Println("moving", f.path, "to", dst)
    err = os.Rename(f.path, dst)
    if err != nil {
        log.Fatal("failed", err)
        return
    }
    f.path = dst
    done_files[f.hash] = f

}

func GuessExt(path string) string {
    path = strings.ToLower(path)
    path = strings.TrimRight(path, "~")
    ext := filepath.Ext(path)
    if ext == "" {
        return ext
    }
    if strings.HasPrefix(ext, ".~") == false {
        return ext
    }
    return GuessExt(strings.TrimSuffix(path, ext))
}

func NewFileInfo(info os.FileInfo, path string) *file_info {

    f := file_info{os_info: info, path: path}

    f.month = info.ModTime().Format("2006-01")
    f.date = info.ModTime().Format("02")
    f.ext = GuessExt(path)

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
        todo_files[i].MoveFile(setting, done_files)
    }

}
