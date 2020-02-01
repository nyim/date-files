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

type fileInfo struct {
    path string
    modTime time.Time
    ext string
    hash string
}


type processor struct {
    rootDir string
    tmpDir string
    todoFiles []*fileInfo
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
    p.rootDir = path
    log.Println("root directory is", p.rootDir)
    return nil
}

func (p* processor) scanTodos() error {

    err := filepath.Walk(p.rootDir, func (path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
            return err
        }
        if !info.IsDir() {
            p.todoFiles = append(p.todoFiles, newFileInfo(path, info))
        }
        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    return err
}

func (p* processor) removeDuplicates() error {

    infoMap := make(map[string]*fileInfo)

    var err error

    for _, todo := range p.todoFiles {
        done, found := infoMap[todo.hash]

        if !found {
            infoMap[todo.hash] = todo
            continue
        }

        if todo.modTime.Before(done.modTime) {
            err = done.remove()
        } else {
            err =todo.remove()
        }

        if err != nil {
            return err
        }
    }

    p.todoFiles = nil

    for _, todo := range infoMap {
        p.todoFiles = append(p.todoFiles, todo)
    }

    return nil
}

func (p* processor) filterDones() error {
    all := p.todoFiles
    p.todoFiles = nil

    for _, todo := range all {
        if todo.path != todo.dstPath(p.rootDir) {
            p.todoFiles = append(p.todoFiles, todo)
        }
    }

    return nil
}

func (p* processor) moveToTmpdir() error {

    if len(p.todoFiles) == 0 {
        return nil
    }

    var err error

    p.tmpDir, err = ioutil.TempDir(p.rootDir, "tmp_")
    log.Println("temp dir is", p.tmpDir)
    if err != nil {
        log.Fatal("fail create temp dir", err)
        return err
    }

    for _, todo := range p.todoFiles {
        dstPath := fmt.Sprintf("%s/%s%s", p.tmpDir, todo.hash, todo.ext)
        err = todo.moveTo(dstPath)
        if err != nil {
            return err
        }
    }

    return nil
}

func (p* processor) store() error {
    for _, todo := range p.todoFiles {
        err := todo.moveTo(todo.dstPath(p.rootDir))
        if err != nil {
            return err
        }
    }
    return nil
}

func (p* processor) clear() error {
    if p.tmpDir != "" {
        err := os.Remove(p.tmpDir)
        if err != nil {
            log.Printf("failed remove tmp dir %v error %v\n", p.tmpDir, err)
        }
    }
    return nil
}

func (f* fileInfo) remove() error {
    err := os.Remove(f.path)
    if err != nil {
        log.Fatal("failed: ", err)
    }
    return err
}

func (f* fileInfo) moveTo(dstPath string) error {
    log.Printf("moving %v to %v\n", f.path, dstPath)

    err := os.Rename(f.path, dstPath)
    if err != nil && os.IsNotExist(err) {
        err = os.MkdirAll(filepath.Dir(dstPath), 0755)
    }
    if err != nil {
        log.Fatal("failed: ", err)
        return err
    }
    f.path = dstPath
    return nil
}


func (f* fileInfo) dstDir(rootDir string) string {
    return fmt.Sprintf("%v/%v",
        rootDir, f.modTime.Format("2006-01"))
}

func (f* fileInfo) dstPath(rootDir string) string {
    return fmt.Sprintf("%v/%v-%v%v",
        f.dstDir(rootDir), f.modTime.Format("02"), f.hash, f.ext)
}

func (f* fileInfo) Hash() string {
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

func guessExt(path string) string {
    path = strings.ToLower(path)
    path = strings.TrimRight(path, "~")
    ext := filepath.Ext(path)
    if ext == "" {
        return ext
    }
    if strings.HasPrefix(ext, ".~") == false {
        return ext
    }
    return guessExt(strings.TrimSuffix(path, ext))
}

func newFileInfo(path string, info os.FileInfo) *fileInfo {

    f := fileInfo{modTime: info.ModTime(), path: path}
    f.ext = guessExt(path)

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
