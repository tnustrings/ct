// tex.go was generated from tex.ct. 
// if you edit tex.go, run `ct -tc tex.go` afterwards to update tex.ct.
// when editing this file leave the `` tags.
//``ct version 0.1.1
//``ct files tex.ct tex.go

// # tex
// 
// tex generates latex doc from ct.  1``//tex.go:
package ct
import (
    "io"
    "os/exec"
    "fmt"
    "regexp"
    "strings"
    "github.com/tnustrings/ct/internal/fc"
)
// totex gives latex doc for ct text.  2``/totex:
func Totex(text string, ctfile string, mdtotex string) string {
    // run codetext.  3.``
    Ct(text, ctfile)
    // `out` collects our returned text.  4``
    out := ""
    // put in a 'don't edit' message.  5``
    if ctfile != "" {
        out += fmt.Sprintf("%% this file was generated from %s. please edit %v.\n", ctfile, ctfile)
    } else {
        out += fmt.Sprintf("%% this file is generated. please edit the .ct file from which it stems.\n")
    }
    // referencing specific code-chunks appended to a node: we'd like to be
    // able to jump backward and forward to where text was appended
    // previously to a node (if it was) and to where it is appended next (if
    // it is). since the text of all the chunks appending to a node ends up
    // in the same node, maybe a way to do this is to keep a counter for each
    // node at which chunk number (0-indexed) we are. this counter can be put
    // with # behind a nodepath for referencing.
    // 
    // set the housekeeping variables for the loop.  6``
    // are we in a chunk?
    inchunk := false

    // the current chunk number for each node, 0-indexed
    chunu := make(map[string]int) // was currentchunknum
    // a shorthand for the current chunk num
    //chunknum := -1

    // the current path
    path := ""

    // collected text between chunks
    betweentext := ""

    // are we outside verbatim?
    outsideverbatim := true
    // we go over the ctlines.  7``
    for i, line := range ctlines {
        if Chop(i) {
            // are we at a chunk opening line? print the collected between-chunk text
            // and reset it.  8``./opening:
            out += convert(betweentext, mdtotex)
            betweentext = ""
            // remember that we're in a chunk for setting child labels later.  9``
            inchunk = true
            // outsideverbatim helps us keep track of child ref formatting in chunks.  10``
            outsideverbatim = false
            // get the node and the path.  11``
            node := nodeatict[i] // todo change this to include the opening line
            path = pwd(node)
            // if we're at the first chunk that's appended to this node, init
            // chunu.  12``
            if _, ok := chunu[path]; !ok {
                chunu[path] = 0
            }
            //chunknum = chunu[path]
            // stick together the latex labels for referencing the parent, the
            // forward and backward chunk.  13``
            parent := node.parent

            parentlabel := ""
            backlabel := ""
            fwdlabel := ""
            // the parent label says that this node is referenced from the ith chunk
            // in its parent.  14``
            if !node.isroot() {
                parentlabel = pwd(parent) + ":" + itoa(node.chup)
            }
            // the backward chunk is the previous chunk appended to this node.  15``
            if !(chunu[path] == 0) {
                backlabel = path + ":" + itoa(chunu[path]-1)
            }
            // the forward chunk is the next chunk appended to this node.  16``
            if chunu[path] < len(node.chunks) - 1 {
                fwdlabel = path + ":" + itoa(chunu[path]+1)
	    }
            // label for this chunk and its number.  17``
            thislabel := path + ":" + itoa(chunu[path])
            // now all the refs should be there.
            // 
            // print the path like given in the ct (maybe option to always print the
            // whole path?)  18``
            printpath := getname(line)
            // insert vertical space before code chunk, since we set the verbatim
            // margins to zero.  19``
            out += "\n\\vspace{5mm}\n\n" // maybe use \addvspace\medskipamount
            // put the caption in the margin. this caption is invisible, it's for
            // ref-counting.  20``
            out += "\\marginpar{\\captionof{code}{" + label(thislabel) + "}}\n"
            // put the references on the margin. \marginpar doesn't work in
            // floats. try not to include blanks between the commands, they seem to
            // have an effect on the layout.  21``
            // \textsf: sans-serif font
            marginnote := "\\marginnote{\\textsf{\\scriptsize{\\color{gray}\\textbf{" + ref(thislabel) + "}"
            if parentlabel != "" {
                marginnote += " p" + ref(parentlabel) + "\\\\"
            }
	    if backlabel != "" {
                marginnote += " b" + ref(backlabel) 
            }
	    if fwdlabel != "" {
                marginnote += " f" + ref(fwdlabel)
            } // close color, scriptsize and marginnote
            marginnote += "}}}"
            out += marginnote + "\n"

            // print the chunk path.  22``
            //make the chunk path sans-serif
            //printpath = re.sub("/", "\\/", printpath)
            printpath = strings.ReplaceAll(printpath, "_", "\\_")
            out += "\n"
            out += "\\textsf{\\footnotesize{\\color{gray}{" + printpath + "}}}\n"
            out += "\n"
            // start the chunk and skip the rest of the loop.  23``
            out += "\\begin{lstlisting}\n"
	    continue
            //``8
	} else if inchunk == true && isreference(line) {
	    // if we're at a reference to a child.
	    // 
	    // get the node and the child.  24``../child ref:
            node := nodeatict[i]
            nodepath := pwd(node)
            child := node.caict[i]
            childpath := pwd(child)
	    // if we come to a child ref line, we set outsideverbatim to true even
	    // though we're in a chunk, to avoid empty \begin{verbatim}\end{verbatim}
	    // blocks between child refs that are in neighboring lines.  25``
            if ! outsideverbatim {
                out += "\\end{lstlisting}"
                outsideverbatim = true
            } else {
                // if we haven't just left the lstlisting, seperate the
                // child refs by a line break
                out += "\\\\" + "\n"
            }
	    // make a phantom section, so that links pointing to this label directly
	    // jump to this line and not to the start of the latex section this label
	    // is in.  26``
            out += "\\phantomsection" + "\n"
	    // make an outgoing link to the first chunk of child.
	    // 
	    // put leading spaces before the link, they get lost in link text.  27``
	    r := regexp.MustCompile(`^\s*`)
	    f := r.FindStringIndex(line)
            leadingspace := line[f[0]:f[1]]
	    // lstinline apparently doesn't print spaces. so use texttt and replace
	    // the spaces to ~ to get them printed. make sure that the style here
	    // (\small\texttt) is the same as passed to \lstset in the preamble in
	    // the basicstyle argument.
	    // 
	    // don't break the line, let it be on the same line as the following
	    // link.  28``
	    leadingspace = strings.ReplaceAll(leadingspace, " ", "~")
            out += "\\small{\\texttt{" + leadingspace + "}}"
	    // delete the spaces we just printed.  29``
            line = r.ReplaceAllString(line, "")
	    // make the text for the link and the link.  30``
	    // first the line of the code-chunk.
            text = "\\lstinline{" + strings.ReplaceAll(line, "\n", "") + "}" // was \texttt
            // an extra space
            text += "\\ \\ "
            // then the tex reference to the child
            text += "\\textsf{\\scriptsize{\\color{gray}{" + ref(childpath + ":0") + "}}}"
            
            // make the link
            out += hyperref(childpath + ":0", text) + "\n"
	    // make the label for incoming links # todo let a parent hyperref in
	    // marginnote point to this.  31``
            out += label(nodepath + " => " + childpath) + "\n"
	    // don't begin a verbatim again, cause the next line might also be a
	    // child ref, and a \end{verbatim}\begin{verbatim} would put in an empty
	    // line.  32``
            // print("\\begin{lstlisting}")
	    // skip to the next line.  33``
            continue
	    //``24
	} else if Chclo(i) {
	    // we're at a chunk-closing line.  34``../close:
            inchunk = false
	    // close the verbatim if needed - if the last chunk-line(s) were child-references we're already outside verbatim.  35``
            if !outsideverbatim {
                out += "\\end{lstlisting}" + "\n"
            }
            // now we're outside verbatim in any case
            outsideverbatim = true
	    // insert vertical space before code chunk, since we set the verbatim
	    // margins to zero.  36``
            out += "\n\\vspace{5mm}\n" + "\n"
	    // count the chunk number up for this node.  37``
            chunu[path] = chunu[path] + 1
	    //``34
	} else {
            if inchunk {
	        // if we're in a chunk line, if we were put outside verbatim by a child
	        // reference, put us inside verbatim again.  38``../chunk line:
                if outsideverbatim {
                    out += "\\begin{lstlisting}" + "\n"
                    outsideverbatim = false
		}
	        // in-chunk line that's not a child reference, just print it. the ct lines come with linebreak, don't add it.  39``
                out += line
	        //``38
	    } else {
	        // collect between-chunk lines, for maybe converting them to tex if asked
	        // to.  40``../between line:
                betweentext += line 
	        //``40
	    }
	}
    }
    // after the loop, append the last between text and return.  41``..
    out += convert(betweentext, mdtotex)

    return out
    //``3
}
//``2
// doctext converts the text between chunks from md to tex if asked to.
// 
// todo fill it in.  42``/convert:
func convert(text string, mdtotex string) string {
    // chunk the command in words and run it.  43.``
    a := strings.Split(mdtotex, " ")
    cmd := exec.Command(a[0], a[1:]...)
    // send the text to stdin.  44``
    stdin, _ := cmd.StdinPipe()
    go func() {
        defer stdin.Close()
	io.WriteString(stdin, text)
    }()
    // get what was written to stdout.  45``
    out, _ := cmd.CombinedOutput()
    return string(out)
    //``43
}
//``42
// hyperref gives a hyperref to label.  46``/hyperref:
func hyperref(tolabel string, text string) string {
    // try replacing slashes
    r := regexp.MustCompile(`/`)
    tolabel = r.ReplaceAllString(tolabel, ":")

    return "\\hyperref[" + tolabel + "]{" + text + "}"
}
//``46
// ref gives a ref to label.  47``/ref:
func ref(label string) string {
    // try replacing slashes
    r := regexp.MustCompile(`/`)
    label = r.ReplaceAllString(label, ":")

    return "\\ref{" + label + "}"
}
//``47
// label gives a label.  48``/label:
func label(label string) string {
    // try replacing slashes
    r := regexp.MustCompile("/")
    label = r.ReplaceAllString(label, ":")
    return "\\label{" + label + "}"
}
//``48
// pageref gives a pageref.  49``/pageref:
func pageref(label string) string {
    // try replacing slashes
    r := regexp.MustCompile("/")
    label = r.ReplaceAllString(label, ":")
    return "\\pageref{" + label + "}"
}
//``49
// TexHeader returns a latex header  50``/texheader:
func TexHeader(lowercase bool) string {
    h := headertex

    // redefine strings as lowercase
    if lowercase {
        h += `
% lowercasings

% display the contents name in lower case
\renewcommand{\contentsname}{contents}

% lowercase figure name
\captionsetup{figurename=figure}
        `
    }
    return h
}
//``50
// gettemplate returns a latex template referencing the header  51``/textemplate:
func TexTemplate(header string) string {
    out := tmpltex
    if header != "" {
        // get the name of the header file without extension and path
        hstem := fc.Stem(header)
        out = strings.Replace(out, "<header>", hstem, 1)
    }
    return out
}
//``51
// the tex for the template.  52``/texstring:
var tmpltex = `
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
`
// the tex for the header.  53``/texstring
var headertex = `
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
%\usepackage{etoolbox}
%\makeatletter
%\preto{\@verbatim}{\topsep=0pt \partopsep=0pt }
%\makeatother

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
% \renewcommand{\thesection}{\thechapter-\arabic{section}}

% no section number
\renewcommand{\thesection}{}
\renewcommand{\thesubsection}{}
\renewcommand{\thesubsubsection}{}


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
\lstset{aboveskip=0pt,belowskip=0pt}
`
//``52
//``1
