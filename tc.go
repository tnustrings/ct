// tc.go does reverse-codetext, builing a ct file from source files.

package ct

import (
  "bufio"
  "fmt"
  "os"
  "path/filepath"
  "regexp"
  "strconv"  
  "strings"
  "github.com/tnustrings/ct/internal/fc"  
)

// variables

// stack holds the nct (number in ct file) of the current chunk
// for pop see https://yourbasic.org/golang/implement-stack/
var stack *Stack[int]

// chunks holds the chunk at each nct (number in ct file)
// in the end, iterating the chunks in order gives the contents of the ct file
var chunks map[int]*Chunk

// lastopened holds the nct (number in ct file) of the last opened chunk
var lastopened int

// Stack is a stack based on a slice
// TODO put in util package?
type Stack[T any] struct {
    // a holds the underlying slice
    a []T
}

// top gives the element on top of the stack
func (st *Stack[T]) top() T {
    // if len(st.a) == 0 { return nil }  // TODO what to do?
    return st.a[len(st.a)-1]
}

// push puts an element on top of the stack
func (st *Stack[T]) push(elem T) {
    st.a = append(st.a, elem)
}

// pop removes the top element from stack and returns it
func (st *Stack[T]) pop() T {
    // if len(st.a) == 0 { return nil } // TODO what to do?
    elem := st.a[len(st.a)-1]
    // remove the top element
    st.a = st.a[:len(st.a)]
    // return
    return elem
}

// n gives the number of elements on the stack
func (st *Stack[T]) n() int {
    return len(st.a)
}

// Tcwrite runs reverse-codetext and writes the resulting ct file.
func Tcwrite(genfiles []string, ctfile string) error {
    tcout, err := Tc(genfiles)
    if err != nil { return err }
    f, err := os.Create(ctfile)
    if err != nil { return err }    
    defer f.Close()
    _, err = f.WriteString(tcout)
    if err != nil { return err }
    return nil
}

// Tc runs reverse-codetext, building ct from source file(s).
func Tc(genfiles []string) (string, error) {

    // load the config
    conf, err := Loadconf()
    if err != nil { return "", err }

    // reset variables
    stack = &Stack[int]{}
    chunks = make(map[int]*Chunk)
    lastopened = -1

    // concat the lines from the files.  TODO maybe later something like line.File
    lines := []Line{}
    for _, genfile := range genfiles {
        // read the file and split into lines
        b, _ := os.ReadFile(genfile)
        text := string(b)
        a := strings.Split(text, "\n") // TODO generic line split method?
        // get the progamming language
        prog := Getpl(conf, filepath.Ext(genfile))
        
        // make Line instances.  skip the ct header.
        skip := true
        for j, l := range a {
            // skip lines before and including the first `` line.
            if skip && isctheaderend(l, prog) {
                skip = false // don't skip anymore
                continue // but skip this line still
            }
            if skip { continue }
            
            // make a Line and append it
            line := Line{Txt: l, Igen: j, Genfile: genfile}
            lines = append(lines, line)
        }
    }

    // we collect comments, cause at the beginning of a comment we don't know whether it's a normal code comment or whether it's last line contains a chunk-opening tag, which would make the comment txta of the opened chunk.
    cmt := []Line{}

    // go over the lines and fill the chunks map, from which the ct is built.
    for _, l := range lines {
        line := l.Txt
        // get the programming language from the file extension.
        prog := Getpl(conf, filepath.Ext(l.Genfile))

        if ischunkopening(line, prog) { // this line opens a chunk.

            // append the line to collected comments
            cmt = append(cmt, l)

            // parse the line.
            _, nct, dot, tag := parseopening(line, prog)

            // if there is no nct given, set it to one higher than the last opened.
            if nct == -1 {
                nct = lastopened + 1
            }

            // update last opened.
            lastopened = nct

            // create a new chunk at nct.
            chunk := &Chunk{}
            chunk.Tag = Line{Txt: tag}
            chunks[nct] = chunk

            // link the new chunk to the first chunk of its node, to keep track of which chunks belong together in a node.
            // if it isn't a declaration (node-opening) chunk, link to the same first chunk as the previous chunk did.
            if !isdeclaration(chunk.Tag.Txt) {
                chunk.Fin = chunks[stack.top()].Fin
            } else {
                // if it's a declaration chunk, it is starting a new node, link to itself.
                chunk.Fin = nct
            }

            // if it's a declaration line, put a reference tag into the code of the mother (the chunk currently on top of the stack).
            if isdeclaration(chunk.Tag.Txt) {
                txt := getleadingspace(line) + "``" + getname(chunk.Tag.Txt) + "``"
                chunks[stack.top()].Code = append(chunks[stack.top()].Code, Line{Txt: txt})
            }
            // if there is a dot before the tag, put in ``.`` for a ghost node
            if dot == "." {
                txt := getleadingspace(line) + "``.``"
                chunks[stack.top()].Code = append(chunks[stack.top()].Code, Line{Txt: txt})
            }

            // set the collected comments as this chunk's txta
            chunk.Txta = maketxt(cmt, prog)

            // clear the comment collect
            cmt = []Line{}

            // now put the new chunk on top of the stack
            stack.push(nct)
        }
        if isnodeclosing(line, prog) { // this line closes a node.

            // only nodes are closed explicitly, chunks are closed implicitly by the comment opening the next chunk.
            
            // get the nct of the first chunk in the node this line closes.
            nctclose := parsenodeclose(line, prog)
            
            // if no nct wasn't given in the node-close, -1 is returned. in this case, we close the current node.
            if nctclose == -1 {
                nctclose = chunks[stack.top()].Fin
            }

            // nctclose now refers to the node on top of the stack or to a node buried inside the stack.

            // we pop nodes from the stack now, until and including the nctclose node.

            // if the nctclose node isn't on top of the stack, we pop the nodes that come before it (all chunks whose fin isn't nctclose).
            for chunks[stack.top()].Fin != nctclose {
                // pop
                stack.pop()
            }

            // now we are at the nctclose node. pop all its chunks.
            for chunks[stack.top()].Fin == nctclose {
                // pop
                stack.pop()
            }

            // now when we append new code we append it to the chunk we were in before we opened the node that was just closed.

        } else if istxtsep(line) { // this line is a text sep.
            // add the collected comments as txtb of the current chunk
            chunks[stack.top()].Txtb = maketxt(cmt, prog)

            // add a text sep marker ``= to txtb for ct building later
            chunks[stack.top()].Txtb = append(chunks[stack.top()].Txtb, Line{Txt: "``="})

            // clear the comment collection
            cmt = []Line{}
        } else if iscmt(line, prog) { // this line is a normal comment line without ct notation.
            // collect the comment
            cmt = append(cmt, l)
        } else { // this line is a code line.
            // if there's something in collected comments, they would be normal code comments without ct functionality. append them.
            if len(cmt) > 0 {
                chunks[stack.top()].Code = append(chunks[stack.top()].Code, cmt...)
                // clear comment collection
                cmt = []Line{}
            }

            // append the code to the current chunk.
            chunks[stack.top()].Code = append(chunks[stack.top()].Code, l)
        }
    }

    // either in this function or in a Tcwrite funktion:
    // build the ct by going over the chunks in order and concatenating the text, tag and code inside them.
    out := []string{}
    for nct := 1; nct <= len(chunks); nct++ {
        chunk := chunks[nct]
        // append the preceeding text
        for _, line := range chunk.Txta { out = append(out, line.Txt) }
        // append the chunk tag
        out = append(out, chunk.Tag.Txt)
        // append the code
        for _, line := range chunk.Code { out = append(out, line.Txt) }

        // append the text after the code (if any)
        for _, line := range chunk.Txtb { out = append(out, line.Txt) }
    }

    // return the complete ct string
    return strings.Join(out, "\n"), nil
}

// chunkopenre is the string from which the chunk open regexp is built. it has a placeholder %s to dynamically insert the comment mark for the programming language.  it holds four capture groups: the comment text, the chunk nct, a dot if this line opens a ghost node, and the chunk tag.  if one of the optional capture groups is not used, the others retain their positions in string array returned by FindStringSubmatch.
var chunkopenre = "^\\s*%s (.*)\\s+(\\d+)?(\\.)?(``.*)\\s*"

// ischunkopening says whether this comment line is chunk opening.  a minimal chunk opening line is
// # ``
// (with space after the comment mark), a longer chunk-opening line is:
// # my last line of txta  3``/my/path:
// denoting that this is the third chunk and its tag is ``/my path:
func ischunkopening(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(chunkopenre, prog.Cmtmark))
    return re.MatchString(line)
}

// parseopening parses a comment line opening a chunk. it returns (txt, nct, dot, tag), where txt is the text of the comment line, nct is the nct of the opened chunk, dot is "." if the chunk resulted from opening a ghost node and tag is the tag (path) of the chunk.
func parseopening(line string, prog *Prog) (string, int, string, string) {
    re := regexp.MustCompile(fmt.Sprintf(chunkopenre, prog.Cmtmark))
    // get the matches from the capture groups.  matches retain their position if optional capture groups are not used.
    matches := re.FindStringSubmatch(line)
    // parse the nct to int
    nct := -1
    if matches[2] != "" {
       nct, _ = strconv.Atoi(matches[2])
    }
    return matches[1], nct, matches[3], matches[4]
}
    
// chunkclosere is the string from which the chunk close regexp is built. like chunkopenre, it has a placeholder %s for the comment mark used by the programming language.
var chunkclosere = "^\\s*%s``\\s(\\d+)?*$"

// isnodeclosing says whether this comment line closes a note (the chunks that make up a node are closed implicitly by the next chunk).  a node-close line consists of the last line of code in this chunk followed by a comment marker immediately followed by `` (without whitespace) followed by an optional number referencing the opening chunk in the node this line closes.
func isnodeclosing(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(chunkclosere, prog.Cmtmark))
    return re.MatchString(line)
}

// parsenodeclose parses a node-close line and returns the nct in this line denoting the first chunk in the closed node.
func parsenodeclose(line string, prog *Prog) int {
    re := regexp.MustCompile(fmt.Sprintf(chunkclosere, prog.Cmtmark))
    matches := re.FindStringSubmatch(line)
    // parse the nct to int
    nct := -1
    if matches[1] != "" {
       nct, _ = strconv.Atoi(matches[1])
    }
    return nct    
}

// cmtre is the string from which the comment regexp is built. it has a placeholder %s for the comment mark used by the programming language and a capture group to capture the comment text.
var cmtre = "^\\s*%s\\s*(.*)$"

// iscmt says whether this line is a comment in the given programming lang.
func iscmt(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(cmtre, prog.Cmtmark))
    return re.MatchString(line)
}

// ctheaderendre fishes for the last line of the ct header. it contains a placeholder %s for the comment mark.
var ctheaderendre = "^%s`` (.*)$"

// isctheaderend says whether this line is the last line of the ct header, that's a line starting with `` and containing the name of the ct file and all generated files.
func isctheaderend(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(ctheaderendre, prog.Cmtmark))
    return re.MatchString(line)
}

// ctheaderfiles gets the files list from a ctheaderend line.
func ctheaderfiles(line string, prog *Prog) []string {
    re := regexp.MustCompile(fmt.Sprintf(ctheaderendre, prog.Cmtmark))
    match := re.FindStringSubmatch(line)
    if len(match) < 1 { return []string{} }
    return strings.Split(match[1], " ") // TODO quote-sensitive split?
}

// cmttxt returns whatever text is behind the first comment mark in the line.
func cmttxt(line string, prog *Prog) string {
    re := regexp.MustCompile(fmt.Sprintf(cmtre, prog.Cmtmark))
    matches := re.FindStringSubmatch(line)
    return matches[1]
}

// getname returns the name (last part of the path) from a chunk tag.
func getname(chunktag string) string {
    // strip the double ticks
    path := stripdblticks(chunktag)
    // strip the declaration colon
    path = stripdeclcolon(path)
    // get the path elements
    elems := strings.Split(path, "/")
    // return the last element
    if len(elems) == 0 { return "" }
    return elems[len(elems)-1]
}

// maketxt turns lines of comments into lines of between-chunk text.
func maketxt(cmt []Line, prog *Prog) []Line {
    out := []Line{}
    for _, line := range cmt {
        txt := ""
        // if the line was a chunk opening line extract the text.
        if ischunkopening(line.Txt, prog) {
            txt, _, _, _ = parseopening(line.Txt, prog)
        } else {
            txt = cmttxt(line.Txt, prog)
        }
        out = append(out, Line{Txt: txt})
    }
    return out
}


// Filesfromgen returns the paths of associated files (ct file and
// generated files) from the header of a generated file. a header line would be:
// #`` myprog.ct gen1.py gen2.py 
func Filesfromgen(genfile string, conf *Conf) (string, []string, error) {
    // get the file extension, dir, and load the programming language info.
    ext := filepath.Ext(genfile)
    dir := fc.Dir(genfile)    
    prog := Getpl(conf, ext)
    
    // open the file and make a scanner.
    file, err := os.Open(genfile)
    if err != nil { return "", []string{}, err }
    defer file.Close()
    scanner := bufio.NewScanner(file)

    // read the lines, see whether a line matches the end of a ct header.
    i := 0
    for scanner.Scan() {
        line := scanner.Text()
        if isctheaderend(line, prog) {
            // extract the file names from the end-of-header line.
            filenames := ctheaderfiles(line, prog)
            if len(filenames) == 0 {
                return nil, error(genfile + ": list of filenames expected in line " + strconv.Itoa(i+1))
            }
            // return the files as complete paths.
            paths := []string{}
            for _, path := range filenames {
                if dir != "" { path = dir + "/" + path } // TODO okay so?
                paths = append(paths, path)
            }
            return paths[0], paths[1:], nil
        }
        i++
    }
    // if no ct line found, error.

    log.Fatal("no ct-line in " + genfile)
}

// Filesfromct returns the paths of the files generated from a ct file.
func Filesfromct(ctfile) ([]string, err) {
    // run codetext.
    err := Ct(text, ctfile)
    if err != nil { return nil, err }
    // get the directory.
    dir := fc.Dir(path)
    
    // collect the generated filenames, return the complete paths.
    out := []string{}
    for path, _ := range roots { // TODO run this on a ct instance? or let Ct() return the roots?
        if dir != "" { path = dir + "/" + path } // TODO okay so?

        out = append(out, path)
    }
    return out, nil
}