package ct

import (
  "cmp"
  "errors"
  "fmt"
  "os"
  "regexp"
  "slices"
  "strconv"
  "strings"  
)

// variables

// save the ctlines for printtex
// lines keep their \n
var ctlines []string

// for each generated file, map its line numbers to the original line numbers in ct file
var ctlinenum map[string]map[int]int

// currentnode is the node we're currently at
var currentnode *node

// openghost is the open ghost node. if the last chunk opened a ghostnode, it's this one
var openghost *node

// roots holds the root nodes
var roots map[string]*node

// roottext holds the generated code for each root
var roottext map[string]string

// nat (node-at-ct-line) holds the node at a specific ct line (if there is one) was nodeatctline
var nat map[int]*node

// is this ctline opening a chunk? was ischunkopening
var chop map[int]bool 

// is this ctline closing a chunk? was ischunkclose
var chclo map[int]bool 



// a node holds lines of code belonging to multiple code chunks sharing the same path.
type node struct {

    /* the node's name.

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
    
    // lines: keep the text as lines instead of a text block, cause splitting on '\n' on empty text gives length 1 (length 0 is wanted)
    lines []string

    // ctline: for each text line, ctline holds its original line num in ct file, zero-indexed
    ctline map[int]int // was ctlinenum

    // d: has this node been declared with a colon ':'. every node except ghost nodes needs to have been declared.
    d bool

    // r: has this node been referenced. every node except root nodes needs to have been referenced.
    r bool

    // lip (line in parent): at which line of the parent is this node?
    lip int // was lineinparent

    // chat (child at ct line): the child (named or ghost) at this ct line (if any)
    chat map[int]*node // was childatctline

    // nchunks: the number of code chunks appended to this node
    nchunks int

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
    n.ctline = make(map[int]int)
    n.chat = make(map[int]*node)
    //n.lines = []string{}
    return &n    
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

/* assemble assembles a codechunk recursively, filling up its leading
 space to leadingspace. this way we can take chunks that are already
 (or partly) indented with respect to their parent in the editor, and
 chunks that are not.  */
func assemble(n *node, leadingspace string, rootname string, genlinenum int) (string, int) {

    var lastnamedp *node

    // if it's a ghost node, remember the last named parent up the tree
    if isghost(n.name) {
        lastnamedp = lnp(n)
    }    
    /*
    find out a first line how much this chunk is alredy indented
    and determine how much needs to be filled up
    */
    var alreadyspace string
    // leading space already there (in first line)
    if len(n.lines) > 0 {
        re := regexp.MustCompile("^\\s*")
	f := re.FindStringIndex(n.lines[0])
        alreadyspace = n.lines[0][f[0]:f[1]]
    } else {
        alreadyspace = "" // no line, so no leading space already there
    }
    // space that needs to be added
    addspace := ""
    if len(leadingspace) > len(alreadyspace) {
        addspace = leadingspace[0:len(leadingspace)-len(alreadyspace)]
    }

    // if the rootname isn't in ctlinenum yet, put it there
    if _, ok := ctlinenum[rootname]; !ok {
        ctlinenum[rootname] = make(map[int]int)
    }

    out := ""
    ighost := 0
    var outnew string
    // for line in lines:
    for i, line := range n.lines {
        if isname(line) {

            // remember leading whitespace
	    re := regexp.MustCompile("^\\s*")
            childleadingspace := re.FindString(line) + addspace 
            name := getname(line)
            if name == "." {   // assemble a ghost-child
                outnew, genlinenum = assemble(n.ghostchilds[ighost], childleadingspace, rootname, genlinenum) 
                out += outnew //  + "\n" # why add \n?
                ighost += 1
            } else {             // assemble a named child
	        var child *node
                if isghost(n.name) {
                    // if we're at a ghost node, we get to the child via the last named ancestor
                    child = lastnamedp.childs[name]
                } else {
                    child = n.childs[name]
		}
		outnew, genlinenum = assemble(child, childleadingspace, rootname, genlinenum)
                out += outnew // + "\n" # why add \n?
	    }
        } else { // normal line
            out += addspace + line + "\n"
            // map from the line number in the generated source to the original line number in the ct
            ctlinenum[rootname][genlinenum] = n.ctline[i] 
            genlinenum += 1 // we added one line to root, so count up
	}
    }
    // return the generated text and the new root line number (should be the same as the number of lines in out, so maybe don't return it?)
    return out, genlinenum
}


// put puts text in tree under relative or absolute path
func put(path string, text string, ctlinenum int) {
    //debug("put(" + path + ")")

    // create a ghostnode if called for
    if (path == "." || path == "") && openghost != nil {
        currentnode = openghost

        // we enter the ghost node the first time, this implicitly declares it
        currentnode.d = true
        
        openghost = nil // necessary?
    } else {
        // named node (new or append) or ghost node (append)

        // if the path would need a node to cling to but there isn't one
        if currentnode == nil && ! isfromroot(path) {
            fmt.Printf("error (line %d): there's no file to attach '%s' to, should it start with '//'?\n", ctlinenum, path)
            os.Exit(-1)
	}

        // a colon at the path end indicates that this is a declaration
	r1 := regexp.MustCompile(":\\s*$")
        isdeclaration := r1.MatchString(path)

        // remove the colon from path
	r2 := regexp.MustCompile(":\\s*$")
        path := r2.ReplaceAllString(path, "")

        // find the node, if not there, create it
        node := cdmk(currentnode, path, ctlinenum)
	//r := cdroot(node, 0)
	//printtree(r)


        /* we'd like to check that a node needs to have been declared with : before text can be appended to it. for that, it doesn't help to check if a node is there, cause it might have already been created as a parent of a node. so we introduce the node.d property. */

        if isdeclaration && node.d {
            fmt.Printf("error (line %d): chunk %s has already been declared, maybe drop the colon ':'\n", ctlinenum, path)
            os.Exit(-1)
        } else if !isdeclaration && !node.d {
            fmt.Printf("error (line %d): chunk %s needs to be declared with ':' before text is appended to it\n", ctlinenum, path)
	    	   
	    //fmt.Printf("node.d: %s, node.name: %s\n", node.d, node.name)
            os.Exit(-1)
	}

        // set that the node has been declared
        if isdeclaration {
            node.d = true
	}

        // all should be well, we can set the node as the current node
        currentnode = node
    }
    // append the text to node
    concatcreatechilds(currentnode, text, ctlinenum)
}

/* cdmk walks the path from node and creates nodes if needed along the way.
 it returns the node it ended up at */
// don't pass node as a pointer to not change it?
func cdmk(n *node, path string, ctlinenum int) *node {

    /* if our path is absolute (starting from a root), we can't just jump to the root, because when changing positions in the tree, we need to make sure that ghostnodes are exited properly.  cdone takes care of that, so we go backward node by node with cdone.  cdroot does this recursively. */
    re := regexp.MustCompile("^/")
    if re.MatchString(path) {
        //fmt.Printf("node: %s\n", n)
        //fmt.Printf("cdroot from %s\n", n)
        // exit open ghost nodes along the way
        n = cdroot(n, ctlinenum)
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
                fmt.Printf("error (line %d): more than one nodes named %s in sub-tree of %s\n", ctlinenum, elem, pwd(n))
                os.Exit(-1)
            } else if (len(res) == 0) {
                fmt.Printf("error (line %d): no nodes named %s in sub-tree of %s\n", ctlinenum, elem, pwd(n))
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
func cdroot(node *node, ctlinenum int) *node {
    if node == nil { return nil }
    if node.isroot() { // we're at a root
        return node
    }
    // continue via the parent
    return cdroot(node.cdone(".."), ctlinenum) // it's probably not very necessary to pass the ctlinenum here, cause it only check's that the step isn't a '#' that would walk into a ghostnode
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

// concatcreatechilds concatenates text to node and creates children from text (named or ghost)
// this is the only place where text gets added to nodes
func concatcreatechilds(n *node, text string, ctlinenum int) {

    openghost = nil // why here? not so clear. but we need to reset it somewhere, that only the direct next code chunk can fill a ghost node

    // replace the last \n so that spli doesn't produce an empty line at the end
    re := regexp.MustCompile("\n$")
    text = re.ReplaceAllString(text, "")
    newlines := strings.Split(text, "\n")

    // map from the line number in node to original line number in ct (get existing line count before new lines are added to node)

    N := len(n.lines) 
    for i, _ := range newlines {
        // go through the new lines
        n.ctline[N+i] = ctlinenum + i
        // map from the ct line to the node
        nat[ctlinenum + i] = n
    }
    
    // also set nat for the opening and the closing line of a chunk 
    // opening line
    nat[ctlinenum - 1] = n
    // closing line
    nat[ctlinenum + len(newlines)] = n
        
    //debug("N: " + str(N))
    //debug("node.ctlinenum: " + str(n.ctlinenum))

    // put the new lines into node
    n.lines = append(n.lines, newlines...)

    // generate the child nodes
    for i, line := range newlines {
        if !isname(line) {
            continue
	}
        // why do we create the children when concating text? maybe because here we know where childs of ghost nodes end up in the tree. """

        // the newly created child
        var child *node

        name := getname(line)
        if name == "." { // ghost child
            // if we're not at the first ghost chunk here
            if openghost != nil {
                fmt.Printf("error (line %d): only one ghost child per text chunk allowed\n", ctlinenum+i)
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
        child.lip = i+N

        // at this line, the parent has a child
        n.chat[ctlinenum+i] = child

        /* we're just appending the nth chunk of this node, this is the
         chunk that references to the child (used for linking to the
         specific parent chunk in doc) */
        child.chup = n.nchunks
    }
    // we've appended a codechunk to the node, so increase the number of chunks
    // do this after child.chup was set in the loop
    n.nchunks += 1
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
func Ct(text string) bool {

    // ctok := true  // todo listen to put?

    // reset variables
    roots = make(map[string]*node)
    roottext = make(map[string]string)
    currentnode = nil
    openghost = nil
    ctlinenum = make(map[string]map[int]int)
    nat = make(map[int]*node)
    ctlines = ctlines[:0]
    chop = make(map[int]bool)
    chclo = make(map[int]bool)

    // f.readlines() # readlines keeps the \n for each line, 
    lines := strings.Split(text, "\n")
    // put the \n that split removed back to each line, text concat in nodes relies on that.
    for i, _ := range lines {
        lines[i] += "\n"
    }
    
    // save the lines, for totex or so
    ctlines = lines
    
    // put in the chunks

    // are we in chunk?
    inchunk := false
    // current chunk content
    chunk := ""
    // current chunk name/path
    var path string
    // start line of chunk in ct file
    chunkstart := 0
    
    for i, line := range lines {
        //fmt.Print(line)

        /* we can't decide for sure whether we're opening or closing a chunk by looking at the backticks alone, cause an unnamed chunk is opend the same way it is closed.  so in addition, check that inchunk is false. */
	re := regexp.MustCompile("^``[^`]*")
        if re.MatchString(line) && inchunk == false {
	    //debug("in chunk")
            // we're in a chunk
            inchunk = true
            // remember its path
            path = getname(line)
            // remember the start line of chunk in ct file
            // add one for the chunk text starts in the next line, not this
            // (treat the line numbers as 0-indexed)
            chunkstart = i+1 

            // remember that this line is opening a chunk
            chop[i] = true
            
        } else if isdblticks(line) { // at the end of chunk save chunk
	    //debug("out of chunk")
            inchunk = false
            // debug(f"calling put for: {path}")
            // debug("split chunk: " + str(chunk.split("\n")))
            put(path, chunk, chunkstart)
            // reset variables
            chunk = ""
	    path = ""

            // remember that this line is closing a chunk
            chclo[i] = true

        } else if inchunk { // when we're in chunk remember line
            chunk += line
	}
        //else:
        //    debug(line, end="") # for debugging
    }

    /* in the end, exit un-exited ghost nodes on the way from
    currentnode to root by calling cdroot a last time. (exiting ghost
    nodes puts their named children into the child list of the last
    named parent, from where they can be accessed). */

    cdroot(currentnode, 0)

    refok := true
    declok := true
    // check that no references or declarations are missing.
    for _, root := range roots {
        if ok := checkref(root); !ok { refok = false }
	if ok := checkdecl(root); !ok { declok = false }
    }
    // don't continue if something's wrong.
    if !refok || !declok {
        fmt.Println("aborting.")
        return false
    }
    
    // at the end, write all files (each file is a root) 
    for _, filename := range keys(roots) {
        // todo: add don't edit comment like before
        
        // assemble the code
        out, _ := assemble(roots[filename], "", filename, 0)
        // printtree(roots[filename])
        
        // save the generated text
        roottext[filename] = out // todo error out is a tuple?
    }
    return true
}

// checkref checks that each node except root nodes has been referenced
func checkref(n *node) bool {
    ok := true
    // if the node is not root and hasn't been referenced, error
    if !n.isroot() && !n.r {
        fmt.Printf("error: node %s hasn't been referenced from another node.\n", pwd(n)) // todo give line number? what about empty chunks where n.ctline[0] wouldn't work? could put pass the ctline of the chunk opening, and node save this in a property, in case no lines get added to the node?
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
func Ctwrite(text string, dir string) {
    //fmt.Printf("hello ctwrite\n")
    
    // run codetext
    ok := Ct(text)
    if !ok { return }

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
}

// Ctline returns at which line in the ct file a line from a generated file is, zero-indexed.
func Ctline(genfile string, genline int) (int, error) {
    if _, ok := ctlinenum[genfile]; !ok {
        return 0, errors.New(fmt.Sprintf("there is no %s", genfile))
    }
    if _, ok := ctlinenum[genfile][genline]; !ok {
        fmt.Println("ctlinenum: ")
	/*for k, v := range ctlinenum[genfile] {
	    fmt.Printf("%d: %d; ", k, v)
	}*/
	fmt.Println(ctlinenum[genfile])
        return 0, errors.New(fmt.Sprintf("there is no line %d in %s.", genline+1, genfile))
    }
    return ctlinenum[genfile][genline], nil
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
func debug(s string) {
    fmt.Println(s)
}