package main

import (
    "io/ioutil"
    "os"
    "time"
    "testing"
)

func setUp(t *testing.T) func(t *testing.T) {
    rootDir, err := os.Getwd()
    if err != nil {
        t.Error("fail Getwd", err)
    }
    wd, err := ioutil.TempDir("test/data", "tmp_")
    if err != nil {
        t.Error("fail create temp dir", err)
    }
    err = os.Chdir(wd)
    if err != nil {
        t.Error("fail entering temp dir", err)
        return nil
    }
    return func(t *testing.T) {
        err := os.Chdir(rootDir)
        if err != nil {
            t.Error("fail backing to root dir", rootDir)
            return
        }
        err = os.RemoveAll(wd)
        if err != nil {
            t.Error("fail clean temp dir", rootDir)
            return
        }
    }
}

func TestKeepEmptyExt(t *testing.T) {
    tearDown := setUp(t)
    defer tearDown(t)
    bytes := []byte{'a'}

    err := ioutil.WriteFile("abc", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("abc", mtime, mtime)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    p := processor{}
    p.Process()

    _, err = os.Lstat("abc")
    if err != nil {
        t.Error("abc without ext name should not be moved")
    }
}

func TestRemoveDone(t *testing.T) {
    tearDown := setUp(t)
    defer tearDown(t)

    bytes := []byte{'a'}

    err := ioutil.WriteFile("abc.jpg", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("abc.jpg", mtime, mtime)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    err = ioutil.WriteFile("aaa.jpg", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    mtime = time.Date(2007, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("aaa.jpg", mtime, mtime)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    p := processor{}
    p.Process()

    _, err = os.Lstat("2006-02/01-86f7e437faa5a7fce15d1ddcb9eaeaea377667b8.jpg")
    if err != nil {
        t.Error("fail move to", err)
    }

    _, err = os.Lstat("abc.jpg")
    if err == nil {
        t.Error("fail move, abc.jpg remains")
    }
}

func TestOneFile(t *testing.T) {
    tearDown := setUp(t)
    defer tearDown(t)

    bytes := []byte{'a'}

    err := ioutil.WriteFile("abc.jpg", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    time := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("abc.jpg", time, time)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    p := processor{}
    p.Process()

    _, err = os.Lstat("2006-02/01-86f7e437faa5a7fce15d1ddcb9eaeaea377667b8.jpg")
    if err != nil {
        t.Error("fail move to", err)
    }

    _, err = os.Lstat("abc.jpg")
    if err == nil {
        t.Error("fail move, abc.jpg remains")
    }
}

func TestRemoveTodo(t *testing.T) {
    tearDown := setUp(t)
    defer tearDown(t)

    bytes := []byte{'a'}

    err := ioutil.WriteFile("abc.jpg", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    mtime := time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("abc.jpg", mtime, mtime)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    err = ioutil.WriteFile("new.jpg", bytes, 0644)
    if err != nil {
        t.Error("fail write file", err)
    }

    mtime = time.Date(2007, time.February, 1, 3, 4, 5, 0, time.UTC)
    err = os.Chtimes("new.jpg", mtime, mtime)

    if err != nil {
        t.Error("fail Chtimes", err)
    }

    p := processor{}
    p.Process()

    _, err = os.Lstat("2006-02/01-86f7e437faa5a7fce15d1ddcb9eaeaea377667b8.jpg")
    if err != nil {
        t.Error("fail move to", err)
    }

    _, err = os.Lstat("abc.jpg")
    if err == nil {
        t.Error("fail move, abc.jpg remains")
    }
}

