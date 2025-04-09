# tex generates latex doc from ct

from ct import ct
import re
import subprocess
from pathlib import Path

# printtex prints latex doc
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

    # are we outside verbatim?
    outsideverbatim = True

    for i, line in enumerate(ct.ctlines):
        # are we at a chunk opening line?
        if i in ct.ischunkopening and ct.ischunkopening[i] == True:

            # print the collected doc text
            doctext(betweenchunk, mdtotex, shell)
            # reset the between chunk text
            betweenchunk = ""
            
            # save inchunk for setting child labels
            inchunk = True

            # this helps us keep track of child ref formatting in chunks
            outsideverbatim = False

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

            # insert vertical space before code chunk, since we set the verbatim margins to zero
            print("\n\\vspace{5mm}\n") # maybe use \addvspace\medskipamount


            # put the caption in the margin
            # this caption is invisible, it's for ref-counting
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
            print("\\begin{lstlisting}")

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
            if not outsideverbatim:
                print("\\end{lstlisting}")

                # if we come to a child ref line, we set outsideverbatim
                # to true even though we're in a chunk, to avoid empty
                # \begin{verbatim}\end{verbatim} blocks between child refs
                # that are in neighboring lines

                outsideverbatim = True
            else:
                # if we haven't just left the lstlisting, seperate the
                # child refs by a line break
                print("\\\\")

            # make a phantom section, so that links pointing to this label directly jump to this line and not to the start of the latex section this label is in
            print("\\phantomsection")


            
            # make an outgoing link to the first chunk of child

            # put leading spaces before the link, they get lost in
            # link text
            leadingspace = re.search("^\s*", line).group()

            # lstinline apparently doesn't print spaces. so use texttt and replace the spaces to ~ to get them printed. make sure that the style here (\small\texttt) is the same as passed to \lstset in the preamble in the basicstyle argument
            print("\\small{\\texttt{" + re.sub(" ", "~", leadingspace) + "}}", end="") # let it be on the same line as the link
            # delete the spaces we just printed
            line = re.sub(r"^\s*", "", line)

            # make the text for the link
            # first the line of the code-chunk
            text = "\\lstinline{" + re.sub(r"\n", "", line) + "}" # was \texttt
            # an extra space
            text += "\\ \\ "
            # then the tex reference to the child
            text += "\\textsf{\\scriptsize{\\color{gray}{" + ref(childpath + ":0") + "}}}"
            
            # make the link
            print(hyperref(childpath + ":0", text))


            
            # make label for incoming links # todo let a parent hyperref in marginnote point to this
            
            print(label(nodepath + " => " + childpath))

            # don't begin a verbatim again, cause the next line might also be a child ref, and a \end{verbatim}\begin{verbatim} would put in an empty line
            # print("\\begin{lstlisting}")

            continue
        elif i in ct.ischunkclose and ct.ischunkclose[i]:
            # we're at a chunk-closing line
            inchunk = False

            # close the verbatim if needed
            # if the last chunk-line(s) were a child-reference, we're not in verbatim anymore, check for that.
            if not outsideverbatim:
                print("\\end{lstlisting}")

            # now we're outside verbatim in any case
            outsideverbatim = True
            
            # print(line, end="")
                

            # insert vertical space before code chunk, since we set the verbatim margins to zero
            print("\n\\vspace{5mm}\n")


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
                # if we were put outside verbatim by a child reference, put us inside verbatim again
                if outsideverbatim:
                    print("\\begin{lstlisting}")
                    outsideverbatim = False
                    
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

# getheader returns a latex header
def getheader(lowercase:bool=False):
    h = headertex

    # redefine strings as lowercase
    if lowercase:
        h += """
% lowercasings

% display the contents name in lower case
\renewcommand{\contentsname}{contents}

% lowercase figure name
\captionsetup{figurename=figure}
        """

    return h

# gettemplate returns a latex template referencing the header
def gettemplate(header:str=None):

    out = tmpltex
    if header:
        # get the name of the header file without extension and path
        hstem = Path(header).stem
        out = re.sub("<header>", hstem, out)
        
    return out

## latex code


# r before string preserves backslash
tmpltex = r"""
\documentclass[a4paper]{report}
\input{<header>}
\begin{document}

\title{some programs}
\date{}

\maketitle

{\sffamily
\tableofcontents
}

\chapter{hello}

% include your program(s)
\input{}

\end{document}
"""

headertex = r"""
\usepackage{graphics,color,eurosym,latexsym}
%\usepackage{algorithm,algorithmic}
\usepackage{times, verbatim}
\usepackage[utf8]{inputenc}
\usepackage[OT1]{fontenc}
\usepackage{pst-all}

\usepackage{psfrag}
%\usepackage{inconsolata} % todo install
%\usepackage[straightquotes]{newtxtt} % todo install
%\bibliographystyle{plain}

% clickable links
\usepackage[hidelinks]{hyperref}

% remove vertical space before and after verbatim, see https://tex.stackexchange.com/a/43336
\usepackage{etoolbox}
\makeatletter
\preto{\@verbatim}{\topsep=0pt \partopsep=0pt }
\makeatother

% use floatrow to position caption to the left of figure / code?
% error: do not use float package with floatrow
% see https://tex.stackexchange.com/a/29144
% for subfloats see https://tex.stackexchange.com/questions/443732/subfloat-how-to-position-label-to-the-left
%\usepackage{floatrow}
% try positioning to the left with hbox: https://tex.stackexchange.com/a/346452

% create a custom float for code chunks, that behaves like figure or table.
% see https://www.andy-roberts.net/latex/floats_figures_captions/
\usepackage{float}

% latex seems to be able to cope with 18 unprocessed floats, does this here work?
% see https://tex.stackexchange.com/questions/46512/too-many-unprocessed-floats
\usepackage{morefloats}

% latex normally only allows three floats per page, we'd like to put in more codechunks, so change this
% for explanation see https://tex.stackexchange.com/a/39020
% for setcounter see https://tex.stackexchange.com/a/161760
\setcounter{topnumber}{20}
\setcounter{bottomnumber}{20}
\setcounter{totalnumber}{20}

% floatstyle needs to come before \newfloat
\floatstyle{plain}

% make a new float: pass type, placement and extension
% the type 'chunk' is the name of the float; the placement h means 'here' (position the float at the same point it occurs in the source text), and ext 'chunk' is the extension of auxiliary file for list of codes
\newfloat{code}{h!}{code}[chapter]
% [chapter] is the argument for the outer counter. it sets the refs to
% the codes to be by chapter, like 1-1, 1-2 etc. it doesn't seem to
% have an effect on the numbers appearing before code captions made
% with \captionof, use \counterwithin for that.  see
% https://en.wikibooks.org/wiki/LaTeX/Floats,_Figures_and_Captions


% specify how the caption for float should be prefixed, like 'chunk 1.3'
%\floatname{code}{code}
\floatname{code}{}

% count the code chunks by chapter, code 1.1, 1.2, etc
% see https://tex.stackexchange.com/a/28334
% counterwithin doesn't seem to have an effect on the ref numbering for code though, use the [chapter] arg in \newfloat for this.
\counterwithin{code}{chapter}

% make code numberings like 1-1, 1-2 instead of 1.1, 1.2
% see https://tex.stackexchange.com/a/616698
\renewcommand{\thecode}{\thechapter-\arabic{code}}

% number figures with dashes, like 1-1, 1-2
\renewcommand{\thefigure}{\thechapter-\arabic{figure}}

% number sections like 1-1, 1-2 instead of 1.1, 1.2
\renewcommand{\thesection}{\thechapter-\arabic{section}}
% no section number
%\renewcommand{\thesection}{}

% just display numbers for chapters, without preceding 'Chapter'
\renewcommand{\chaptername}{}

%\renewcommand{\thechapter}{\textsf{\arabic{chapter}}}



% let captions float left (raggedright)
% see https://tex.stackexchange.com/a/275141
\usepackage[]{caption}
\captionsetup{justification=raggedright,singlelinecheck=false}

% make code captions invisible
% original plain format: #1: 'Label' + ' ' + '1.1', #2: sep, #3 text
% \DeclareCaptionFormat{plain}{#1#2#3\par}
% the problem seems to be the space between the 'Label' and the
% '1.1', it isn't removed when the caption name is empty, so the 1.1
% becomes indented by that space.
% so we hide the captions and put in an own \marginpar
% we need the captions for reference-counting
% we apparently can't pass no argument, then the references are not set, so we pass #3, the text, and leave it empty
\DeclareCaptionFormat{invisible}{#3}
% use the new format for code-float captions
\captionsetup[code]{format=invisible}


% for margin notes in floats, \marginpar doesn't work there
% see https://tex.stackexchange.com/a/211803
\usepackage{marginnote}

% for coloring marginnotes
\usepackage{xcolor}

% use sectsty display title headings etc in sans-serif font
\usepackage{sectsty}
\allsectionsfont{\sffamily}



% use titling to left-align the title
% see https://tex.stackexchange.com/a/311505
\usepackage{titling}

% make titling elements \sffamily and align-left (raggedright)

\pretitle{\begin{raggedright}\sffamily\LARGE}
\posttitle{\end{raggedright}}

\preauthor{\begin{raggedright}
            \large\sffamily \lineskip 0.5em%
            \begin{tabular}[t]{c}}
\postauthor{\end{tabular}\end{raggedright}}

\predate{\begin{raggedright}\sffamily\large}
\postdate{\end{raggedright}}


% commands for unnumbered section, subsection, subsubsection
% see https://stackoverflow.com/a/77642053
\newcommand{\usection}[1]{\section*{#1}
\addcontentsline{toc}{section}{\protect\numberline{}#1}}

\newcommand{\usubsection}[1]{\subsection*{#1}
\addcontentsline{toc}{subsection}{\protect\numberline{}#1}}

\newcommand{\usubsubsection}[1]{\subsubsection*{#1}
\addcontentsline{toc}{subsubsection}{\protect\numberline{}#1}}


% set paragraph indent to 0
\setlength\parindent{0pt}


% verbatim doesn't seem to wrap lines, so try lstlisting
% see https://tex.stackexchange.com/a/121618
\usepackage{listings}
\lstset{
basicstyle=\small\ttfamily,
columns=flexible,
breaklines=true
}
% remove the vertical space above and below listings, cause listings can stop / resume inside codechunks at child references
% see https://tex.stackexchange.com/a/68903
\lstset{aboveskip=0pt,belowskip=0pt}"""

