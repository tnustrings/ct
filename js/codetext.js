// codetext.js implements codetext in js

// implemented after tangle.py

const fs = require("fs")

console.log("hi")

main()

// main offers command line interface
function main() {
    var file = process.argv[2]
    var text = fs.readFileSync(file) + ""

    // normal compiling
    if (process.argv.length == 3) {
	assembletext(text)
    } else if (process.argv.length == 5) {
	// map from line number in generated source to original line number in ct
	// name of generated file
	var genfile = process.argv[3]
	// line number of generated file
	var genlinenr = parseInt(process.argv[4])
	assembletext(text)
	// print the original line number in ct file
	console.log(ctlinenr[genfile][genlinenr])
    }
}

// assembletext assembles codetext
function assembletext(lines) {
    var lines = text.split("\n")

    for (line of lines) {
	// only look at declaration lines
	if (!isdeclaration(line)) { continue }
	name = getname(line)

	// we're at a root if the name starts with a //
        if (isroot(name)) {
	    // maybe an alias follows the file name, seperated by ': '
            parts = name.split(": ")

	    // remove leading slashes of root name
            filename = parts[0].replace(/^\/\//, "")
            // take the first part of path as filename/alias. todo: split at //?
            filename = filename.split("/")[0]

	    // the root is already created? continue.

	    if (filename in roots) {
                continue
	    }

	    // is it an alias? continue. (this assumes aliases need to appear first in the text before they can be referenced.
            if (filename in fileforalias) {
                continue
	    }

	    roots[filename] = new node(filename, null)

	    if (parts.length > 1) {
		alias = parts[1]
		// alias already used for other file
                if (alias in fileforalias && fileforalias[alias] != filename) {
                    print(f"error: alias {alias} already references file {fileforalias[alias]}")
                    process.exit()
		}
		// alias it
                fileforalias[alias] = filename
	    }
	} // end of isroot
    } // end of line loop

    // no files, exit
    if (roots.length == 0) {
	process.exit()
    }

    // print("roots: " + str(roots))

    // now that we got the roots, we can start putting in the chunks. 

    var inchunk = false // are we in a chunk
    var chunk = "" // current chunk content
    var path = null // current chunk name/path
    var chunkstart = 0 // start line of chunk in ct file

    // for (line of lines) {
    for (var i = 0; i < lines.length; i++) {
	var line = lines[i]
        if (isdeclaration(line)) { // at the beginning of chunk remember its name
            inchunk = true
            // print("#getname 4")
            path = getname(line)
	    // remember the start line of chunk in ct file
            // add two: one, for line numbers start with one not zero, another, for the chunk text starts in the next line, not this
	    chunkstart = i+2
	} else if (isend(line)) {
	    inchunk = false
            // print(f"calling put for: {path}")
            put(path, chunk, chunkstart)
            // reset variables
            chunk = ""
            path = null
            // print("")
	} else if(inchunk) { // when we're in chunk remember line
	    chunk += line
	} else {
            // console.log(line) // for debugging
	}

    }

    /* in the end we need to exit all un-exited ghost nodes so that their
    named children end up as the named children of the last named parent
    where we can access them.  */

    cdroot(currentnode)

    // at the end, write all files (each file is a root) 
    for (filename of roots) {
        // todo: add don't edit comment like before
        
        // assemble the code
        [out, _] = assemble(roots[filename], "", filename, 1)
        // printtree(roots[filename])

	// and write it to file
        // print(f"write {filename}")
	fs.writeFile(filename, out)
    }
}



// isdeclaration returns true if line is the declaration line of a code chunk
function isdeclaration(line) {
    return line.match(/^<<[^\>]*$/)
}


// getname gets the chunkname from a declaration or in-chunk line
function getname(line) {
    name = line.replace(/.*<</, "")
    name = name.replace(/>>.*/, "") // chunk declarations do not have this
    name = name.replace(/\n$/, "")
    // print(f"getname({line}): '{name}'")

    return name
}

// isroot says whether the name is a root
function isroot(name) {
    return name.match(/^\/\//)
}

// getroot returns root referenced by path and chops off root in path, except if there's only one un-aliased root.
function getroot(path) {
    // remove the leading // of root path
    path = path.replace(/^\/+/, "")
    
    // split the path
    p = path.split("/")

    // allow splitting file paths and subsequent chunk paths by //?
    // p = path.split("//")
    // return (resolveroot(p[0]), p[1])

    return [resolveroot(p[0]), "/".join(p.slice(1))] // slice(1): all elements including and after index 1
}

// resolveroot returns the rootnode for a filename or a alias
function resolveroot(name) {
    if (name in roots) {// name is a filename
        return roots[name]
    }
    else if (name in fileforalias) { // name is an alias
        return roots[fileforalias[name]]
    }
    console.log("error: root name '" + name + "' not found")
    process.exit()
}

// node code-chunks are represented as nodes in a tree. """
class node {
    constructor(name, parent) {
        this.name = name
        this.cd = {}
        this.cd["."] = this
        if (parent == null) {
            this.cd[".."] = this
	} else {
            this.cd[".."] = parent
	}
        //self.parent = parent
        //this.text = ""
	// keep the text as lines, cause splitting on '\n' on empty text gives length 1 (length 0 is wanted)
	this.lines = []
	// the ghost children. why more than one?
        this.ghostchilds = []
	// for each text line, ctlinenr holds its original line nr in ct file
	this.ctlinenr = {}
    }
        
    // ls lists the named childs
    ls() {
	var out = []
        // return all except . and ..
	for (var key of Object.keys(this.cd)) {
	    if (k == "." || k == "..") { continue }
	    out.append(k)
	}
    }
}

var currentnode = null // the node we're currently at
var openghost = null // if the last chunk opened a ghostnode, its this one

    
// put puts text in tree under relative or absolute path
function put(path, text, ctlinenr) {
    // if at file path, take only the path and chop off the alias
    if (path.match(/^\/\//)) {
        parts = path.split(": ")
        path = parts[0]
    }

    // create a ghostnode if called for
    if ((path == "." || path == "") && openghost != null) {
        currentnode = openghost
        openghost = null // necessary?
    } else {
	// if the path would need a node to cling to but there isn't noe
        if (currentnode == null && !isroot(path)) {
            console.log("error: there's no file to attach '" + path + "' to, should it start with '//'?")
            process.exit()
	}
	// go to node, if necessary create it along the way
        currentnode = cdmk(currentnode, path)
    }

    // append the text to node
    concatcreatechilds(currentnode, text, ctlinenr)
}


// cdmk walks the path from node and creates nodes if needed along the way.
// it returns the node it ended up at
function cdmk(node, path) {

    // if file path // switch roots
    if (path.match(/^\/\//)) {
        // we need to exit open ghost nodes. cdroot does this along the way.
        cdroot(node)
        var [node, path] = getroot(path)
    }
    // if absolute path / go to root
    else if (path.match(/^\//)) {
        // we need to exit open ghost nodes. cdroot does this along the way.
        node = cdroot(node)
    }

    // remove leading / of absolute path
    path = path.replace(/^\/+/, "") // todo check

    // follow the path
        
    elems = path.split("/")

    search = false // search for the next name

    for (var elem of elems) {
        // do we start a sub-tree search?
        if (elem == "*") {
            search = true
            continue
	}
	if (search == true) {
            // search for the current name
            search = false // reset
            res = []
	    bfs(node, elem, res) // search elem in node's subtree
            if (res.length > 1) {
		console.log("error: more than one nodes named " + elem + " in sub-tree of " + pwd(node))
		process.exit()
	    }
	    else if (len(res) == 0) {
		console.log("error: no nodes named " + elem + " in sub-tree of " + pwd(node))
		process.exit()
	    }
	    else {
		node = res[0]
	    }
            continue
	}

	// standard:
        // walk one step
        walk = cdone(node, elem)
        // if child not there, create it
        if (walk == null) {
            walk = createadd(elem, node)
	}
        node = walk
    }
    // print("put return: " + str(node.name))
    return node // the node we ended up at
}

// concatcreatechilds concatenates text to node and creates children from text (named or ghost)
// this is the only place where text gets added to nodes
function concatcreatechilds(node, text) {

    openghost = null // why here? not so clear. but we need to reset it somewhere, that only the direct next code chunk can fill a ghost node

    // replace the last \n so that spli doesn't produce an empty line at the end
    text = text.replace(/\n$/, "")
    var newlines = text.split("\n")

    // map from the line number in node to original line number in ct (get existing line count before new lines are added to node)
    var N = node.lines.length
    for (var i = 0; i < newlines.length; i++) {
	node.ctlinenr[N+i] = ctlinenr + i
    }

    //node.text += text
    node.lines.push(...newlines)
    
    for (var line of newlines) {
	if (!isname(line)) { continue }

	// why do we create the children when concating text? maybe because here we know where childs of ghost nodes end up in the tree. """
	// print("#getname 2")
	name = getname(line)
	if (name == ".") {// ghost child
            // if we're not at the first ghost chunk here
            if (openghost != null) {
		console.log("error: only one ghost child per text chunk allowed")
		process.exit()
	    }
            // create the ghost chunk
            openghost = createadd(GHOST, node)
	} else {  // we're at a name
            if (!name in node.ls()) {
		createadd(name, node)
	    }
	}
    }
}

// createadd creates a named or ghost node and adds it to its parent
function createadd(name, parent) {

    var node = new node(name, parent)
    // print(f"createadd: {pwd(node)}")
    
    // if we're creating a ghost node
    if (node.name == GHOST) {
	// print(f"creating a ghost child for {parent.name}")
        // add it to its parent's ghost nodes
        parent.ghostchilds.append(node)
    } else {
        // we're creating a name node
        
        /* if the parent is a ghost node, this node could have already been created before with its non-ghost path (an earlier chunk in the codetext might have declared it and put text into it, with children/ghost children, etc), then we move it as a named child from the last named parent to here */
	
        // if a node with this name is already child of last named parent, move it here
        if (parent.name == GHOST) {
            namedp = lastnamed(node)
	    if (node.name in namedp.ls()) {
		node = namedp.cd[name]
                namedp.cd.delete(name)
                node.cd[".."] = parent
	    }
	}


	// add named node to parent, if it was created or moved
        parent.cd[name] = node
    }

    return node
}

// lastnamed returns the last named parent node
function lastnamed(node) {
    if (node == null) { return null }
    if (node.name != GHOST) { return node }
    return lastnamed(node.cd[".."])
}


// bfs breath-first searches for all nodes named 'name' starting from 'node' and puts them in 'out'
function bfs(node, name, out) {
    //  print(f"bfs {node}")
    if (node.name == name) {
        out.append(node)
    }
    // search the node's childs
    for (var childname of node.ls()) {
        bfs(node.cd[childname], name, out)
    }
    // do we need to search the gostchilds?
    for (var child of node.ghostchilds) {
        bfs(child, name, out)
    }
}

// cdone walks one step from node
function cdone(node, step) {
    if (step == GHOST) {
        console.log("error: we may not walk into a ghost node via path")
        process.exit()
    }
    if (step == "") {
        step = "."
    }
    if (step == "..") { // up the tree
        // print(f"call exitghost for {pwd(node)}")
        exitghost(node)
    }
    if (step in node.cd) {
        return node.cd[step]
    }
    return null
}

// cdroot cds back to root. side effect: ghosts are exited
function cdroot(node) {
    if (node == null) { return null }
    if (node.cd[".."] == node) { // we're at a root
        return node
    }
    // continue via the parent
    return cdroot(cdone(node, ".."))
}

// exitghost moves ghost node's named children to last named parent. needs to be called after leaving a ghost node
function exitghost(ghost) {
    // not a ghost? do nothing
    if (ghost == null || ghost.name != GHOST) {
        return
    }

    /* if we exit a ghost node, we move all its named childs to the ghost node's parent and let the ghostnode be the childs' ghostparent (from where they can get e.g. their indent) */
    // for name, child of node.namedchilds.items():
    for (var name of ghost.ls()) {
	child = ghost.cd[name]
        child.ghostparent = ghost
        // set child's parent to ghost's parent
        child.cd[".."] = ghost.cd[".."]

	/* when putting the child in the parent's namedchilds, we don't need to worry about the name already being taken, because we moved every child that could be touched here that was already there inside the ghostnode upon creating it. */
        // hang the child to ghost's parent
        parent = ghost.cd[".."]
	parent.cd[name] = child
        // delete child from ghost
        ghost.cd.delete(name)
    }
}

// for each generated file, map its line numbers to the original line numbers in ct file
var ctlinenr = {}

// the name of ghost nodes
var GHOST = "#" 

/* assemble assembles a codechunk recursively, filling up its leading
space to leadingspace. this way we can take chunks that are already
(or partly) indented with respect to their parent in the editor, and
chunks that are not.  */

function assemble(node, leadingspace, rootname, genlinenr) {

    // if it's a ghost node, remember the last named parent up the tree
    if (node.name == GHOST) {
        var lnp = lastnamed(node)
    }

    //var out = ""
    //var lines = node.text.split("\n")

    /* 
    find out a first line how much this chunk is alredy indented
    and determine how much needs to be filled up
    */
    // leading space already there
    if (node.lines.length > 0) {
	alreadyspace = node.lines[0].match(/^\s*/)[0]
    } else {
	alreadyspace = "" // no line, so no leading space already there
    }
    // space that needs to be added
    addspace = leadingspace.replace(alreadyspace, "") // replace only the first of alreadyspace

    // if the rootname isn't in ctlinenr yet, put it there
    if (!rootname in ctlinenr) {
        ctlinenr[rootname] = {}
    }

    var out = ""
    var ighost = 0 
    // for (var line of lines) {
    for (var i = 0; i < node.lines.length; i++) {
	var line = node.lines[i]
        if (isname(line)) {

            // remember leading whitespace
            childleadingspace = line.match(/^\s*/)[0] + addspace
            // print("#getname 1")
	    name = getname(line)
            if (name == ".") {  // assemble a ghost-child
		[outnew, genlinenr] = assemble(node.ghostchilds[ighost], childleadingspace, rootname, genlinenr)
		// append the text
		outnew += out
                ighost += 1
	    } else {            // assemble a name child
                if (node.name == GHOST) {
                    // if at ghost node, we get to the child via the last named parent
		    child = lnp.cd[name]
		} else {
		    child = node.cd[name]
		}
                [outnew, genlinenr] = assemble(child, childleadingspace, rootname, genlinenr)
		out += outnew
	    }
	}
	else {  // not name line, normal line
            out += addspace + line + "\n"
	    // map from the line number in the generated source to the original line number in the ct
	    ctlinenr[rootname][genlinenr] = node.ctlinenr[i]
	    // we added one line to root, so count up
	    genlinenr += 1
	}
    }
    
    return [out, genlinenr]
}







