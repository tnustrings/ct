# ct - codetext

code with text (vscode extension [install](https://marketplace.visualstudio.com/items?itemName=tnustrings.codetext) or [github](https://github.com/tnustrings/ct-vscode)).

ct lets you embed code in text. it's a bit like jupyter notebooks with
the addition that code chunks are named and can be nested like
directories.

here's an example:

```
begin a file named foo.py.

``//foo.py: #py
print("ooh, baby you're a fool to py") 
if bar {
  ``bar``
}
``

the line

   ``//foo.py: #py

opens a code chunk named //foo.py. the two slashes // mark foo.py as a
file. the colon : indicates that this chunk is opened for the first
time. #py signals syntax highlighting, you can leave it out.

the chunk //foo.py references a child-chunk, ``bar``. we put code into
bar by opening a second code chunk, /bar.

``/bar: #py
   print("my bar")
   ``baz``
   ``boz``
``

bar in turn references two child chunks, ``baz`` and ``boz``. to put
code into baz, you could use the full path /bar/baz or the relative
path starting from bar, ./baz.

``./baz: #py
   my = "baz code"
``

if you'd like to stay in the same chunk you can leave out its path.

`` #py
   print("still in baz")
``

if you'd like to change to baz's sibling chunk boz, you could say
../boz, /bar/boz, //foo.py/bar/boz or /*/boz, if boz's name is unique
in foo.

``../boz: #py
   print("in boz")
``

if there's a loop or so, and you'd like the next unnamed chunk to be
inside of the loop you can specify where it should be put with ``.``:

`` #py
   for i = 0; i < n; i++ {
      ``.``
   }
``

now the following unnamed chunk won't be appended to the end of the
previous chunk but instead where the ``.`` is.

`` #py
   print("inside the loop")
``

you need to exit this 'ghost'-chunk via ../ to append code after the
loop again:

``../ #py
   print("after the loop")
``

we open a second file, named zoo.py.

``//zoo.py: #py
  welcome to the zoo
  ``dolphins``
``

chunk paths assume to start at the last opened file (unless the
filename is explicitly given). our last opened file is zoo.py now, so
the path /dolphins adds code to zoo.py.

``/dolphins: #py
  print("are there dolphins in the zoo?")
``

you can switch back to a chunk in foo.py like this:

``//foo.py/bar #py
  print("hello bar again")
``

now our file is assumed to be foo.py again.

``
  print("still in foo.py")
``

```

the file `foo.ct` contains the above example. you can assemble your
code chunks into `foo.py` and `zoo.py` by saying `ct foo.ct`.

## use

assemble code from a ct file.

```
ct foo.ct
```

get line number from an assembled file in the .ct file. here, line 1
of `foo.py` is on line 6 in `foo.ct`.

```
ct -g foo.py:1 foo.ct
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

## install

**on linux**

download [ct-\<version\>.deb](https://github.com/tnustrings/codetext/releases) and install with apt:

```
sudo apt install ./ct-<version>.deb
```

**on any os with go**

install go from [here](https://go.dev/doc/install), then run

```
go install github.com/tnustrings/codetext/cmd/ct@latest
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

error ct -g wiki.go:25 wiki.ct
2025/05/06 11:30:43 there is no wiki.go

but there is a wiki.go