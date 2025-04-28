package fc
import (
    "cmp"
    "slices"
    "os"
    "path/filepath"
    "strings"
    "log"
)
func Keys[K cmp.Ordered, V any] (m map[K]V) []K {
    var names []K
    for name, _ := range m {
        names = append(names, name)
    }
    slices.Sort(names)
    return names
}
func IsDir(path string) bool {
    file, e := os.Open(path)
    if e != nil { return false } // maybe error cause file doesn't exist
    defer file.Close()
    fileinfo, e := file.Stat()
    Handle(e)
    return fileinfo.IsDir()
}
func Dir(path string) string {
    if IsDir(path) {
        return path
    } else {
        return filepath.Dir(path)
    }
}
func Stem(path string) string {
    base := filepath.Base(path)
    ext := filepath.Ext(base)
    return strings.TrimSuffix(base, ext)
}
func Handle(err error) {
    if err != nil {
        log.Fatal(err)
    }
}
