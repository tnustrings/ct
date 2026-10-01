// main.go was generated from main.ct. 
// you can edit main.ct or main.go and then run `ct main.ct` to update the other file.
// when editing this file leave the `` tags.
//`` main.ct main.go

// # main
// 
// the main cli.  1``//main.go:
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
// # main function
// 
// main kicks off tangling or tex generating.  2``/main:
func main() {
  // # flags
  // 
  // get the flags.  3``./flags:
    fl_tex := flag.Bool("tex", false, "print doc as latex")
    fl_from_org := flag.Bool("from_org", false, "input is a .org file")
    fl_mdtotex := flag.String("mdtotex", "", "for latex doc generation. a command to convert markdown between codechunks to tex, e.g. 'pandoc -f markdown -t latex'")
    //parser.add_argument("--shell", action="store_true", help="run mdtotex command as shell script.")
    fl_o := flag.String("o", "", "out file for latex generation. run with with --tex. if no ct file given, latex template is produced.")
    fl_l := flag.String("l", "", "for a line number from a generated file, which line is it in the ct file? e.g. -l genfile.js:9")
    fl_header := flag.String("header", "", "latex template header file")
    fl_lower := flag.Bool("lower", false, "lowercase tex template")

    flag.Parse()
  // get the positional args.  4``
    args := flag.Args()
  // take the file as positional arg.
  // 
  // be able to handle more than one ct file?  5``
    var file string
    switch len(args) {
    case 0:
         fmt.Println("please specify a .ct file: ct myfile.ct");
         os.Exit(0);
    case 1:
    	file = args[0]
    }
  //``3
  // if no ct file was given and the -tex flag is set generate the latex
  // template and header.  6``/main/tex tmpl:
    if file == "" {
        if *fl_tex {
	    // if no names for the ouput files were given, give an error asking for a
	    // name.  7.``
            if *fl_o == "" && *fl_header == "" {
                fmt.Println("please specify -o for latex template file and/or --header for header file")
                os.Exit(0)
            }
	    // if only the header name isn't given, put the header into the same
	    // directory as the template file (and name it cthead.tex).  8``
	    var headerpath string
            if *fl_header == "" {
                tmplpath := fc.Dir(*fl_o)
                headerpath = path.Join(tmplpath, "cthead.tex")
            } else {
                headerpath = *fl_header
	    }
	    // generate template and header, and write them.  9``
            tmpl := ct.TexTemplate(headerpath)
            header := ct.TexHeader(*fl_lower)

            if *fl_o != "" {
                // notify if the files is already there
                checkoverwrite(*fl_o, tmpl)
	    }
            checkoverwrite(headerpath, header)
	    //``7
	}
        return
    }
  //``6
  // ## read and convert
  // 
  // if a codetext file was given, read it.  10``/main/read:
    b, _ := os.ReadFile(file)
    text := string(b)
  //``10
  // should the text be converted from org to ct?  11``/main/convert org:
    if *fl_from_org {
        text = ct.Orgtoct(text)
        // print(text)
    }
  //``11
  // ## compile
  // 
  // now comes the part of main involving assembly of the given ct file.  12``/main/compile:
    if *fl_l != "" { 
        // ### go to line
        // 
        // map from line number in generated source to original line number in ct.
        // 
        // get the name of the generated file and line number from flag arg.  18``/main/compile/go to:
        a := strings.Split(*fl_l, ":")
	genfile := a[0]
	igen, _ := strconv.Atoi(a[1])
        // first run ct without writing files.  19``
        ct.Ct(text, filepath.Base(file))
        // then print the original line number.
        // 
        // subtract and add 1 to go from one-indexed input number to zero-indexed
        // and vice versa.  20``
	ict, err := ct.Ict(genfile, igen-1)
	fc.Handle(err)
        fmt.Println(ict+1)
        //``18
    } else if len(args) == 1 {
        if *fl_tex == true {
	    // ### generate tex
	    // 
	    // generate a tex file from ct.  13``/main/compile/tex:
	    var out string
	    out = ct.Totex(text, file, *fl_mdtotex)
	    // if no out name for tex given, take it from the ct file.  14``
            a := strings.Split(file, ".")
	    var outname string
            if *fl_o == "" {
                outname = a[0] + ".tex"
            } else if fc.IsDir(*fl_o) { // if just dir given, use the name from the ct file
                outname = path.Join(*fl_o, a[0] + ".tex")
            } else { // path to file given
                outname = *fl_o
	    }
	    // write and say which file was written.  15``
	    f, _ := os.Create(outname)
	    defer f.Close()
	    _, _ = f.WriteString(out)
	    fmt.Println(outname)
	    //``13
        } else {
	    // ### ct or tc
	    // 
	    // the user can pass a ct file or a generated file.  based on which of
	    // the files is newer, run ct or tc (reverse ct).
	    // 
	    // find out all generated files.
	    // 
	    // if the user passed a generated file, read the ct file and eventual other
	    // generated files from its header.
	    // 
	    // if the user passed a ct file, compile it and get the generated files
	    // from there.  16``/main/compile/ct or tc:
            // load the config
            conf, err := ct.Loadconf()
            if err != nil { log.Fatal(err) }

            var ctfile string
            var genfiles []string
            if filepath.Ext(file) != ".ct" {
                ctfile, genfiles, err = ct.Filesfromgen(file, conf)
                if err != nil { log.Fatal(err) }
            } else {
                ctfile = file
                genfiles, err = ct.Filesfromct(ctfile)
                if err != nil { log.Fatal(err) }
            }
	    // run ct or tc based on the dates of the files.  17``
            if ok, err := shouldtc(ctfile, genfiles); ok {
                err := ct.Tcwrite(genfiles, ctfile)
                if err != nil { log.Fatal(err) }
            } else {
                // TODO uncomment.  for now don't overwrite ct file, until it's safe.bm
                // err := ct.Ctwrite(ctfile) 
                if err != nil { log.Fatal(err) }
            }
	    //``16
	}
    } 
  //``12
}
//``2
// # checkoverwrite
// 
// checkoverwrite writes text to a file and asks before whether an
// existing file should be overwritten.  21``/checkoverwrite:
func checkoverwrite(path string, text string) {
    // return on empty path. ok so?
    // 
    // open it.  22.``
    if path == "" { 
        return
    }
    f, err := os.Open(path)
    defer f.Close()
    // we'd like to check if the file already exists. for now, if there is no
    // error, the file exists. (is it ok to check like this?)  23``
    if err == nil {
        // ask whether to overwrite.  24``./ask:
        fmt.Printf("the file %s already exists. overwrite it? [Y/n]: ", path)
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	resp = strings.TrimSpace(resp)
        if resp != "Y" {
	    //fmt.Println("return")
            return
	}
        //``24
    } 
    // create the file in any case, if it already exists, to truncate
    // (overwrite) it. see https://stackoverflow.com/a/72181498  25``../
    f, err = os.Create(path)
    // defer f.Close()?
    fc.Handle(err)
    // write and print which file was written.  26``
    f.WriteString(text)
    fmt.Println(path)
    //``22
}
//``21
// shouldtc says whether ct or tc should be run: if one of the generated
// files is newer than the ct file, run tc, else run ct.  27``/shouldtc:
func shouldtc(ctfile string, genfiles []string) (bool, error) {
    // get the file info for the ct file.  28``_:
    ctinfo, err := os.Lstat(ctfile)
    if err != nil { return false, err }
    // go over the gen files and see whether one is newer than the ct file.  29``
    for _, genfile := range genfiles {
        geninfo, err := os.Lstat(genfile)
        if err != nil { return false, err }
        if geninfo.ModTime().After(ctinfo.ModTime()) {
            // one generated file was modified after ct file was modified, run tc
            return true, nil
        }
    }
    // if all generated files were modified before the ct file was modified, run ct.  30``
    return false, nil
    //``28
}
//``27
//``1
