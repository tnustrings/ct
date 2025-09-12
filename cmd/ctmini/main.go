// ctmini holds a cli interface for ct that doesn't depend on .ct files (namely org.ct and tex.ct)

package main

import (
    "os" 
    "github.com/tnustrings/ct" 
)

// main kicks off tangling or tex generating
func main() {
    fname := os.Args[1]
    b, _ := os.ReadFile(fname)
    ct.Ctwrite(string(b), "")
}