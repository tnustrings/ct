# codetext

codetext is a way of writing computer programs where your code is
embedded in text like graphics or images embedded in a document.

here's an example:

```
begin a file named foo.py and alias it as foo.

<<foo.py: foo
   <<import>>
   ooh, baby you're a fool to py
   <<bar>>
>>

chunk names are noted as paths, e.g. foo/bar for the chunk bar in
foo. we put some code into the foo/bar chunk.

<<foo/bar
   my bar code
   <<baz>>
   <<boz>>
>>

we can use relative paths and reference the previous chunk via .

<<./baz
   my baz code
>>

this would be baz' absolute path:

<<foo/bar/baz
   and it makes me wonder why
>>

when we don't give a path we append to the same chunk.

<<
   still my baz code.
>>

if we would like to change to baz's sibling boz now, we could say
../boz, foo/bar/boz, or foo/*/boz, if boz's name is unique in foo.

<<foo/*/boz
   in boz
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
   inside the loop
>>

go back via ..

<<..
   appending to the foo/bar/baz code again
>>

we open a second file, named zoo.py:

<<zoo.py: zoo
  welcome to the zoo
>>

if there's only one output file in the codetext and if you don't give
this file an alias, you leave out its name when referring to child
chunks, otherwise you include it like above.

```

`foo.ct` contains the above example. you can extract `foo.py` and `zoo.py`
by saying `./ct foo.ct`.

codetext isn't supported by editors (yet). a related format, .org, is
supported by [vs
code](https://marketplace.visualstudio.com/items?itemName=tootone.org-mode),
[sublime](https://packagecontrol.io/packages/orgmode) and
[emacs](https://orgmode.org/) via plugins. in .org, the opening line
of a codechunk is `#+begin_src <language> <chunk name/path>`. the
closing line is `#+end_src`. chunk naming and referencing works the
same. for the first chunk in the example above, this would look like:

```
#+begin_src python foo.py: foo
   ooh, baby you're a fool to py
   <<bar>>
#+end_src
```

extract code from .org files with `orgct mycode.org`. switch between
codetext and .org with `cttoorg <lang>` and `orgtoct`.

to get the codetext scripts into a directory you develop in, set
$CTPATH in `ctscripts` to your codetext dir, say `chmod u+x
ctscripts` and put the codetext dir on your $PATH. then in your target
directory you can say `ctscripts`. this creates a folder named
`ct`. in your makefile or on the commandline you can then say `./ct/ct
myprogram.ct`.

## possible next steps

a codetext editor plugin that infers the programming language
of a chunk by the file-root it hangs on and syntax-highlights
it accordingly might be handy. it could be written on the basis of
tangle.py and existing org plugins. maybe it could offer to indent
code-chunks based on their parent chunk.