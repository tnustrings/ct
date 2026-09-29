// tc.go does reverse-codetext, builing a ct file from source files.

package ct

// variables

// stack holds the nct (number in ct file) of the current chunk
// for pop see https://yourbasic.org/golang/implement-stack/
var stack []int

// chunks holds the chunk at each nct (number in ct file)
// in the end, iterating the chunks in order gives the contents of the ct file
var chunks map[int]*Chunk

// lastopened holds the nct (number in ct file) of the last opened chunk
var lastopened int

// Tcwrite runs reverse-codetext and writes the resulting ct file.
func Tcwrite(genfiles []string, ctfile string) error {
    tcout, err := Tc(genfiles), "\n"
    if err != nil { return err }
    f, err := os.Create(ctfile)
    if err != nil { return err }    
    defer f.Close()
    _, err := f.WriteString(tcout)
    if err != nil { return err }
    return nil
}

// Tc runs reverse-codetext, building ct from source file(s).
func Tc(genfiles []string) string, error {

    // reset variables
    stack = []int{}
    chunks = make(map[int]*Chunk)
    lastopened = -1

    // concat the lines from the files.  TODO maybe later something like line.File

    // we collect comments, cause at the beginning of a comment we don't know whether it's a normal code comment or whether it's last line contains a chunk-opening tag, which would make the comment txta of the opened chunk.
    cmt := []string{}

    // go over the lines and fill the chunks map, from which the ct is built.
    for _, line := range lines {

        if ischunkopening(line, prog) { // this line opens a chunk.

            // append the line to collected comments
            cmt = append(cmt, line)

            // parse the line.
            _, nct, dot, tag := parseopening(line)

            // if there is no nct given, set it to one higher than the last opened.
            if nct == -1 {
                nct = lastopened + 1
            }

            // update last opened.
            lastopened = nct

            // create a new chunk at nct.
            chunks[nct] = &Chunk{}

            // link the new chunk to the first chunk of its node, to keep track of which chunks belong together in a node.
            // if it isn't a declaration (node-opening) chunk, link to the same first chunk as the previous chunk did.
            if !isdeclaration(chunk.tag) {
                chunk.Fin = chunks[stack[n-1]].first
            } else {
                // if it's a declaration chunk, it is starting a new node, link to itself.
                chunk.Fin = nct
            }

            // if it's a declaration line, put a reference tag into the code of the mother (the chunk currently on top of the stack).
            if isdeclaration(chunk.tag) {
                txt := getleadingspace(line) + "``" + getname(chunk.tag) + "``"
                append(chunks[stack[n-1]].code, txt)
            }

            // set the collected comments as this chunk's txta
            chunk.txta = maketxt(cmt, prog)

            // clear the comment collect
            cmt = []string{}

            // now put the new chunk on top of the stack
            append(stack, nct)
            n++
        }
        if isnodeclosing(line, prog) { // this line closes a node.

            // only nodes are closed explicitly, chunks are closed implicitly by the comment opening the next chunk.
            
            // get the nct of the first chunk in the node this line closes.
            nctclose := parseclosing(line)
            
            // if no nct wasn't given in the node-close, -1 is returned. in this case, we close the current node.
            if nctclose == -1 {
                nctclose = chunks[stack[n-1]].Fin
            }

            // nctclose now refers to the node on top of the stack or to a node buried inside the stack.

            // we pop nodes from the stack now, until and including the nctclose node.

            // if the nctclose node isn't on top of the stack, we pop the nodes that come before it (all chunks whose fin isn't nctclose).
            while chunks[stack[n-1]].Fin != nctclose {
                // pop
                stack = stack[:n]
                n--
            }

            // now we are at the nctclose node. pop all its chunks.
            while chunks[stack[n-1]].first == nctclose {
                // pop
                stack = stack[:n]
                n--
            }

            // now when we append new code we append it to the chunk we were in before we opened the node that was just closed.

        }
        else if istxtsep(line, prog) { // this line is a text sep.
            // add the collected comments as txtb of the current chunk
            chunks[stack[n-1]].txtb = maketxt(cmt, prog)

            // add a text sep marker ``= to txtb for ct building later
            append(chunks[stack[n-1]].txtb, "``=")

            // clear the comment collection
            cmt = []
        } else if iscmt(line, prog) { // this line is a normal comment line without ct notation.
            // collect the comment
            cmt = append(cmt, line)
        } else { // this line is a code line.
            // if there's something in collected comments, they would be normal code comments without ct functionality. append them.
            if len(cmt) > 0 {
                append(chunks[stack[n-1]].code, cmt...)
                // clear comment collection
                cmt = []string{}
            }

            // append the code to the current chunk.
            append(chunks[stack[n-1]].code, line)
        }
    }

    // either in this function or in a Tcwrite funktion:
    // build the ct by going over the chunks in order and concatenating the text, tag and code inside them.
    out := str[]{}
    for nct = 1; nct <= chunks.size; nct++ {
        chunk := chunks[nct]
        // append the preceeding text
        for _, line := range chunk.txta { out = append(out, line) }
        // append the chunk tag
        out = append(out, chunk.tag)
        // append the code
        for _, line := range chunk.code { out = append(out, line) }

        // append the text after the code (if any)
        for _, line := range chunk.txtb { out = append(out, line) }
    }

    // return the complete ct string
    return strings.Join(out, "\n"), nil
}

// chunkopenre is the string from which the chunk open regexp is built. it has a placeholder %s to dynamically insert the comment mark for the programming language.  it holds four capture groups: the comment text, the chunk nct, a dot if this line opens a ghost node, and the chunk tag.  if one of the optional capture groups is not used, the others retain their positions in string array returned by FindStringSubmatch.
var chunkopenre := "^\\s*%s (.*) (\\d+)?(\\.)?(``.*)\\s*"

// ischunkopening says whether this comment line is chunk opening.  a minimal chunk opening line is
// # ``
(with space after the comment mark), a longer chunk-opening line is:
// # my last line of txta 3``/my/path:
// denoting that this is the third chunk and its tag is ``/my path:
func ischunkopening(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(chunkopenre, prog.Cmtmark))
    return re.MatchString(line)
}

// parseopening parses a comment line opening a chunk. it returns (txt, nct, dot, tag), where txt is the text of the comment line, nct is the nct of the opened chunk, dot is "." if the chunk resulted from opening a ghost node and tag is the tag (path) of the chunk.
func parseopening(line string, prog *Prog) (string, int, string, string) {
    re := regexp.MustCompile(fmt.Sprintf(chunkopenre, prog.Cmtmark))
    // get the matches from the capture groups.  matches retain their position if optional capture groups are not used.
    matches = re.FindStringSubmatch(line)
    // parse the nct to int
    nct := -1
    if matches[2] != "" {
       nct = strconv.Atoi(matches[2])
    }
    return (matches[1], nct, matches[3], matches[4])
}
    
// chunkclosere is the string from which the chunk close regexp is built. like chunkopenre, it has a placeholder %s for the comment mark used by the programming language.
var chunkclosere := "^\\s*%s``\\s(\\d+)?*$"

// isnodeclosing says whether this comment line closes a note (the chunks that make up a node are closed implicitly by the next chunk).  a node-close line consists of the last line of code in this chunk followed by a comment marker immediately followed by `` (without whitespace) followed by an optional number referencing the opening chunk in the node this line closes.
func isnodeclosing(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(chunkclosere, prog.Cmtmark))
    return re.MatchString(line)
}

// parseclose parses a chunk-close line and returns the nct in this line denoting the first chunk in the closed node.
func parseclose(line, prog *Prog) int {
    re := regexp.MustCompile(fmt.Sprintf(chunkclosere, prog.Cmtmark))
    matches = re.FindStringSubmatch(line)
    // parse the nct to int
    nct := -1
    if matches[1] != "" {
       nct = strconv.Atoi(matches[1])
    }
    return nct    
}

// cmtre is the string from which the comment regexp is built. it has a placeholder %s for the comment mark used by the programming language and a capture group to capture the comment text.
var cmtre := "^\\s*%s\\s*(.*)$"

// iscmt says whether this line is a comment in the given programming lang.
func iscmt(line string, prog *Prog) bool {
    re := regexp.MustCompile(fmt.Sprintf(cmtre, prog.Cmtmark))
    return re.MatchString(line)
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
func maketxt(cmt []string, prog *Prog) []string {
    out := []string{}
    for _, line := range cmt {
        var txt := ""
        // if the line was a chunk opening line extract the text.
        if ischunkopening (line) {
            txt, _, _, _ = parseopening(line, prog)
        } else {
            txt = cmttxt(line)
        }
        out = append(out, txt)
    }
    return out
}