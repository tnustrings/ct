# totex generates latex doc from ct

from ct import ct
import re
import subprocess

# print tex prints latex doc
def printtex(text, mdtotex:str=None, shell:bool=False):
    
    # run codetext
    ct.ct(text)

    """ referencing specific code-chunks appended to a node: we'd like
    to be able to jump backward and forward to where text was appended
    previously to a node (if it was) and to where it is appended next
    (if it is). since the text of all the chunks appending to a node
    ends up in the same node, maybe a way to do this is to keep a
    counter for each node at which chunk number (0-indexed) we
    are. this counter can be put with # behind a nodepath for
    referencing."""
    
    # are we in a chunk?
    inchunk = False

    # the current chunk number for each node, 0-indexed
    currentchunknum = {} # str to int
    # a shorthand for the current chunk num
    chunknum = None

    # the current path
    path = None

    # collected text between chunks
    betweenchunk = ""

    for i, line in enumerate(ct.ctlines):
        # are we at a chunk opening line?
        if i in ct.ischunkopening and ct.ischunkopening[i] == True:

            # print the collected doc text
            doctext(betweenchunk, mdtotex, shell)
            betweenchunk = ""
            
            # save inchunk for setting child labels
            inchunk = True

            # get the node
            node = ct.nodeatctline[i] # todo change this to include the opening line

            path = ct.pwd(node)

            # if we're at the first chunk that's appended to this node, set currentchunknum
            if not path in currentchunknum:
                currentchunknum[path] = 0

            chunknum = currentchunknum[path]

            parent = node.cd[".."]

            # get the labels for referencing
            parentlabel = None
            backlabel = None
            fwdlabel = None
            
            # parent label
            if not node.isroot():
                # this node is referenced from the ith chunk in its parent
                parentlabel = ct.pwd(parent) + ":" + str(node.iparentchunk)

            # ref to the previous chunk appended to this node
            if not chunknum == 0:
                backlabel = path + ":" + str(chunknum-1)

            # ref to the next chunk appended to this node
            if chunknum < node.nchunks - 1:
                fwdlabel = path + ":" + str(chunknum+1)

            # label for this chunk and it's number
            thislabel = path + ":" + str(chunknum)

            # now all the refs should be there

            # print the path like given in the ct (maybe option to
            # always print the whole path?)
            printpath = ct.getname(line)

            # code is a custom float defined in header.tex, it is
            # referenceable like \figure
            # print("\\begin{code}")

            # increase the counter
            # print("\\captionlistentry{}") # doesn't work outside of float, exept if there is one \labelof before, why?

            # put the caption in the margin
            print("\marginpar{\captionof{code}{" + label(thislabel) + "}}")
            # set the label
            # print(label(thislabel))

            
            #print("\\label{code:b}


            # put the references on the margin. \marginpar doesn't
            # work in floats
            # try not to include blanks between the commands, they seem to have an effect on the layout
            # \textsf: sans-serif font
            marginnote = "\\marginnote{\\textsf{\\scriptsize{\\color{gray}\\textbf{" + ref(thislabel) + "}"
            if parentlabel:
                marginnote += " p" + ref(parentlabel) + "\\\\"
            if backlabel:
                marginnote += " b" + ref(backlabel) 
            if fwdlabel:
                marginnote += " f" + ref(fwdlabel)

            # close color, scriptsize and marginnote
            marginnote += "}}}"
            print(marginnote)

            # start the chunk
            print("\\begin{verbatim}")

            print(printpath)
            
            continue
        elif inchunk == True and ct.isname(line):

            # if we're at a reference to a child
            node = ct.nodeatctline[i]
            nodepath = ct.pwd(node)
            # get the child
            child = node.childatctline[i]
            childpath = ct.pwd(child)

            # set the link of a child

            # leave verbatim for the ref and label
            print("\\end{verbatim}")

            # make a phantom section, so that links pointing to this label directly jump to this line and not to the start of the latex section this label is in
            print("\\phantomsection")

            # make outgoing link to first chunk of child
            print(hyperref(childpath + ":0", "\\texttt{" + re.sub(r"\n", "", line) + "} [" + ref(childpath + ":0") + "]"))

            # make label for incoming links # todo let a parent hyperref in marginnote point to this
            print(label(nodepath + " => " + childpath))

            # begin verbatim again
            print("\\begin{verbatim}")

            continue
        elif i in ct.ischunkclose and ct.ischunkclose[i]:
            # we're at a closing line
            inchunk = False
            
            # print(line, end="")
                
            # close the latex
            print("\\end{verbatim}")

            # captionlistentry increases the reference counter without
            # using a \caption
            # see https://tex.stackexchange.com/a/438500
            # print("\\captionlistentry{}")

            # the label for the code-chunk
            # print(label(thislabel))

            # end the code-chunk
            # print("\\end{code}")

            # and count the chunk number up for this node
            currentchunknum[path] = currentchunknum[path] + 1
        else:
            if inchunk:
                # in-chunk line that's not a child reference, just print it
                print(line, end="")
            else:
                # doc-line, collect it, for maybe converting it to tex if asked to
                betweenchunk += line + "\n"

    # print the last doctext 
    doctext(betweenchunk, mdtotex, shell)

# doctext prints doc text, converting it to tex if asked to
def doctext(text, mdtotex, shell:bool=False):
    # is a command to convert from md to tex?
    if mdtotex and len(mdtotex) == 1:
        # do the converting
        cmd = mdtotex[0]
        
        # pass the command split at blanks
        arr = re.split(r"\s+", cmd)
        # print("arr: " + str(arr))
        p = subprocess.run(arr, input=text, capture_output=True, text=True, shell=shell)

        # get the returned text
        text = p.stdout

    print(text)



# all helper functions replace slashes in labels with colons

# hyperref gives a hyperref to label            
def hyperref(tolabel: str, text: str) -> str:
    # try replacing slashes
    tolabel = re.sub(r"/", ":", tolabel)

    return "\\hyperref[" + tolabel + "]{" + text + "}"

# ref gives a ref to label            
def ref(label: str) -> str:
    # try replacing slashes
    label = re.sub(r"/", ":", label)

    return "\\ref{" + label + "}"

# label gives a label
def label(label: str) -> str:
    # try replacing slashes
    label = re.sub(r"/", ":", label)
    return "\\label{" + label + "}"

# pageref gives a pageref
def pageref(label: str) -> str:
    # try replacing slashes
    label = re.sub(r"/", ":", label)
    return "\\pageref{" + label + "}"
