// tc.go does reverse-codetext, builing a ct file from source files.

// variables

// stack holds the nct (number in ct file) of the current chunk
// for pop see https://yourbasic.org/golang/implement-stack/
var stack []int

// chunks holds the chunk at each nct (number in ct file)
// in the end, iterating the chunks in order gives the contents of the ct file
var chunks map[int]*Chunk

// lastopened holds the nct (number in ct file) of the last opened chunk
var lastopened int

// Tc runs reverse-codetext (building ct from generated source)
func Tc(genfiles []string) {

    // reset variables
    stack = []int
    chunks = make(map[int]*Chunk)
    lastopened = -1
}

