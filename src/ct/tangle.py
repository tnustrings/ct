# tangle.py: codetext tangle with chunkname paths
# usage: tangle.main("file.ct")

# for codetext syntax, see readme.md


# imports

import re
import sys

""" code-chunks are represented as nodes in a tree. """
class Node: 
    def __init__(self, name, parent):
        self.name = name
        self.cd = {}
        self.cd["."] = self
        if parent == None:
            self.cd[".."] = self
        else:
            self.cd[".."] = parent

        # keep the text as lines, cause splitting on '\n' on empty text gives length 1 (length 0 is wanted)
        self.lines = []
        # the ghost children. why more than one?
        self.ghostchilds = []
        # for each text line, ctlinenr holds its original line nr in ct file
        self.ctlinenr = {}
        
    # ls lists the named childs
    def ls(self):
        # return all except . and ..
        return [k for k in self.cd.keys() if k != "." and k != ".."]

# debug offers a turn-offable print
def debug(s):
    i = 0
    # print(s)
    
# isdeclaration returns true if line is the declaration line of a code chunk
def isdeclaration(line):
    # return re.match(r".*<<.*>>=", line)
    starttick = bool(re.search(r"^``[^\`]*$", line))
    wordchar = bool(re.search(r"^``.*\w+.*$", line))
    ret = starttick and wordchar
    debug(f"isdeclaration {line}: {ret}")
    return starttick and wordchar

# isname returns true if line is the referencing name line of a code chunk
def isname(line):
    ret = bool(re.match(r".*``.*``", line))
    debug(f"isname {line}: {ret}")
    return ret

# isend says if the line is the end of a code chunk
def isend(line):
    # return re.match(r"^@$", line) # only allow single @ on line, to avoid mistaking @-code-annotations for doku-markers
    ret = bool(re.match(r"^``$", line))
    debug(f"isend {line}: {ret}")
    return ret

# isroot says whether the name is a root
def isroot(name):
    ret = re.match(r"^//", name)
    debug(f"isroot {name}: {ret}")
    return ret

# getname gets the chunkname from a declaration or in-chunk line
def getname(line):
    name = re.sub(r"^[^`]*``", "", line, 1) # 1: count
    name = re.sub(r"``.*", "", name) # chunk declarations do not have this
    name = re.sub("\n$", "", name)
    
    debug(f"getname({line}): '{name}'")

    return name

# for each generated file, map its line numbers to the original line
# numbers in ct file
ctlinenr = {}

# the name of ghost nodes
GHOST = "#" 

# assemble assembles a codechunk recursively, filling up its leading
# space to leadingspace. this way we can take chunks that are already
# (or partly) indented with respect to their parent in the editor, and
# chunks that are not. 
def assemble(node, leadingspace, rootname, genlinenr):

    global ctlinenr

    if node.name == GHOST:
        lnp = lastnamed(node)
        
    """ 
    find out a first line how much this chunk is alredy indented
    and determine how much needs to be filled up
    """
    # leading space already there (in first line)
    if len(node.lines) > 0:
        alreadyspace = re.search(r"^\s*", node.lines[0]).group()
    else:
        alreadyspace = "" # no line, so no leading space already there
    # space that needs to be added
    addspace = leadingspace.replace(alreadyspace, "", 1) # 1: replace once

    # if the rootname isn't in ctlinenr yet, put it there
    if not rootname in ctlinenr:
        ctlinenr[rootname] = {}

    out = ""
    ighost = 0 
    # for line in lines:
    for i, line in enumerate(node.lines):
        if isname(line):

            # remember leading whitespace
            childleadingspace = re.search(r"^\s*", line).group() + addspace
            # print("#getname 1")
            name = getname(line)
            if name == ".":   # assemble a ghost-child
                (outnew, genlinenr) = assemble(node.ghostchilds[ighost], childleadingspace, rootname, genlinenr) 
                out += outnew #  + "\n" # why add \n?
                ighost += 1
            else:             # assemble a name child
                if node.name == GHOST:
                    # if at ghost node, we get to the child via the last named parent
                    child = lnp.cd[name]
                else:
                    child = node.cd[name]
                (outnew, genlinenr) = assemble(child, childleadingspace, rootname, genlinenr) 
                out += outnew # + "\n" # why add \n?
        else: # normal line
            out += addspace + line + "\n"
            # map from the line number in the generated source to the original line number in the ct
            ctlinenr[rootname][genlinenr] = node.ctlinenr[i]
            genlinenr += 1 # we added one line to root, so count up

    # return the generated text and the new root line number (should be the same as the number of lines in out, so maybe don't return it?)
    return (out, genlinenr)

currentnode = None # the node we're currently at
openghost = None # if the last chunk opened a ghostnode, its this one

# put puts text in tree under relative or absolute path
def put(path, text, ctlinenr):
    global openghost
    global currentnode

    # print("put(" + path + ")")
    
    #if currentnode != None:
    #    print(f"current node in put: {pwd(currentnode)}")

    #if openghost != None:
    #    print(f"openghost: {pwd(openghost)}")

    # remove leading and trailing /

    # its ok to remove the leading / because roots have already been
    # identified and to start a relative path would need to be made
    # explicit with .
    # path = path.strip("/") 
        
    # if at file path, take only the path and chop off the alias
    debug(f"path: {path}")
    if re.match(r"^//", path): 
        parts = path.split(": ")
        path = parts[0]

    # create a ghostnode if called for
    if path == "." or path == "" and openghost != None:
        currentnode = openghost
        openghost = None # necessary?
    else:
        # if the path would need a node to cling to but there isn't noe
        if currentnode is None and not isroot(path):
            print(f"error: there's no file to attach '{path}' to, should it start with '//'?")
            exit
        # go to node, if necessary create it along the way
        currentnode = cdmk(currentnode, path)

    # append the text to node
    concatcreatechilds(currentnode, text, ctlinenr)


# cdmk walks the path from node and creates nodes if needed along the way.
# it returns the node it ended up at
def cdmk(node, path):
    # if path not relative, start from a root
    #if not (re.match(r"^\.", path) or path == ""): # why path == ""?
    #    # we may switch roots here. before that, we need to exit open ghosts. cdroot does this along the way
    #    cdroot(node)
    #    (node, path) = getroot(path)

    # if file path // switch roots
    if re.match(r"^//", path):
        # we need to exit open ghost nodes. cdroot does this along the way.
        cdroot(node)
        (node, path) = getroot(path)
    # if absolute path / go to root
    elif re.match(r"^/", path):
        # we need to exit open ghost nodes. cdroot does this along the way.
        node = cdroot(node)

    # remove leading / of absolute path
    path = path.strip("/")

    # follow the path
        
    elems = path.split("/")

    search = False # search for the next name
    for elem in elems:
        # do we start a sub-tree search?
        if elem == "*":
            search = True
            continue
        if search == True:
            # search for the current name
            search = False # reset
            res = []
            bfs(node, elem, res) # search elem in node's subtree
            if len(res) > 1:
                print(f"error: more than one nodes named {elem} in sub-tree of {pwd(node)}")
                exit
            elif len(les) == 0:
                print(f"error: no nodes named {elem} in sub-tree of {pwd(node)}")
                exit
            else:
                node = res[0]
            continue

        # standard:
        # walk one step
        walk = cdone(node, elem)
        # if child not there, create it
        if walk == None:
            walk = createadd(elem, node)
        node = walk

    # print("put return: " + str(node.name))
    return node # the node we ended up at

# bfs breath-first searches for all nodes named 'name' starting from 'node' and puts them in 'out'
def bfs(node, name, out):
    #    print(f"bfs {node}")
    if node.name == name:
        out.append(node)
    # search the node's childs
    for childname in node.ls():
        bfs(node.cd[childname], name, out)
    # do we need to search the gostchilds?
    for child in node.ghostchilds:
        bfs(child, name, out)

# cdroot cds back to root. side effect: ghosts are exited
def cdroot(node):
    if node == None: return None
    if node.cd[".."] == node: # we're at a root
        return node
    # continue via the parent
    return cdroot(cdone(node, ".."))
    
# cdone walks one step from node
def cdone(node, step):
    if step == GHOST:
        print("error: we may not walk into a ghost node via path")
        exit
    if step == "":
        step = "."
    if step == "..": # up the tree
        # debug(f"call exitghost for {pwd(node)}")
        exitghost(node)

    if step in node.cd:
        return node.cd[step]
    
    return None
    
# exitghost moves ghost node's named children to last named parent. needs to be called after leaving a ghost node
def exitghost(ghost):
    # not a ghost? do nothing
    if ghost == None or ghost.name != GHOST:
        return
    #debug("exitghost")

    """ if we exit a ghost node, we move all its named childs to the ghost node's parent and let the ghostnode be the childs' ghostparent (from where they can get e.g. their indent) """
    # for name, child in node.namedchilds.items():
    for name in ghost.ls():
        child = ghost.cd[name]
        child.ghostparent = ghost
        # set child's parent to ghost's parent
        child.cd[".."] = ghost.cd[".."]
        """ when putting the child in the parent's namedchilds, we don't need to worry about the name already being taken, because we moved every child that could be touched here that was already there inside the ghostnode upon creating it. """
        # hang the child to ghost's parent
        parent = ghost.cd[".."]
        parent.cd[name] = child
        # delete child from ghost
        del ghost.cd[name]

# pwd: print directory of node
def pwd(node):
    out = node.name
    walk = node
    while walk.cd[".."] != walk:
        walk = walk.cd[".."]
        out = walk.name + "/" + out
    return out

# createadd creates a named or ghost node and adds it to its parent
def createadd(name, parent):

    node = Node(name, parent)
    # debug(f"createadd: {pwd(node)}")
    
    # if we're creating a ghost node
    if node.name == GHOST:
        # debug(f"creating a ghost child for {parent.name}")
        # add it to its parent's ghost nodes
        parent.ghostchilds.append(node)
    else:
        # we're creating a name node
        
        """ if the parent is a ghost node, this node could have already been created before with its non-ghost path (an earlier chunk in the codetext might have declared it and put text into it, with children/ghost children, etc), then we move it as a named child from the last named parent to here """
        # if a node with this name is already child of last named parent, move it here
        if parent.name == GHOST:
            namedp = lastnamed(node)
            if node.name in namedp.ls():
                node = namedp.cd[name]
                del namedp.cd[name]
                node.cd[".."] = parent

        # add named node to parent, if it was created or moved
        parent.cd[name] = node

    return node

# concatcreatechilds concatenates text to node and creates children from text (named or ghost)
# this is the only place where text gets added to nodes
def concatcreatechilds(node, text, ctlinenr):
    global openghost
    openghost = None # why here? not so clear. but we need to reset it somewhere, that only the direct next code chunk can fill a ghost node

    # replace the last \n so that spli doesn't produce an empty line at the end
    text = re.sub(r"\n$", "", text)
    newlines = text.split("\n")

    # map from the line number in node to original line number in ct (get existing line count before new lines are added to node)
    N = len(node.lines)
    for i, _ in enumerate(newlines):
        # go through the new lines
        node.ctlinenr[N+i] = ctlinenr + i
    debug("N: " + str(N))
    debug("node.ctlinenr: " + str(node.ctlinenr))

    # put the new lines into node
    node.lines.extend(newlines)

    # generate the child nodes
    for line in newlines:
        if not isname(line):
            continue
        """ why do we create the children when concating text? maybe because here we know where childs of ghost nodes end up in the tree. """
        # debug("#getname 2")
        name = getname(line)
        if name == ".": # ghost child
            # if we're not at the first ghost chunk here
            if openghost != None:
                print("error: only one ghost child per text chunk allowed")
                exit
            # create the ghost chunk
            openghost = createadd(GHOST, node)
        else: # we're at a name
            if not name in node.ls():
                createadd(name, node)

# lastnamed returns the last named parent node
def lastnamed(node):
    if node == None: return None
    if node.name != GHOST: return node
    return lastnamed(node.cd[".."])

fileforalias = {} # from alias to filenames
roots = {}

# getroot returns root referenced by path and chops off root in path, except if there's only one un-aliased root.
def getroot(path):
    # remove the leading // of root path
    path = path.strip("/")
    
    # split the path
    p = path.split("/")

    # allow splitting file paths and subsequent chunk paths by //?
    # p = path.split("//")
    # return (resolveroot(p[0]), p[1])
    """
    # if only one root, it can be made implicit in the path names
    if len(roots) == 1:
        # return this one root along with the whole path
        only = roots[list(roots.keys())[0]]
        # when we're at the root chunk itsself, we want it out of the path in any case
        if len(p) == 1 and (p[0] in roots or p[0] in fileforalias):
            return (only, "")
        elif p[0] in fileforalias: # the first element is aliasing the one root, keep it
            return (only, "/".join(p[1:]))
        else:
            return (only, path)
    else: # the only root is omitted in path
        # return the root as referenced by the first path-part, and the rest of the path
        return (resolveroot(p[0]), "/".join(p[1:]))
    """
    # return the root as referenced by the first path-part, and the rest of the path
    return (resolveroot(p[0]), "/".join(p[1:]))

# resolveroot returns the rootnode for a filename or a alias
def resolveroot(name):
    if name in roots: # name is a filename
        return roots[name]
    elif name in fileforalias: # name is an alias
        return roots[fileforalias[name]]
    print(f"error: root name '{name}' not found")
    exit

# printtree: print node tree recursively
def printtree(node):
    print(f"printtree of {node.name}")
    print(f"ls: {node.ls()}")
    for name in node.ls():
        printtree(node.cd[name])
    for child in node.ghostchilds:
        printtree(child)
        

## main runs codetext for text
def main(f):
    global fileforalias
    
    lines = f.readlines() # readlines keeps the \n for each line, text concat in nodes relies on that

    """
    maybe we need one pass of lines to get all root nodes, so that we know beforehand if there's one or more files. otherwise we keep the first path element of chunk-paths until we arrive at the second file and drop it afterwards, that wouldn't be so good.
    """
    for line in lines:
        # only look at declaration lines
        if not isdeclaration(line):
            continue
        # debug("#getname 3")
        name = getname(line)

        #if re.match(r"\w+\.\w+", name): # todo: path ~ ": " or only one path part
        
        # we're at a root if the name starts with a //
        if isroot(name):
            # maybe an alias follows the file name, seperated by ': '

            parts = name.split(": ")

            # remove leading slashes of root name
            filename = re.sub("^//", "", parts[0])
            # take the first part of path as filename/alias. todo: split at //?
            filename = filename.split("/")[0]
            
            # the root is already created? continue.
            
            # todo: if a root is created at its first reference, who
            # says that the first reference to a root contains its
            # alias?  should we assume this?  or should we first sift
            # through all root references and create the root if we
            # encounter it with its alias?
            if filename in roots:
                continue
            # is it an alias? continue. (this assumes aliases need to appear first in the text before they can be referenced.
            if filename in fileforalias:
                continue
            # create root for this file
            roots[filename] = Node(filename, None)
            if len(parts) > 1:
                alias = parts[1]
                # alias already used for other file
                if alias in fileforalias and fileforalias[alias] != filename:
                    print(f"error: alias {alias} already references file {fileforalias[alias]}")
                    exit
                # alias it
                fileforalias[alias] = filename

    """ no files, exit """
    if len(roots) == 0:
        exit

    # debug("roots: " + str(roots))

    """ now that we got the roots, we can start putting in the chunks. """

    # are we in a chunk
    inchunk = False
    # current chunk content
    chunk = ""
    # current chunk name/path
    path = None
    # start line of chunk in ct file
    chunkstart = 0
    
    # for line in lines:
    for i, line in enumerate(lines):
        #if isdeclaration(line): # at the beginning of chunk remember its name
        """we can't decide for sure whether we're opening or closing a chunk by looking at the backticks alone, cause an unnamed chunk is opend the same way it is closed.  so in addition, check that inchunk is false."""
        if bool(re.search(r"^``[^`]*", line)) and inchunk is False: # at the beginning of chunk remember its name
            inchunk = True
            # debug("#getname 4")
            path = getname(line)
            # remember the start line of chunk in ct file
            # add two: one, for line numbers start with one not zero, another, for the chunk text starts in the next line, not this
            chunkstart = i+2
        elif isend(line): # at the end of chunk save chunk
            inchunk = False
            # debug(f"calling put for: {path}")
            # debug("split chunk: " + str(chunk.split("\n")))
            put(path, chunk, chunkstart)
            # reset variables
            chunk = ""
            path = None
            #print("")
        elif inchunk: # when we're in chunk remember line
            chunk += line
        #else:
        #    debug(line, end="") # for debugging


    """ in the end we need to exit all un-exited ghost nodes so that their
    named children end up as the named children of the last named parent
    where we can access them.  """

    cdroot(currentnode)

    """ at the end, write all files (each file is a root) """
    for filename in roots:
        # todo: add don't edit comment like before
        
        # assemble the code
        (out, _) = assemble(roots[filename], "", filename, 1)
        # printtree(roots[filename])
        
        # and write it to file
        # print(f"write {filename}")
        f = open(filename, "w")
        f.write(out)
