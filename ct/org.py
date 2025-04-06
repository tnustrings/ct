# org transforms org format to ct

import re
# import sys

# orgtoct turns org to ct text
def orgtoct(text: str) -> str:

    lines = text.split("\n")

    # we are not in source
    insrc = False

    # the out text
    out = ""

    for line in lines:
        line += "\n" # add the \n again that split removes
        
        # we encounter a doc chunk (without chunk name in <<>>= brackets)
        if re.match(r"(?i)^#+begin_src[^<]*\n$", line):
            print("doc found")
            # leave it out
            line = ""
        # we encounter a source chunk
        # replace #+begin_src lang with ct chunk opening
        elif re.match(r"(?i)^#+begin_src\s+\w+\s+<<", line): # (?i): case insensitive
            insrc = True
            # remov the begin_src
            line = re.sub(r"(?i)^#+begin_src\s+", "", line)
            # capture the programming language
            pl = re.search(r"\w+", line).group()
            # remove the programming language and beginning of name <<
            line = re.sub(r"\w+ <<", "", line)
            # remove the end marker >>=
            line = re.sub(r">>=\s*$", "", line)
            # now line should only hold the chunkname
            name = line
            # construct the ct-line
            line = "``" + name + " #" + pl + "\n"
        # the end_src tag of either doc- or code chunk
        if re.match(r"(i?)^#+end_src", line):
            # is it the end of a doc chunk? leave it out
            if not insrc:
                line = ""
            else: # it's the end of a source chunk, mark it with ``
                insrc = False
                line = "``\n"

        # if in source
        if insrc:
            # we remove the two leading blanks org-mode adds
            line = re.sub(r"^  ", "", line)

            # we change the brackets surrounding the child references.
            

        out += line

    return out


# sys.exit(main())
