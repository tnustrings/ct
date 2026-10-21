// org.go was generated from org.ct. 
// if you edit org.go, run `ct -tc org.go` afterwards to update org.ct.
// when editing this file leave the `` tags.
//``ct version 0.1.1
//``ct files org.ct org.go

// # org
// 
// transforms org format to ct  1``//org.go:
package ct
import (
    // import strings and regexp.  11``/import:
    "strings"
    "regexp"
    //``11
)
// orgtoct turns org to ct text.  2``/orgtoct:
func Orgtoct(text string) string {
    // get the lines, take care of dos \r\n linebreaks.  3.``
    lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
    // init the loop variables.  4``
    // are we in a code chunk?
    insrc := false
    // are we in a doc chunk?
    indoc := false

    // the out text
    out := ""
    // loop over the lines.  5``
    for _, line := range lines {
        line += "\n" // add the \n again that split removes

        rdoc := regexp.MustCompile(`(?i)^#\+begin_src[^<]*\n$`) // (?i): case insensitive
	rchunk := regexp.MustCompile(`(?i)^#\+begin_src\s+\w+\s+<<`)
	rend := regexp.MustCompile(`(i?)^#\+end_src`)
        // we encounter a doc chunk (without chunk name in <<>>= brackets)
        if rdoc.MatchString(line) {
	    // we encounter a doc chunk (without chunk name in <<>>= brackets).  6``./between text:
            // leave the begin_src line out
            line = ""
            // we're in a doc chunk
            indoc = true
	    //``6
        } else if rchunk.MatchString(line) { 
	    // we encounter a source chunk.  7``../code chunk:
            insrc = true
            // remove the begin_src
	    r1 := regexp.MustCompile(`(?i)^#\+begin_src\s+`)
            line = r1.ReplaceAllString(line, "")
            // capture the programming language
	    r2 := regexp.MustCompile(`\w+`)
            pl := r2.FindString(line)
            // remove the programming language and beginning of name <<
	    r3 := regexp.MustCompile(`\w+ <<`)
            line = r3.ReplaceAllString(line, "")
            // remove the end marker >>=
	    r4 := regexp.MustCompile(`>>=\s*$`)
            line = r4.ReplaceAllString(line, "")
            // now line should only hold the chunkname
            name := line
            // construct the ct-line
            line = "``" + name + " #" + pl + "\n"
	    //``7
        } else if rend.MatchString(line) {
	    // if it's the end of a between-text, leave it out. if it's the end of a
	    // source chunk, mark it with ``.  8``../end src:
	    
	                if !insrc {
	                    line = ""
	                } else { 
	                    line = "``\n"
	                }
	                // it's either the end of a src or doc chunk
	                insrc=false
	                indoc=false
	    
	    //``8
	}

        // in both source and doc chunks we remove the two leading blanks
        // org-mode adds.  9``../add line:
        if insrc || indoc {
            r := regexp.MustCompile("^  ")
            line = r.ReplaceAllString(line, "")
        }
        if insrc {
            // we change the brackets surrounding the child references.
            // only allow spaces before and after?
	    r1 := regexp.MustCompile(".*<<.+>>.*")
            if r1.MatchString(line) {
                line = strings.Replace(line, "<<", "``", 1)
		line = strings.Replace(line, ">>", "``", 1)
            }
	}
        out += line
        //``9
    }
    // after the loop return.  10``../
    return out
    //``3
}
//``2
//``1
