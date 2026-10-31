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
// the imports.  14``/import:
"testing"
"os"
"io"
//``14
)
// TestCtGen tests generating ct code.
// 
// if something in the code generating changed (including ct version),
// manually update try/try_should.go and test/__init_should.py.  2``/TestCtGen:
func TestCtGen(t *testing.T) {
    // for now take try.ct to generate code from ct.  3.``
    err := Ctwrite("try/try.ct")
    if err != nil { t.Errorf("error: %v", err) }
    // maybe in the future call like this:
    // 
    // files, err := CtGen("try/foo.ct")
    // 
    // read one of the generated files and check its contents.  4``
    b, _ := os.ReadFile("try/try.go")
    gen := string(b)
    b, _ = os.ReadFile("try/try_should.go")
    should := string(b)
    if gen != should {
        t.Errorf("try/try.go and try/try_should.go don't match.")
    }
    // try for the init.ct of traction.  5``
    err = Ctwrite("test/init.ct")
    if err != nil { t.Errorf("error: %v", err) }
    // test the contents.  6``
    b, _ = os.ReadFile("test/__init__.py")
    gen = string(b)
    b, _ = os.ReadFile("test/__init__should.py")
    should = string(b)
    if gen != should {
        t.Errorf("test/__init__.py and test/__init__should.py don't match.")
    }
    //``3
}
//``2
// TestTc tests reversing ct.
// 
// if something in the reverse output changed, update test/init_should.ct
// manually.  7``/TestTc:
func TestTc(t *testing.T) {
    // try for the init.ct of traction.  8``_:
    err := Tcwrite([]string{"test/__init__.py"}, "test/init.ct")
    if err != nil { t.Errorf("error: %v", err) }
    // test the contents.  9``
    b, _ := os.ReadFile("test/init.ct")
    gen := string(b)
    b, _ = os.ReadFile("test/init_should.ct")
    should := string(b)
    if gen != should {
        t.Errorf("test/init.ct (generated with -tc) and test/init_should.ct don't match.")
    }
    // turn back.  10``
    Cp("test/init_should.ct", "test/init.ct")
    //``8
}
//``7
// Cp copies filea to fileb.
// 
// TODO search whether there is a go library for common posix tasks?  11``/Cp:
func Cp(filea string, fileb string) error {
    // read the files.  12``_:
    src, _ := os.Open(filea)
    defer src.Close()
    dst, _ := os.Open(fileb)
    defer dst.Close()
    // copy.  13``
    _, err := io.Copy(src, dst)
    return err
    //``12
}
//``11
//``1
