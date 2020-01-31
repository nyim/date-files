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
    "time"
)

type file_info struct {
    path string
    mod_time time.Time
    ext string
    hash string
}


type processor struct {
    root_dir string
    tmp_dir string
    todo_files []*file_info
}

func (p* processor) Process() error {
    err := p.init()
    if err != nil {
        return err
    }
    err = p.scanTodos()
    if err != nil {
        return err
    }
    err = p.removeDuplicates()
    if err != nil {
        return err
    }
    err = p.filterDones()
    if err != nil {
        return err
    }
    err = p.moveToTmpdir()
    if err != nil {
        return err
    }
    err = p.store()
    if err != nil {
        return err
    }
    return p.clear()
}

func (p* processor) init () error {
    path, err := os.Getwd()
    if err != nil {
        log.Println(err)
        return err
    }
    p.root_dir = path
    log.Println("root directory is", p.root_dir)
    return nil
}

func (p* processor) scanTodos() error {

    err := filepath.Walk(p.root_dir, func (path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
            return err
        }
        if !info.IsDir() {
            p.todo_files = append(p.todo_files, NewFileInfo(path, info))
        }
        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    return err
}

func (p* processor) removeDuplicates() error {

    info_map := make(map[string]*file_info)

    var err error

    for _, todo := range p.todo_files {
        done, found := info_map[todo.hash]

        if !found {
            info_map[todo.hash] = todo
            continue
        }

        if todo.mod_time.Before(done.mod_time) {
            err = done.remove()
        } else {
            err =todo.remove()
        }

        if err != nil {
            return err
        }
    }

    p.todo_files = nil

    for _, todo := range info_map {
        p.todo_files = append(p.todo_files, todo)
    }

    return nil
}

func (p* processor) filterDones() error {
    all := p.todo_files
    p.todo_files = nil

    for _, todo := range all {
        if todo.path != todo.dstPath(p.root_dir) {
            p.todo_files = append(p.todo_files, todo)
        }
    }

    return nil
}

func (p* processor) moveToTmpdir() error {

    if len(p.todo_files) == 0 {
        return nil
    }

    var err error

    p.tmp_dir, err = ioutil.TempDir(p.root_dir, "")
    if err != nil {
        log.Fatal("fail create temp dir", err)
        return err
    }

    for _, todo := range p.todo_files {
        dst_path := fmt.Sprintf("%q/%q%q", p.tmp_dir, todo.hash, todo.ext)
        err = todo.moveTo(dst_path)
        if err != nil {
            return err
        }
    }

    return nil
}

func (p* processor) store() error {
    for _, todo := range p.todo_files {
        err := todo.moveTo(todo.dstPath(p.root_dir))
        if err != nil {
            return err
        }
    }
    return nil
}

func (p* processor) clear() error {
    if p.tmp_dir != "" {
        return os.Remove(p.tmp_dir)
    }
    return nil
}

func (f* file_info) remove() error {
    err := os.Remove(f.path)
    if err != nil {
        log.Fatal("failed: ", err)
    }
    return err
}

func (f* file_info) moveTo(dst_path string) error {
    err := os.Rename(f.path, dst_path)
    if err != nil {
        log.Fatal("failed", err)
        return err
    }
    f.path = dst_path
    return nil
}


func (f* file_info) dstDir(root_dir string) string {
    return fmt.Sprintf("%q/%q",
        root_dir, f.mod_time.Format("2006-01"))
}

func (f* file_info) dstPath(root_dir string) string {
    return fmt.Sprintf("%q/%q-%q%q",
        f.dstDir(root_dir), f.mod_time.Format("02"), f.hash, f.ext)
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

func NewFileInfo(path string, info os.FileInfo) *file_info {

    f := file_info{mod_time: info.ModTime(), path: path}
    f.ext = GuessExt(path)

    if info.Size() > 1024 * 1024 * 5 {
        log.Println("hashing", path)
    }

    f.Hash()

    return &f
}



func main() {

    p := processor{}
    p.Process()

}
