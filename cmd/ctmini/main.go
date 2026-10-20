// ctmini holds a cli interface for ct that doesn't depend on .ct files (namely org.ct and tex.ct)

package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "github.com/tnustrings/ct" 
)

// main kicks off tangling or tex generating
func main() {
    fname := os.Args[1]
    //b, _ := os.ReadFile(fname)
    //ct.Ctwrite(string(b), "", fname)
    // for ct files run Ctwrite
    if filepath.Ext(fname) == ".ct" {
        ct.Ctwrite(fname) 
    } else { // for non ct files run tc and print result for now.
        // get the genfiles
        _, genfiles, err := ct.Filesfromgen(fname)
        if err != nil { log.Fatal(err) }
        // run tc and print result
        out, err := ct.Tc(genfiles)
        if err != nil { log.Fatal(err) }
        fmt.Println(out)
    }
}