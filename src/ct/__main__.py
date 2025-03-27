# __main__ kicks off ct from command line and python -m

import sys
from ct import tangle

# main kicks off tangling
def main():
    f = sys.argv[1]

    # normal compilation
    if len(sys.argv) == 2:
        tangle.main(open(f))
    elif len(sys.argv) == 4:
        # map from line number in generated source to original line number in ct

        # name of generated file
        genfilename = sys.argv[2]
        genlinenr = int(sys.argv[3])
        tangle.main(open(f))

        # print the original line number
        print(tangle.ctlinenr[genfilename][genlinenr])

sys.exit(main())
