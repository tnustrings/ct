package ct
import (
    "strings"
    "regexp"
)
func Orgtoct(text string) string {
    lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
    // are we in a code chunk?
    insrc := false
    // are we in a doc chunk?
    indoc := false

    // the out text
    out := ""
    for _, line := range lines {
        line += "\n" // add the \n again that split removes

        rdoc := regexp.MustCompile(`(?i)^#\+begin_src[^<]*\n$`) // (?i): case insensitive
	rchunk := regexp.MustCompile(`(?i)^#\+begin_src\s+\w+\s+<<`)
	rend := regexp.MustCompile(`(i?)^#\+end_src`)
        // we encounter a doc chunk (without chunk name in <<>>= brackets)
        if rdoc.MatchString(line) {
            // leave the begin_src line out
            line = ""
            // we're in a doc chunk
            indoc = true
        } else if rchunk.MatchString(line) { 
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
        } else if rend.MatchString(line) {
	    
	                if !insrc {
	                    line = ""
	                } else { 
	                    line = "``\n"
	                }
	                // it's either the end of a src or doc chunk
	                insrc=false
	                indoc=false
	    
	}

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
    }
    return out
}
