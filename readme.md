# codetext

codetext is a way of writing computer programs where your code is
embedded in text like graphics or images floating in a document. it's
a bit like a jupyter notebook with the addition that code chunks are
named and you can nest them into one another by referencing the names.

here's an example:

```
begin a file named foo.py and alias it as foo.

<</foo.py: foo
print("ooh, baby you're a fool to py") 
if bar {
  <<bar>>
}
>>

the preceeding / in /foo.py marks this as a file-root.

chunk names are noted as paths, e.g. the path 'foo/bar' for a chunk
named 'bar' in the chunk 'foo'. we put some code into the 'foo/bar'
chunk.

<<foo/bar
   print("my bar")
   <<baz>>
   <<boz>>
>>

we can use relative paths and reference the previous chunk via .

<<./baz
   my = "baz code"
>>

this would be baz' absolute path:

<<foo/bar/baz
   print("and it makes me wonder why")
>>

when we don't give a path we append to the same chunk.

<<
   print("still my baz code.")
>>

if we would like to change to baz's sibling boz now, we could say
../boz, foo/bar/boz, or foo/*/boz, if boz's name is unique in foo.

<<foo/*/boz
   print("in boz")
>>

if there's a loop etc, and we would like the next unnamed chunk in the
text to be inside the loop instead of appended to the end of the chunk
we can say <<.>>:

<<
   for i = 0; i < n; i++ {
      <<.>>
   }
>>

then the following chunk will be put where the <<.>> tag is and not to
the end of the chunk.

<<
   print("inside the loop")
>>

go back via ..

<<..
   print("appending to the foo/bar/baz code again")
>>

we open a second file, named zoo.py:

<</zoo.py: zoo
  print("welcome to the zoo")
>>

if there's only one output file in the codetext and if you don't give
this file an alias, you leave out its name when referring to child
chunks, otherwise you include it like above.

```

`foo.ct` contains the above example. you can extract `foo.py` and `zoo.py`
by saying `ct foo.ct` (after building and installing).

## build

build with python:

```
python3 -m build
```

## install

install with pip:

```
pip install dist/*.whl
```

## issues

codetext takes a chunk's indent from the first chunk line, and tries
to fill up accordingly. if you write chunks that are already indented,
take care that the first line is indented like the rest of the chunk,
so that codetext doesn't think it needs to indent when it actually
doesn't need to.

fix: allow dashes in filename (-)

## possible next steps

a codetext editor plugin that infers the programming language
of a chunk by the file-root it hangs on and syntax-highlights
it accordingly might be handy. it could be written on the basis of
tangle.py and existing org plugins. maybe it could offer to indent
code-chunks based on their parent chunk.