# codetext

code with text (vscode extension [here](https://github.com/tnustrings/ct-vscode)).

codetext lets you write computer programs where the code is
embedded in text like graphics or images floating in a document. it's
a bit like a jupyter notebook with the addition that code chunks are
named and you can use the names to nest chunks.

here's an example:

```
begin a file named foo.py.

``//foo.py:
print("ooh, baby you're a fool to py") 
if bar {
  ``bar``
}
``

when we open a chunk for the first time, like //foo.py above,
it is followed by a colon.

the preceeding // in //foo.py marks this code chunk as a file root.

chunk names are noted as paths, e.g. /bar for the chunk named bar in
the last opened file. we put some code into the /bar chunk.

``/bar:
   print("my bar")
   ``baz``
   ``boz``
``

we can use relative paths and reference the previous chunk (/bar) via .

``./baz:
   my = "baz code"
``

this would be baz' absolute path:

``/bar/baz
   and it makes me 
``

this appends to the baz chunk. when we append, we don't write the
colon after the chunk name.

this would be it's path starting from the file:

``//foo.py/bar/baz
   wonder why
``

when we don't give a path we append to the same chunk.

``
   print("still my baz code.")
``

if we would like to change to baz's sibling boz now, we could say
../boz, /bar/boz, //foo.py/bar/boz or /*/boz, if boz's name is unique
in foo.

``/*/boz:
   print("in boz")
``

if there's a loop etc, and we would like the next unnamed chunk in the
text to be inside the loop instead of appended to the end of the chunk
we can say ``.``:

``
   for i = 0; i < n; i++ {
      ``.``
   }
``

then the following chunk will be put where the ``.`` tag is and not to
the end of the chunk.

``
   print("inside the loop")
``

go back via ..

``
   print("appending to the foo/bar/baz code again")
``

we open a second file, named zoo.py.

``//zoo.py:
  welcome to the zoo
  ``dolphins``
``

now the last opened file is zoo.py, so /dolphins takes us to the chunk in zoo.py

``/dolphins:
  print("are there dolphins in the zoo?")
``

if you'd like to switch back to foo.py like this:

``//foo.py
  print("hello foo again")
``

if there's only one output file in the codetext and if you don't give
this file an alias, you leave out its name when referring to child
chunks, otherwise you include it like above.

```

`foo.ct` contains the above example. you can assemble `foo.py` and
`zoo.py` by saying `ct foo.ct`.

## install

install with pip:

```
pip install dist/*.whl
```

## dev

build with python:

```
python3 -m build
```

## issues

codetext takes a chunk's indent from the first chunk line, and tries
to fill up accordingly. if you write chunks that are already indented,
take care that the first line is indented like the rest of the chunk,
so that codetext doesn't think it needs to indent when it actually
doesn't need to.

fix: allow dashes in filename (-)

