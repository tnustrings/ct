package main
import (
  "bufio"
  "flag"
  "fmt"
  "os"
  "log"
  "path"
  "path/filepath"
  "strconv"
  "strings"
  "github.com/tnustrings/ct"
  "github.com/tnustrings/ct/internal/fc"
)
func main() {
    fl_tex := flag.Bool("tex", false, "print doc as latex")
    fl_from_org := flag.Bool("from_org", false, "input is a .org file")
    fl_mdtotex := flag.String("mdtotex", "", "for latex doc generation. a command to convert markdown between codechunks to tex, e.g. 'pandoc -f markdown -t latex'")
    //parser.add_argument("--shell", action="store_true", help="run mdtotex command as shell script.")
    fl_o := flag.String("o", "", "out file for latex generation. run with with --tex. if no ct file given, latex template is produced.")
    fl_l := flag.String("l", "", "for a line number from a generated file, which line is it in the ct file? e.g. -l genfile.js:9")
    fl_header := flag.String("header", "", "latex template header file")
    fl_lower := flag.Bool("lower", false, "lowercase tex template")

    flag.Parse()
    args := flag.Args()
    var ctfile string
    switch len(args) {
    case 0:
         fmt.Println("please specify a .ct file: ct myfile.ct");
         os.Exit(0);
    case 1:
    	ctfile = args[0]
    }
    if ctfile == "" {
        if *fl_tex {
            if *fl_o == "" && *fl_header == "" {
                fmt.Println("please specify -o for latex template file and/or --header for header file")
                os.Exit(0)
            }
	    var headerpath string
            if *fl_header == "" {
                tmplpath := fc.Dir(*fl_o)
                headerpath = path.Join(tmplpath, "cthead.tex")
            } else {
                headerpath = *fl_header
	    }
            tmpl := ct.TexTemplate(headerpath)
            header := ct.TexHeader(*fl_lower)

            if *fl_o != "" {
                // notify if the files is already there
                checkoverwrite(*fl_o, tmpl)
	    }
            checkoverwrite(headerpath, header)
	}
        return
    }
    b, _ := os.ReadFile(ctfile)
    text := string(b)
    if *fl_from_org {
        text = ct.Orgtoct(text)
        // print(text)
    }
    if *fl_l != "" { 
        a := strings.Split(*fl_l, ":")
	genfile := a[0]
	genline, _ := strconv.Atoi(a[1])
        ct.Ct(text, filepath.Base(ctfile))
	ctline, err := ct.Ctline(genfile, genline-1)
	fc.Handle(err)
        fmt.Println(ctline+1)
    } else if len(args) == 1 {
        if *fl_tex == true {
	    var out string
	    out = ct.Totex(text, ctfile, *fl_mdtotex)
            a := strings.Split(ctfile, ".")
	    var outname string
            if *fl_o == "" {
                outname = a[0] + ".tex"
            } else if fc.IsDir(*fl_o) { // if just dir given, use the name from the ct file
                outname = path.Join(*fl_o, a[0] + ".tex")
            } else { // path to file given
                outname = *fl_o
	    }
	    f, _ := os.Create(outname)
	    defer f.Close()
	    _, _ = f.WriteString(out)
	    fmt.Println(outname)
        } else {
            err := ct.Ctwrite(text, fc.Dir(ctfile), filepath.Base(ctfile))
	    if err != nil {
	      log.Fatal(err)
	    }
	}
    } 
}
func checkoverwrite(path string, text string) {
    if path == "" { 
        return
    }
    f, err := os.Open(path)
    defer f.Close()
    if err == nil {
        fmt.Printf("the file %s already exists. overwrite it? [Y/n]: ", path)
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(resp)
        if resp != "Y" {
	    //fmt.Println("return")
            return
	}
    } 
    f, err = os.Create(path)
    // defer f.Close()?
    fc.Handle(err)
    f.WriteString(text)
    fmt.Println(path)
}
