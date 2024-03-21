# codetext

codetext is a way of writing computer programs where your code is
embedded in text like graphics or images embedded in a document.

here's an example:

begin a file named foo.py and alias it as foo.

<<foo.py: foo
   ooh, baby you're a fool to py
   <<bar>>
>>

chunk names are written as paths, e.g. foo/bar for the chunk bar in
the chunk foo. we put some code into the foo/bar chunk and give it
another child, baz.

<<foo/bar
   my bar code
   <<baz>>
>>

we can use relative paths and reference the previous chunk via .

<<./baz
   my baz code
>>

this would be baz' absolute path:

<<foo/bar/baz
   and it makes me wonder why
>>

when we don't give a path we append to the previous chunk.

<<
   still my baz code.
>>

if there's a loop etc, and we would like the next unnamed chunk in the
text to be inside the loop instead of being appended at the end of the chunk
we can say <<.>>:

<<
   for i = 0; i < n; i++ {
      <<.>>
   }
>>

then the following chunk will be put where the <<.>> tag is and not be appended at the end of the chunk.

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

if there's only one root-file in the codetext and if you don't give
this file an alias, leave out its name when referring to child
chunks, otherwise you include root-name or alias like above.

if you save this in foo.ct you can extract foo.py and zoo.py via `./ct
foo.ct`.

codetext isn't supported by editors yet. editors like vs studio,
sublime and emacs do support the .org format via plugins. in org, the
opening line of a codechunk is `#+begin_src <language> <chunk
name/path>` the closing line is `#+end_src`. chunk naming and
referencing works the same. extract code from org files with `orgct
mycode.org`. switch between formats with `cttoorg <lang>` and
`orgtoct`.

## possible next steps

a ct-plugin for vs code / sublime / emacs or other editor on the basis
of tangle.py and existing org plugins that infers the programming
language of a chunk by the file-root it hangs on and syntax-highlights
accordingly, and that maybe offers to indent code-chunks based on
their parent chunk.