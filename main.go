package main

import (
    "crypto/sha1"
    "fmt"
    "io"
    //"io/ioutil"
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
    return nil
}

func (p* processor) filterDones() error {
    return nil
}

func (p* processor) moveToTmpdir() error {
    return nil
}

func (p* processor) store() error {
    return nil
}

func (p* processor) clear() error {
    return nil
}

func (f* file_info) DstDir(root_dir string) string {
    return fmt.Sprintf("%q/%q",
        root_dir, f.mod_time.Format("2006-01"))
}

func (f* file_info) DstPath(root_dir string) string {
    return fmt.Sprintf("%q/%q-%q%q",
        f.DstDir(root_dir), f.mod_time.Format("02"), f.hash, f.ext)
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


//func (f* file_info) MoveFile(setting env_setting, done_files map[string]*file_info) {
//
//    done, found:= done_files[f.hash]
//
//    if found {
//        if f.os_info.ModTime().Before(done.os_info.ModTime()) {
//            log.Println("removing done file:", done.path, "newer than todo file:", f.path)
//            if err := os.Remove(f.path); err != nil {
//                log.Fatal("failed, ", err)
//            }
//            //todo file to be moved
//
//        } else {
//            log.Println("removing todo file:", f.path, "newer than done file:", done.path)
//            if err := os.Remove(f.path); err != nil {
//                log.Fatal("failed, ", err)
//            }
//            //file in place
//            return
//        }
//    }
//
//    dir := f.DstDir(setting)
//    dst := f.DstPath(setting)
//
//    if dst == f.path {
//        log.Println("skip:", f.path)
//        done_files[f.hash] = f
//        return
//    }
//
//    _, err := os.Lstat(dst)
//    //do not overwrite file, move to tmp name
//    if err == nil {
//        tmpfile, err := ioutil.TempFile(setting.root, "*"+f.ext)
//        if err != nil {
//            log.Fatal("fail creating temp file, error", err)
//            return
//        }
//        dst = tmpfile.Name()
//        tmpfile.Close()
//        log.Println("moving", f.path, "to", dst)
//        err = os.Rename(f.path, dst)
//        if err != nil {
//            log.Fatal("failed", err)
//            return
//        }
//        return
//    }
//
//    err = os.MkdirAll(dir, 0755)
//    if err != nil {
//        log.Fatal("fail create dir: ", dir, "error:", err)
//        return
//    }
//
//    log.Println("moving", f.path, "to", dst)
//    err = os.Rename(f.path, dst)
//    if err != nil {
//        log.Fatal("failed", err)
//        return
//    }
//    f.path = dst
//    done_files[f.hash] = f
//
//}

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
