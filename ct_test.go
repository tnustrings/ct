package ct
import (
"testing"
"os"
)
func TestCtGen(t *testing.T) {
    b, _ := os.ReadFile("try/foo.ct")
    text := string(b)
    
    //files, err := CtGen("try/foo.ct")

    err := Ctwrite(text, "try", "foo.ct")
    if err != nil {
      t.Errorf("error: %v", err)
    }
    b, _ = os.ReadFile("try/zoo.py")
    text = string(b)
    if text != `# zoo.py is automatically generated from foo.ct. please edit foo.ct.
  welcome to the zoo
  print("are there dolphins in the zoo?")
` {
        t.Errorf("generated text for try/zoo.py is different than expected.")
    }
}
