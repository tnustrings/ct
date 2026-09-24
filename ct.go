package ct

import (
  "cmp"
  "embed"
  "errors"
  "fmt"
  "encoding/json"
  "os"
  "path/filepath"
  "regexp"
  "slices"
  "strconv"
  "strings"  
)

// embed the conf folder
//go:embed conf/*
var embedded embed.FS

// variables

// save the ctlines for printtex
// lines keep their \n
var ctlines []string

// for each generated file, map its line numbers to the original line numbers in ct file
var ictmap map[string]map[int]int

// currentnode is the node we're currently at
var currentnode *node

// openghost is the open ghost node. if the last chunk opened a ghostnode, it's this one
var openghost *node

// roots holds the root nodes
var roots map[string]*node

// roottext holds the generated code for each root
var roottext map[string]string

// nodeatict (node-at-index-ct) holds the node at a specific ct line (if there is one) 
var nodeatict map[int]*node // was nat

// is this ctline opening a chunk? was ischunkopening
var chop map[int]bool 

// is this ctline closing a chunk? was ischunkclose
var chclo map[int]bool 

// Line holds a line of text and its index in the ct file
type Line struct {  // all lower case? does this clash with variables named 'line'?
    Txt string // the text of the line
    Ict int // the index of the line in the ct file
}

// Chunk holds lines of codes, and the text before and after.
type Chunk struct {
    // Txta holds the lines of text before the code
    Txta []Line
    // Tag holds the tag line opening the chunk
    Tag Line
    // Code holds the lines of code in the chunk
    Code []Line
    // Txtb holds optional lines of text after the chunk, if they were seperated from the text belonging to the next chunk by ``=
    Txtb []Line
    // Nct says that this is the nth chunk in the ct file
    Nct int
}

// a node holds multiple code chunks sharing the same path, and references the child nodes the code in the chunks spawns.
type node struct {

    /* name holds the node's name.

    if it's a ghost node, the name starts with dot '.' and is followed by
    the node's index in its parent's ghostchilds.
    
    although ghost names can never be used to reference or go to a
    ghost node it's handy if the names of a node's ghostchilds are
    distinct for latex links */

    name string

    // parent is the node's from which this node is spawned/referenced. if it is nil this node is a root.
    parent *node

    // childs: the named children
    childs map[string]*node

    // the ghostparent
    ghostparent *node

    /* ghostchilds: the ghost children. although each code chunk may only spawn one ghostchild, there can be multiple chunks appended to one node, so we can end up with multiple ghostchilds. */
    ghostchilds []*node

    // chunks holds the code chunks at this node
    chunks []*Chunk
    
    // d: has this node been declared with a colon ':'. every node except ghost nodes needs to have been declared.
    d bool

    // r: has this node been referenced. every node except root nodes needs to have been referenced.
    r bool

    // iip (index in parent): at which line of the parent is this node?
    iip int // was lineinparent

    // caict (child at ict): the child (named or ghost) at this ct line (if any)
    caict map[int]*node // was childatctline

    // chup (chunk in parent): this node is referenced from the ith chunk in the parent node
    chup int // was iparentchunk
}

// ls lists the named childs
func (n *node) ls() []string {
    return keys(n.childs)
}

// newnode makes a new node
func newnode(name string, parent *node) *node {
    n := node{name: name, parent: parent}
    n.childs = make(map[string]*node)
    n.caict = make(map[int]*node)
    return &n    
}

// ct holds the ct conf
/*type ct struct {
    conf Conf
}*/

// Conf holds the config
type Conf struct {
    Proglang []Prog
}

// Prog holds the proglang of conf
type Prog struct {
    Name string // language name
    Ext []string // file extensions
    Fncre string // regexp for function declaration
    Cmtmark string // mark for line comment
    Cmtopen string // opening mark for multiline comment
    Cmtduring string // mark during multiline comment
    Cmtclose string // closing mark for multiline comment
    Cmtindent string // comment indent
    Fnccmt string // function comment
}

// loadconf loads the conf file. if it's not there, create it.
// should we make the conf accessible or just package it with the program?
func loadconf() (*Conf, error) {
    // create an empty conf and unmarshal the conf into it
    conf := Conf{}
    b, err := embedded.ReadFile("conf/conf.json")
    if err != nil { return nil, err }
    err = json.Unmarshal(b, &conf)
    if err != nil { return nil, err  }
    //fmt.Printf("conf: %v\n", conf)

    // return the config
    return &conf, err
}

// keys returns the keys of a map
func keys[K cmp.Ordered, V any] (m map[K]V) []K {
    var names []K
    for name, _ := range m {
        names = append(names, name)
    }
    slices.Sort(names)
    return names
}

/* cdone walks one step from node. if it goes back from a ghost, put
the named childs of a ghost to the ghost's last named parent (lnp). a
ghost node isn't accessible after exiting it, but its named childs
should be. so the named childs end up with two parents, the
ghostparent that determines at which line they are assembled and the
lnp via which they can be accessed in paths.*/

func (n *node) cdone(step string) *node {
    //fmt.Printf("cdone %s\n", step)
    if isghost(step) {
        // we may not walk into a ghost node via path
        fmt.Printf("error: node names starting with . are not allowed.\n")
	os.Exit(-1)
    }
    // stay at node
    if step == "" || step == "." {
        return n
    }
    // go back
    if step == ".." { 
        if isghost(n.name) { exitghost(n) }
	return n.parent
    }
    // go to a named child
    if _, ok := n.childs[step]; ok {
        return n.childs[step]
    }
    // step not found
    return nil
}

// isroot says wether the node is a root, that is whether its parent is nil
func (n *node) isroot() bool {
   if n.parent == nil {
       return true
   }
   return false
}

// exitghost moves a ghost node's named children to its last named parent. needs to be called after leaving a ghost node when building the node tree.
func exitghost(ghost *node) {
    if ghost == nil || !isghost(ghost.name) || len(ghost.childs) == 0 {
        return
    }
    
    /* we exit a ghost node with named children. move all its named
    childs to the ghost node's parent so that they can be accessed
    from there and set the ghostnode as the childs' ghostparent (from
    where a child can get e.g. its indent) */

    // get the last named parent
    lnp := lnp(ghost)
    for name, child := range ghost.childs {
        // remember the ghostparent
        child.ghostparent = ghost
	
	// move the child from ghost to lnp

	/* when putting the child in lnp's named childs, we don't need
	to worry about the name already being taken, because we moved
	every child that could be touched here that was already there
	inside the ghostnode upon creating it. */ // still true?

	// reset the parent
	child.parent = lnp
	// add the child to lnp's named children
	lnp.childs[child.name] = child
	// delete it from ghost's named children
	delete(ghost.childs, name)
    }	
}

// isname returns true if line is the referencing name line of a code chunk
func isname(line string) bool {
    // the name needs to contain at least one non-tick to distinguish it from three-tick ``` markdown code-block openings
    re := regexp.MustCompile(".*``[^`]+``")
    return re.MatchString(line)
}

// isdblticks says if the line consists of two ticks only (not considering programming-language hashtag)
// this could either be a start line of an unnamed chunk or an end line of a chunk
func isdblticks(line string) bool {
    // return re.match(r"^@$", line) # only allow single @ on line, to avoid mistaking @-code-annotations for doku-markers
    re := regexp.MustCompile("^``(\\s+#\\w+)?\\s*$")
    return re.MatchString(line)
}

// istxtsep says whether a line is the text seperator ``= that divides text after and before a chunk if needed.
func istxtsep(line string) bool {
    re := regexp.MustCompile("^``=\\s*$")
    return re.MatchString(line)
}

// isghost says whether it's a ghost-name (starting with a .) # to js
func isghost(name string) bool {
    // check that it's not a dot followed by a non-dot, maybe it would be enough to check that it's not a dot followed by a number.
    re := regexp.MustCompile(`^\.[^\.]`)
    return re.MatchString(name)
}

// isfromroot says whether the name starts from a root
func isfromroot(name string) bool {
     re := regexp.MustCompile("^//")
     return re.MatchString(name)
}

// getname gets the chunkname from a chunk-opening or in-chunk reference
func getname(line string) string {
    // remove the leading ticks (openings and references)
    r1 := regexp.MustCompile("^[^`]*``")
    // replace only the first occurence
    found := r1.FindString(line)
    name := strings.Replace(line, found, "", 1)
    
    // remove the trailing ticks (references only)
    r2 := regexp.MustCompile("``.*")
    name = r2.ReplaceAllString(name, "")
        
    // remove the newline (openings only)
    r3 := regexp.MustCompile("\n$")
    name = r3.ReplaceAllString(name, "")

    // remove the programming language hashtag (if there) (openings only)
    r4 := regexp.MustCompile(`\s+#\w+$`)
    name = r4.ReplaceAllString(name, "")

    // don't remove the declaration colon, we need it in put()
    
    // debug(f"getname({line}): '{name}'")

    return name
}

// trimlines removes empty lines at beginning and end of slice.
func trimlines(lines []Line) []Line {
    out := []Line{}
    empty := regexp.MustCompile(`^\s*$`)
    i := 0
    // empty lines at the beginning
    for i = 0; i < len(lines); i++ {
        if !empty.MatchString(lines[i].Txt) { break }
    }
    j := 0
    // empty lines at the end
    for j = len(lines)-1; j >= 0; j-- {
        if !empty.MatchString(lines[j].Txt) { break }
    }
    if i <= j { return lines[i:j] }
    return []Line{}
    return out
}

// assemble assembles a codechunk recursively, filling up its leading
// space to leadingspace. this way we can take chunks that are already
// (or partly) indented with respect to their parent in the editor, and
// chunks that are not.
// TODO: don't return a slice of lines but write into a slice of lines? how would that work?
func assemble(n *node, leadingspace string, rootname string, proglang string, ctfile string, conf *Conf) []Line { // don't pass the conf here

    var lastnamedp *node

    // if it's a ghost node, remember the last named parent up the tree
    if isghost(n.name) {
        lastnamedp = lnp(n)
    }
    
    // find out a first line how much this chunk is already indented
    // and determine how much needs to be filled up
    var alreadyspace string

    leadspacere := regexp.MustCompile("^\\s*")
    // leading space already there (in first line)
    if len(n.chunks) > 0 && len(n.chunks[0].Code) > 0 {
        firstcodeline := n.chunks[0].Code[0]
	f := leadspacere.FindStringIndex(firstcodeline.Txt)
        alreadyspace = firstcodeline.Txt[f[0]:f[1]]
    } else {
        alreadyspace = "" // no line, so no leading space already there
    }
    // space that needs to be added
    addspace := ""
    if len(leadingspace) > len(alreadyspace) {
        addspace = leadingspace[0:len(leadingspace)-len(alreadyspace)]
    }

    prog := getpl(conf, proglang)

     // insert comments from previous text nodes.  do this here because the programming language is now safe to be known after all the nodes have been put.  line referencing depends on whether lines were inserted, so do it here also.
    // outlines := insertcmt(n.lines, n.prevlines, proglang, n.isroot(), ctfile, conf)  // TODO uncomment

    // if the rootname isn't in ictmap yet, put it there
    if _, ok := ictmap[rootname]; !ok {
        ictmap[rootname] = make(map[int]int)
    }

    // the current ghost node
    ighost := 0

    out := []Line{}
    
    for _, chunk := range n.chunks {
        // add comments to be able to reconstruct the ct file
        outnew := addopening(chunk, leadingspace, prog)
	out = append(out, outnew...)

        // assemble the code
        for _, line := range chunk.Code {
	
	    if isname(line.Txt) {

		// remember leading whitespace
		childleadingspace := leadspacere.FindString(line.Txt) + addspace 
		name := getname(line.Txt)
		if name == "." {   // assemble a ghost-child
		    outnew := assemble(n.ghostchilds[ighost], childleadingspace, rootname, proglang, ctfile, conf)
		    out = append(out, outnew...)
		    ighost += 1
		} else {             // assemble a named child
		    var child *node
		    if isghost(n.name) {
			// if we're at a ghost node, we get to the child via the last named ancestor
			child = lastnamedp.childs[name]
		    } else {
			child = n.childs[name]
		    }
		    outnew := assemble(child, childleadingspace, rootname, proglang, ctfile, conf)
		    out = append(out, outnew...)
		}
	    } else { // normal code line
	        // append the line to the output
	        txt := addspace + line.Txt
		out = append(out, Line{Txt: txt, Ict: line.Ict })
	    }
	}

        // add comments to be able to reconstruct the ct file
	outnew = addclosing(chunk, leadingspace, prog)
	out = append(out, outnew...)
    }
    
    // add a node close comment to be able to reconstruct the ct file        
    if len(n.chunks) > 0 {
        nctfirst := n.chunks[0].Nct
        lastchunk := n.chunks[len(n.chunks)-1]

        outnew := addnodeclose(nctfist, lastchunk, leadingspace, prog)
        out = append(out, outnew...)
    }
    
    //debug("out:")
    //debug(out)
    return out
}

// addopening adds text lines, chunk tag and chunk number as comment as to be able to reconstruct the ct file
func addopening(chunk *Chunk, leadingspace string, prog *Prog) []Line {
    out := []Line{}
    for i, line := range(chunk.Txta) {
        // make a comment from the between-chunk text
        txt := leadingspace + prog.Cmtmark + " " + line.Txt
	// add the chunk tag to the last comment line
	if i == len(chunk.Txta)-1 {
	    txt += " " + itoa(chunk.Nct) + chunk.Tag.Txt
	}
	out = append(out, Line{Txt: txt, Ict: line.Ict})
    }
    return out
}

// addclosing adds the txtb of a chunk to the generated code as to be able to reconstruct the ct file
func addclosing(chunk *Chunk, leadingspace string, prog *Prog) []Line {
    out := []Line{}
    for _, line := range(chunk.Txtb) {
        // add the text after a chunk
        txt := leadingspace + prog.Cmtmark + " " + line.Txt
	out = append(out, Line{Txt:txt, Ict:line.Ict})
    }
    return out
}

// addnodeclose adds a comment to signify that the node that started with chunk nct is closed with lastchunk.
func addnodeclose(nct int, lastchunk *Chunk, leadingspace string, prog *Prog) [] Line {
    out := []Line{}
    // add a comment signifying chunk closure
    txt := leadingspace + prog.Cmtmark + "``" + itoa(nct)
    out = append(out, Line{Txt:txt, Ict: chunk.Tag.Ict + len(chunk.Code)+1})
    return out
}

// insertcmt inserts potential function comments from prevlines into lines. it also inserts don't-edit comments.
func insertcmt(lines []Line, prevlines map[int][]Line, proglang string, isroot bool, ctfile string, conf *Conf) []Line {

     prog := getpl(conf, proglang)
     
    // make a regexp to recognize (and extract) function names for the node's programming language.
    funcre := regexp.MustCompile(prog.Fncre)
    
    // are comments inserted before or after function declaration?
    cmtbefore := true
    if prog.Fnccmt == "after" { cmtbefore = false }

    // the text lines preceeding the current chunk
    var myprevlines []Line

    // leading space regexp
    leadspacere := regexp.MustCompile("^\\s*")

    // the output lines
    out := []Line{} 
    
    // map from the line number in the node to original line number in ct (get existing line count before new lines are added to node).
    // this loop inserts comment lines and sets ict for all lines (including inserted comments), nothing else.
    for i := 0; i < len(lines); i++ {
    
         // if this is a root chunk and the first line,
	 // insert a 'don't edit' message.
	 if isroot && i == 0 && prog != nil && prog.Cmtmark != "" {
	   // make the comment and insert it as first line
	   comment := prog.Cmtmark + " automatically generated, DON'T EDIT. please edit " + ctfile + " from where this file stems."
	   out = append(out, Line{comment, -1})
	 }

         line := lines[i]

         // go through the  lines.
         
         // is this line the beginnig of a chunk?
	 // (subtracting the added lines, are there prevlines
	 // for this line index?)
         if _, ok := prevlines[i]; ok {
             myprevlines = prevlines[i]
         }

        // is the line a function declaration? put in
	// prevtxt starting from a line that begins with
	// the function name.
        if funcre.MatchString(line.Txt) {
        
            // get the name of the function
            matches := funcre.FindStringSubmatch(line.Txt) // sth like this?
	    funcname := matches[funcre.SubexpIndex("name")] // see https://stackoverflow.com/a/66053163

            // comments need to inherit the identation of their function declaration line, cause that isn't added later.
            // this is done apart from alreadyspace in assemble, cause functions might not be declared on the first line of their chunk, which might be intended differently.
            f := leadspacere.FindStringIndex(line.Txt)
            funcspace := line.Txt[f[0]:f[1]]

            // make a regexp for lines beginning with the function name
            funcnamere := regexp.MustCompile("^" + funcname)
            
            // skip the lines before a line starts with the function name.
	    skip := 0
            for ; skip < len(myprevlines) && !funcnamere.MatchString(myprevlines[skip].Txt); skip++ {
                // skip
            }

            // lencmt holds the length of the comment in myprevlines
            lencmt := len(myprevlines) - skip

	    // if the function declaration comes before the comment
	    // insert it here
	    if (!cmtbefore) {
	        out = append(out, line)
	    }

            // insert opening comment mark, if given.
            if prog.Cmtopen != "" {
	        cmt := funcspace + prog.Cmtindent + prog.Cmtopen
                out = append(out, Line{cmt, -1})
            }

            // insert the comment lines.
            for j := 0; j < lencmt; j++ {
	    	// figure out the comment mark during the comment.
		// if it's a multiline comment, take cmtduring
		// (if there). if it's not a multiline comment
		// take cmtmark.
		cmtmark := ""
		if prog.Cmtopen != "" { // multiline comment
		   cmtmark = prog.Cmtduring
		} else { // single lines of comment
		   cmtmark = prog.Cmtmark
		}
		
		// make the comment. 
		cmt := funcspace + prog.Cmtindent + cmtmark +
		   " " + myprevlines[skip + j].Txt
                ict := myprevlines[skip + j].Ict
                
                // insert the comment line
                out = append(out, Line{cmt, ict})
            }
            
            // insert the closing comment mark, if given. 
            if prog.Cmtclose != "" {
                cmt := funcspace + prog.Cmtindent + prog.Cmtclose
		out = append(out, Line{cmt, -1})
            }

	    // if the function declaration comes after the comment,
	    // insert it here.
	    if cmtbefore {
	        out = append(out, line)
	    }

        } else {
	  // it's a normal line, append it.
	  out = append(out, line)
	}
    }
    return out
}


// donteditcmt inserts a don't edit comment as first line.
// it assumes that the lines come from a root node
// func donteditcmt(lines []Line) []Line {



// put puts a chunk to its node in the tree under relative or absolute path.
func put(chunk *Chunk) {

    // remove leading and trailing blank lines of between-text
    //chunk.Txta = trimlines(chunk.Txta)
    //chunk.Txtb = trimlines(chunk.Txtb)

    path := getname(chunk.Tag.Txt)
    //debug("put(" + path + ")")

    // create a ghostnode if called for.
    if (path == "." || path == "") && openghost != nil {
        currentnode = openghost

        // we enter the ghost node the first time, this implicitly declares it.
        currentnode.d = true
        
        openghost = nil // necessary?
    } else {
        // named node (new or append) or ghost node (append).

        // if the path would need a node to cling to but there isn't one.
        if currentnode == nil && ! isfromroot(path) {
            fmt.Printf("error (line %d): there's no file to attach '%s' to, should it start with '//'?\n", chunk.Tag.Ict, path)
            os.Exit(-1)
	}

        // a colon at the path end indicates that this is a declaration.
	r1 := regexp.MustCompile(":\\s*$")
        isdeclaration := r1.MatchString(path)

        // remove the colon from path.
	r2 := regexp.MustCompile(":\\s*$")
        path := r2.ReplaceAllString(path, "")

        // find the node, if not there, create it.
        node := cdmk(currentnode, path, chunk.Tag.Ict)
	//r := cdroot(node, 0)
	//printtree(r)


        /* we'd like to check that a node needs to have been declared with : before text can be appended to it. for that, it doesn't help to check if a node is there, cause it might have already been created as a parent of a node. so we introduce the node.d property. */

        if isdeclaration && node.d {
            fmt.Printf("error (line %d): chunk %s has already been declared, maybe drop the colon ':'\n", chunk.Tag.Ict, path)
            os.Exit(-1)
        } else if !isdeclaration && !node.d {
            fmt.Printf("error (line %d): chunk %s needs to be declared with ':' before text is appended to it\n", chunk.Tag.Ict, path)
	    	   
	    //fmt.Printf("node.d: %s, node.name: %s\n", node.d, node.name)
            os.Exit(-1)
	}

        // remember that the node has been declared.
        if isdeclaration {
            node.d = true
	}

        // all should be well, we can set the node as the current node.
        currentnode = node
    }
    // append the chunk to the current node.
    addcreatechilds(currentnode, chunk)
}

/* cdmk walks the path from node and creates nodes if needed along the way.
 it returns the node it ended up at */
// don't pass node as a pointer to not change it?
func cdmk(n *node, path string, ict int) *node {

    /* if our path is absolute (starting from a root), we can't just jump to the root, because when changing positions in the tree, we need to make sure that ghostnodes are exited properly.  cdone exits ghostnodes properly.  here we call cdroot, which in turn recursively calls cdone to step out of each node.  */
    re := regexp.MustCompile("^/")
    if re.MatchString(path) {
        //fmt.Printf("node: %s\n", n)
        //fmt.Printf("cdroot from %s\n", n)
        // exit open ghost nodes along the way
        n = cdroot(n, ict)
    }
    
    // if the path starts with // we might need to change roots.
    r2 := regexp.MustCompile("^//")
    if r2.MatchString(path) {

        // remove the leading // of root path
        path = strings.Trim(path, "/")
    
        // split the path
        p := strings.Split(path, "/")

        // the first part of the path is the rootname
        rootname := p[0]

        // root not there? create it
        if _, ok := roots[rootname]; !ok {
	    //fmt.Printf("creating root %s\n", rootname)
            roots[rootname] = newnode(rootname, nil)
	}

        // set the node to the root
        n = roots[rootname]

        // stitch the rest of the path together to walk it
        path = strings.Join(p[1:], "/")
	//fmt.Printf("path: %s\n", path)
    }
    // for absolute paths, we should be at the right root now

    // remove leading / of absolute path
    path = strings.Trim(path, "/")

    // follow the path
        
    elems := strings.Split(path, "/")

    search := false // search for the next name
    for _, elem := range elems {
        // do we start a sub-tree search?
        if elem == "*" {
            search = true
            continue
        }
	if search == true {
            // search for the current name
            search = false // reset
            var res []*node
            bfs(n, elem, &res) // search elem in node's subtree
            if len(res) > 1 {
                fmt.Printf("error (line %d): more than one nodes named %s in sub-tree of %s\n", ict, elem, pwd(n))
                os.Exit(-1)
            } else if (len(res) == 0) {
                fmt.Printf("error (line %d): no nodes named %s in sub-tree of %s\n", ict, elem, pwd(n))
                os.Exit(-1)
            } else {
                n = res[0]
            }
            continue
        }
        // standard:
        // walk one step
        walk := n.cdone(elem)
        // if child not there, create it
        if walk == nil {
            walk = createadd(elem, n)
        }
	n = walk
    }
    // print("put return: " + str(n.name))
    return n // the node we ended up at
}


// bfs breath-first searches for all nodes named 'name' starting from 'node' and puts them in 'out'
func bfs(node *node, name string, out *[]*node) {
    //fmt.Printf("bfs for %s in %s\n", name, node.name)
    if node.name == name {
        //fmt.Println("found")
        *out = append(*out, node)
    }
    // search the node's childs
    for _, childname := range node.ls() {
        bfs(node.childs[childname], name, out)
    }
    // do we need to search the gostchilds?
    for _, child := range node.ghostchilds {
        bfs(child, name, out)
    }
}

// cdroot cds back to root. side effect: ghosts are exited
func cdroot(node *node, ict int) *node {
    if node == nil { return nil }
    if node.isroot() { // we're at a root
        return node
    }
    // continue via the parent
    return cdroot(node.cdone(".."), ict) // it's probably not very necessary to pass the ict here, cause it only check's that the step isn't a '#' that would walk into a ghostnode
}

// pwd: print the path to a node starting from its root
func pwd(node *node) string {
    out := node.name
    walk := node
    // append the name of each parent node to the left side of path
    for !walk.isroot() { // todo maybe say while not walk.isroot()
        walk = walk.parent
        out = walk.name + "/" + out
    }
    // append the file marker
    if _, ok := roots[walk.name]; ok { // is root check necessary?
        out = "//" + out
    }
    return out
}

// createadd creates a named or ghost node and adds it to its parent
func createadd(name string, parent *node) *node {

    node := newnode(name, parent)
    // debug(f"createadd: {pwd(node)}")
    
    // if we're creating a ghost node
    if isghost(node.name) { 
        // debug(f"creating a ghost child for {parent.name}")
        // add it to its parent's ghost nodes
        parent.ghostchilds = append(parent.ghostchilds, node)
    } else {
        // we're creating a name node
        
        /* if the parent is a ghost node, this node could have already been created before with its non-ghost path (an earlier chunk in the codetext might have declared it and put text into it, with children/ghost children, etc), then we move it as a named child from the last named parent to here */
        /* if a node with this name is already child of last named parent, move it here */
        if isghost(parent.name) {
            lnp := lnp(node)
            if _, ok := lnp.childs[node.name]; ok { 
                node = lnp.childs[name]
                delete(lnp.childs, name)
                node.parent = parent
	    }
	}

        // add named node to parent, if it was created or moved
        parent.childs[name] = node
    }
    return node
}

// addcreatechilds adds a chunk to a node and creates the children from the tags in the code.
// this is the only place where chunks, and this way, text, gets added to nodes.
func addcreatechilds(n *node, chunk *Chunk) {

    // reset the open ghost.  why reset the openghost here?  not so clear.  but we need to reset it somewhere, that only the direct next code chunk can fill a ghost node.
    openghost = nil 
    
    // map from the ct lines to this node.
    for _, line := range chunk.Txta {
      nodeatict[line.Ict] = n
    }
    for _, line := range chunk.Code {
      nodeatict[line.Ict] = n
    }
    for _, line := range chunk.Txtb {
      nodeatict[line.Ict] = n
    }

    // map from the opening and closing lines to this node.
    // opening line
    nodeatict[chunk.Tag.Ict] = n             
    // closing line
    nodeatict[chunk.Tag.Ict + len(chunk.Code)+1] = n

    // generate the child nodes
    for i, line := range chunk.Code {
        if !isname(line.Txt) {
            continue
	}
	
        // why do we create the children when adding the chunks?
	// maybe because here we know where childs of ghost nodes end up in the tree. 

        // the newly created child
        var child *node

        name := getname(line.Txt)
        if name == "." { // ghost child
            // if we're not at the first ghost chunk here
            if openghost != nil {
                fmt.Printf("error (line %d): only one ghost child per text chunk allowed\n", line.Ict)
                os.Exit(-1)
	    }
            // create a ghost chunk
            // it's name is a dot followed by it's index in the parent's ghostchilds array
            // openghost = createadd(GHOST, n)
            openghost = createadd("." + itoa(len(n.ghostchilds)), n)
            child = openghost
        } else { // we're at a name
            // if name not yet in child nodes
            if _, ok := n.childs[name]; !ok {  
                // create a new child node and add it
                child = createadd(name, n)
            }
	}

	// the child has been referenced (not technically necessary to set this for ghost childs?)
	child.r = true
	
        // at which line of the parent is the child?
        child.iip = i            // todo was i + len(n.lines), why?

        // at this line, the parent has a child
        n.caict[chunk.Tag.Ict + i] = child

        /* we're just appending the nth chunk of this node, this is the
         chunk that references to the child (used for linking to the
         specific parent chunk in doc) */
        child.chup = len(n.chunks)
    }
    // add the chunk.  
    // do this after child.chup was set in the loop
    n.chunks = append(n.chunks, chunk)
}

// makelines turns a slice of strings into a slice of Lines counting up their index in the ctfile
func makelines(a []string, ict int) []Line {
    out := []Line{}
    for i, s := range a {
      out = append(out, Line{s, ict+i})
    }
    return out
}

// lnp returns the last named parent node
func lnp(node *node) *node {
    if node == nil { return nil }

    if !isghost(node.name) { return node }
    return lnp(node.parent)
}

// printtree: print node tree recursively
func printtree(node *node) {
    fmt.Printf("printtree of %s\n", node.name)
    fmt.Printf("ls: %s\n", node.ls())
    for _, name := range node.ls() {
        printtree(node.childs[name])
    }
    for _, child := range node.ghostchilds {
        printtree(child)
    }
}



// Ct runs codetext
func Ct(text string, ctfile string) error {

    // load the config
    conf, err := loadconf()
    if err != nil {
        return err
    }

    // ctok := true  // todo listen to put?

    // reset variables
    roots = make(map[string]*node)
    roottext = make(map[string]string)
    currentnode = nil
    openghost = nil
    ictmap = make(map[string]map[int]int)
    nodeatict = make(map[int]*node)
    ctlines = ctlines[:0]
    chop = make(map[int]bool)
    chclo = make(map[int]bool)

    // f.readlines() # readlines keeps the \n for each line, 
    // take care of dos line breaks \r\n
    lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
    
    // save the lines, for totex or so
    ctlines = lines
    
    // put in the chunks

    // are we in code?
    incode := false
    // current collected text
    txt := []Line{}
    // number of chunks encountered
    n := 0
    // the current chunk    
    var currentchunk *Chunk
    // dbltickre says that a line could be chunk-opening or chunk-closing, and excludes text separating ``= lines.
    dbltickre := regexp.MustCompile("^``[^=`]*")
    
    for i, txtline := range lines {
        //print("line: " + txtline)
        line := Line{Txt: txtline, Ict: i}

        // we need to keep track whether we're in code or not cause two ticks `` could both close and open (an unnamed) chunk.
        if dbltickre.MatchString(line.Txt) && incode == false { // we're are at the beginning of code
	    incode = true
	    // if there was a preceeding chunk, put it.
	    if currentchunk != nil {
	        put(currentchunk)
		n++
            }
	    // make a new current chunk
            currentchunk = &Chunk{Nct:n+1}
	    // whatever preceeding text was collected in txt, save it as this chunk's txta
	    for _, l := range txt { currentchunk.Txta = append(currentchunk.Txta, l) }
	    // clear text collection
	    txt = []Line{}
	    // remember the tag
	    currentchunk.Tag = line
        } else if isdblticks(line.Txt) { // we're at the end of code
            incode = false
	} else if istxtsep(line.Txt) { // we're at a text sep line ``=
	    // append whatever was collected as text before ``= to the current chunk's txtb
	    for _, l := range txt { currentchunk.Txtb = append(currentchunk.Txtb, l) }
	    // clear text collection
	    txt = []Line{}
        } else if incode { // we're in code
	    // append to the code
            currentchunk.Code = append(currentchunk.Code, line)
	} else { // we're in text between code
            // collect the text
	    txt = append(txt, line)
        }
    }
    // the last current chunk wasn't put yet, put it.
    if currentchunk != nil {
        // append collected text as this chunk's txtb
        for _, l := range txt { currentchunk.Txtb = append(currentchunk.Txtb, l) }
	// put the chunk
	put(currentchunk)
    }

    /* in the end, exit un-exited ghost nodes on the way from
    currentnode to root by calling cdroot a last time. (exiting ghost
    nodes puts their named children into the child list of the last
    named parent, from where they can be accessed). */

    cdroot(currentnode, 0)

    // check that no references or declarations are missing.
    refok := true
    declok := true
    for _, root := range roots {
        if ok := checkref(root); !ok { refok = false }
	if ok := checkdecl(root); !ok { declok = false }
    }
    // don't continue if something's wrong.
    if !refok || !declok {
        return fmt.Errorf("chunk refs not working out.")
    }
    
    // at the end, write all files (each file is a root) 
    for _, filename := range keys(roots) {
        // todo: add don't edit comment like before

        // get the proglang. for now from the filename, maybe later also from hashtag, in case filename has no suffix
        proglang := filepath.Ext(filename)
        // remove the dot at the beginnig of the string Ext returns
        proglang = strings.Replace(proglang, ".", "", 1)
        
        // assemble the code
        //out := []Line{}
	out := assemble(roots[filename], "", filename, proglang, ctfile, conf)
	//debug(out)
	// check what node tree was generated
        //printtree(roots[filename])
	
	// map from the line number in the generated source to the original line number in the ct
	for i, line := range out {
	    ictmap[filename][i] = line.Ict
	}
	
	// concat the output text for this root
	outtxt := ""
	for _, line := range out {
	    outtxt += line.Txt + "\n"
	}
        
        // save the generated text
        roottext[filename] = outtxt // todo error out is a tuple?
    }
    return nil // ok
}

// checkref checks that each node except root nodes has been referenced
func checkref(n *node) bool {
    ok := true
    // if the node is not root and hasn't been referenced, error
    if !n.isroot() && !n.r {
        fmt.Printf("error: node %s hasn't been referenced from another node.\n", pwd(n)) // todo give line number? what about empty chunks where n.ict[0] wouldn't work? could put pass the ctline of the chunk opening, and node save this in a property, in case no lines get added to the node?
	ok = false
    }
    // check the children
    for _, child := range n.childs {
        if childok := checkref(child); !childok {
            ok = false
	}
    }
    return ok
}

// checkdecl checks that each node except ghost nodes has been declared
// doubles with checks in put(), but necessary, cause there could just be a reference to a chunk that's never opened, put() wouldn't catch this.
func checkdecl(n *node) bool {
    ok := true
    if !n.d {
        fmt.Printf("error: node %s hasn't been declared.\n", pwd(n))
	ok = false
    }
    // check the childs
    for _, child := range n.childs {
        if childok := checkdecl(child); !childok {
	    ok = false
	}
    }
    return ok // is the tree hanging on this node ok?
}

// Ctwrite runs codetext and writes the assembled files        
func Ctwrite(text string, dir string, ctfile string) error {
    //fmt.Printf("hello ctwrite\n")
    
    // run codetext
    err := Ct(text, ctfile)
    if err != nil { return err }

    // write the assembled text for each root
    for filename, _ := range roots {

        //fmt.Printf(filename)
        
        // assemble the code
        txt := roottext[filename]
        // printtree(roots[filename])

        path := filename
        // and write it to file
        if dir != "" { // ok so?
            path = dir + "/" + filename
        }
	f, _ := os.Create(path)
	defer f.Close()
	_, _ = f.WriteString(txt)
	
        // say which file was written
        fmt.Println(path)
    }
    return nil // no error
}

// Ict returns at which line in the ct file a line from a generated file is, zero-indexed.
func Ict(genfile string, igen int) (int, error) {
    // fmt.Println(ictmap)
    if _, ok := ictmap[genfile]; !ok {
        return 0, errors.New(fmt.Sprintf("there is no %s", genfile))
    }
    if _, ok := ictmap[genfile][igen]; !ok {
        fmt.Println("ictmap: ")
	/*for k, v := range ictmap[genfile] {
	    fmt.Printf("%d: %d; ", k, v)
	}*/
	fmt.Println(ictmap[genfile])
	// count up plus one to map from 0-based to 1-based
        return 0, errors.New(fmt.Sprintf("there is no line %d in %s.", igen + 1, genfile))
    }
    return ictmap[genfile][igen], nil
}

// Chop says whether a ct line is a chunk opening
func Chop(line int) bool {
    if _, ok := chop[line]; !ok { return false }
    return chop[line]
}

// Chclo says whether a ct line is a chunk close
func Chclo(line int) bool {
    if _, ok := chclo[line]; !ok { return false }
    return chclo[line]
}

// itoa converts int to string
func itoa(i int) string {
    a := strconv.Itoa(i)
    return a
}

// HelloCt says hello
func HelloCt() {
  fmt.Println("hello ct")
}

// debug prints s
func debug(a any) {
    fmt.Println(a)
}

// getpl gets the entry for a programming language from conf
func getpl(conf *Conf, pl string) *Prog {
    for _, prog := range conf.Proglang {
        // does the name match?
        if prog.Name == pl {
	    return &prog
	}
	// do any of the extensions match?
	for _, ext := range prog.Ext {
	    if ext == pl {
	        return &prog
            }
	}
    }
    return nil
}

