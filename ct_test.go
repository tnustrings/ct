// ct_test.go was generated from ct_test.ct. 
// if you edit ct_test.go, run `ct -tc ct_test.go` afterwards to update ct_test.ct.
// when editing this file leave the `` tags.
//``ct version 0.1.1
//``ct files ct_test.ct ct_test.go

// ct_test tests ct
// 
// see https://go.dev/doc/tutorial/add-a-test  1``//ct_test.go:
package ct
import (
// the imports.  5``/import:
"testing"
"os"
//``5
)
// TestCtGen tests generating ct code.  2``/TestCtGen:
func TestCtGen(t *testing.T) {
    // for now take foo.ct to generate code from ct.  3.``
    b, _ := os.ReadFile("try/foo.ct")
    text := string(b)
    
    err := Ctwrite(text, "try", "foo.ct")
    if err != nil {
      t.Errorf("error: %v", err)
    }
    // maybe in the future call like this:
    // 
    // files, err := CtGen("try/foo.ct")
    // 
    // read one of the generated files and check its contents.  4``
    b, _ = os.ReadFile("try/zoo.py")
    text = string(b)
    if text != `# zoo.py is automatically generated from foo.ct. please edit foo.ct.
  welcome to the zoo
  print("are there dolphins in the zoo?")
` {
        t.Errorf("generated text for try/zoo.py is different than expected.")
    }
    //``3
}
//``2
//``1
