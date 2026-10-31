# ct - codetext

code with text (vscode extension [here](https://marketplace.visualstudio.com/items?itemName=tnustrings.codetext)).

ct lets you embed code in text. it's a bit like jupyter notebooks with
the addition that code chunks are named and can be nested like
directories.

here's an example: 

```
begin a file named foo.py.

``//foo.py: #py
import sys
print("ooh, baby you're a foo to py.") 
if len(sys.argv) > 1:
    ``bar``
``
write something code in the ``bar`` chunk:

``/bar:
    print("hi, " + sys.argv[1])
``
```

if this is saved as `minifoo.ct`, extract the code by calling `ct minifoo.ct`.

## install

**on any os with go**

install go from [here](https://go.dev/doc/install), then run

```
go install github.com/tnustrings/ct/cmd/ct@latest
```

**.deb on linux**

download [ct-\<version\>.deb](https://github.com/tnustrings/codetext/releases) and install with apt:

```
sudo apt install ./ct-<version>.deb
```

## longer example

this is a longer ct example.

```
begin a file named foo.py.

``//foo.py: #py
import sys
print("ooh, baby you're a foo to py") 
if len(sys.argv) > 1:
    ``bar``
``

the line

   ``//foo.py: #py

opens a code chunk named //foo.py. the two slashes // mark foo.py as a
file. the colon : indicates that this chunk is opened for the first
time. #py signals syntax highlighting, you can leave it out.

the chunk //foo.py references a child-chunk, ``bar``. put code into
the bar chunk by opening a second code chunk, /bar.

``/bar: #py
    print("welcome to bar, " + sys.argv[1])
    ``baz``
    ``boz``
``

bar references two child chunks, ``baz`` and ``boz``. to put code into
baz, use the full path /bar/baz or the relative path from bar, ./baz.

``./baz: #py
    my = "baz code"
``

to stay in the same chunk leave out the path.

`` #py
    print("still in baz")
``

to change into baz's sibling chunk boz, say ../boz.

``../boz: #py
    print("in boz")
``

instead of ../boz you could get to the boz chunk with /bar/boz,
//foo.py/bar/boz or, if boz's name is unique in foo, /*/boz.

if there's a loop or so, and the following code should go into the
loop, specify where the code should go with a ghost chunk ``.``:

`` #py
    for i in range(10):
        ``.``
``

now the following chunk will be appended where the ``.`` is.

`` #py
        print("inside the loop")
``

here, this doesn't make much of a difference, but it can come in handy
for c-style loops that come with opening and closing brackets:

for (i = 0; i < 10; i++) {
    ``.``
}

exit the ``.`` ghost-chunk with ../ to continue code after the loop
(or use the explicit path /bar/boz).

``../ #py
    print("after the loop")
``

open a second file, named zoo.py.

``//zoo.py: #py
print("welcome to the zoo.")
``dolphins``
``

chunk paths assume to start at the last opened file (unless the
filename is explicitly given). our last opened file is zoo.py now, so
the path /dolphins adds code to zoo.py.

``/dolphins: #py
print("are there dolphins in the zoo?")
``

switch back to a chunk in foo.py like this. (notice that the colon is
lacking cause we opened the chunk before).

``//foo.py/bar #py
    print("hello bar again")
``

now we are in foo.py again.

``
    print("still in foo.py")
``

```

the file [`try/foo.ct`](try/foo.ct) contains the above example. you
can assemble your code chunks into `foo.py` and `zoo.py` by saying `ct
try/foo.ct`.

## use

assemble code from a ct file.

```
ct foo.ct
```

get line number from an assembled file in the .ct file. here, line 1
of `foo.py` is on line 6 in `foo.ct`.

```
ct -l foo.py:1 foo.ct
6
```

generate latex from `foo.ct` named `foo-in.tex`.

```
ct -tex -o pdf/foo-in.tex foo.ct
```

if you leave out the ct file, you generate a latex document wrapper.

```
ct -tex -o pdf/foo.tex
```

include `foo-in.tex` in `pdf/foo.tex`.

```
\input{foo-in}
```

## dev

build:

```
make
```

## issues

codetext takes a chunk's indent from the first chunk line, and tries
to fill up accordingly. if you write chunks that are already indented,
take care that the first line is indented like the rest of the chunk,
so that codetext doesn't think it needs to indent when it actually
doesn't need to.

fix: allow dashes in filename (-)

error ct -l wiki.go:25 wiki.ct
2025/05/06 11:30:43 there is no wiki.go

but there is a wiki.go
