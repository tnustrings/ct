package ct
import (
    "io"
    "os/exec"
    "fmt"
    "regexp"
    "strings"
    "tnustrings/fc"
)
func Totex(text string, ctfile string, mdtotex string) string {
    Ct(text)
    out := ""
    if ctfile != "" {
        out += fmt.Sprintf("% this file was generated from %s. please edit %s.\n", ctfile, ctfile)
    } else {
        out += fmt.Sprintf("% this file is generated. please edit the .ct file from which it stems.\n")
    }
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
    for i, line := range ctlines {
        if Chop(i) {
            out += convert(betweentext, mdtotex)
            betweentext = ""
            inchunk = true
            outsideverbatim = false
            node := nat[i] // todo change this to include the opening line
            path = pwd(node)
            if _, ok := chunu[path]; !ok {
                chunu[path] = 0
            }
            //chunknum = chunu[path]
            parent := node.parent

            parentlabel := ""
            backlabel := ""
            fwdlabel := ""
            if !node.isroot() {
                parentlabel = pwd(parent) + ":" + itoa(node.chup)
            }
            if !(chunu[path] == 0) {
                backlabel = path + ":" + itoa(chunu[path]-1)
            }
            if chunu[path] < node.nchunks - 1 {
                fwdlabel = path + ":" + itoa(chunu[path]+1)
	    }
            thislabel := path + ":" + itoa(chunu[path])
            printpath := getname(line)
            out += "\n\\vspace{5mm}\n\n" // maybe use \addvspace\medskipamount
            out += "\\marginpar{\\captionof{code}{" + label(thislabel) + "}}\n"
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

            //make the chunk path sans-serif
            //printpath = re.sub("/", "\\/", printpath)
            printpath = strings.ReplaceAll(printpath, "_", "\\_")
            out += "\n"
            out += "\\textsf{\\footnotesize{\\color{gray}{" + printpath + "}}}\n"
            out += "\n"
            out += "\\begin{lstlisting}\n"
	    continue
	} else if inchunk == true && isname(line) {
            node := nat[i]
            nodepath := pwd(node)
            child := node.chat[i]
            childpath := pwd(child)
            if ! outsideverbatim {
                out += "\\end{lstlisting}"
                outsideverbatim = true
            } else {
                // if we haven't just left the lstlisting, seperate the
                // child refs by a line break
                out += "\\\\" + "\n"
            }
            out += "\\phantomsection" + "\n"
	    r := regexp.MustCompile(`^\s*`)
	    f := r.FindStringIndex(line)
            leadingspace := line[f[0]:f[1]]
	    leadingspace = strings.ReplaceAll(leadingspace, " ", "~")
            out += "\\small{\\texttt{" + leadingspace + "}}"
            line = r.ReplaceAllString(line, "")
	    // first the line of the code-chunk.
            text = "\\lstinline{" + strings.ReplaceAll(line, "\n", "") + "}" // was \texttt
            // an extra space
            text += "\\ \\ "
            // then the tex reference to the child
            text += "\\textsf{\\scriptsize{\\color{gray}{" + ref(childpath + ":0") + "}}}"
            
            // make the link
            out += hyperref(childpath + ":0", text) + "\n"
            out += label(nodepath + " => " + childpath) + "\n"
            // print("\\begin{lstlisting}")
            continue
	} else if Chclo(i) {
            inchunk = false
            if !outsideverbatim {
                out += "\\end{lstlisting}" + "\n"
            }
            // now we're outside verbatim in any case
            outsideverbatim = true
            out += "\n\\vspace{5mm}\n" + "\n"
            chunu[path] = chunu[path] + 1
	} else {
            if inchunk {
                if outsideverbatim {
                    out += "\\begin{lstlisting}" + "\n"
                    outsideverbatim = false
		}
                out += line
	    } else {
                betweentext += line 
	    }
	}
    }
    out += convert(betweentext, mdtotex)

    return out
}
func convert(text string, mdtotex string) string {
    a := strings.Split(mdtotex, " ")
    cmd := exec.Command(a[0], a[1:]...)
    stdin, _ := cmd.StdinPipe()
    go func() {
        defer stdin.Close()
	io.WriteString(stdin, text)
    }()
    out, _ := cmd.CombinedOutput()
    return string(out)
}
func hyperref(tolabel string, text string) string {
    // try replacing slashes
    r := regexp.MustCompile(`/`)
    tolabel = r.ReplaceAllString(tolabel, ":")

    return "\\hyperref[" + tolabel + "]{" + text + "}"
}
func ref(label string) string {
    // try replacing slashes
    r := regexp.MustCompile(`/`)
    label = r.ReplaceAllString(label, ":")

    return "\\ref{" + label + "}"
}
func label(label string) string {
    // try replacing slashes
    r := regexp.MustCompile("/")
    label = r.ReplaceAllString(label, ":")
    return "\\label{" + label + "}"
}
func pageref(label string) string {
    // try replacing slashes
    r := regexp.MustCompile("/")
    label = r.ReplaceAllString(label, ":")
    return "\\pageref{" + label + "}"
}
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
func TexTemplate(header string) string {
    out := tmpltex
    if header != "" {
        // get the name of the header file without extension and path
        hstem := fc.Stem(header)
        out = strings.Replace(out, "<header>", hstem, 1)
    }
    return out
}
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
