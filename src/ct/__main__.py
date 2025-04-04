# __main__ kicks off ct from command line and python -m

import sys
from ct import ct
import argparse

# main kicks off tangling
def main():

    # parse the command line argument(s)
    parser = argparse.ArgumentParser()

    # the .ct file
    parser.add_argument("ct_file", help="codetext file")
    # optional generated file
    parser.add_argument("generated_file", help="go to line number from generated file in ct", nargs="?") 
    # optional line number in generated file
    parser.add_argument("line_number", help="line number from generated file", nargs="?")

    args = parser.parse_args()

    f = args.ct_file
    
    # normal compilation
    if args.generated_file is None:
        ct.main(open(f))
    elif len(sys.argv) == 4:
        # map from line number in generated source to original line number in ct

        # name of generated file
        ct.main(open(f))

        # print the original line number
        print(ct.ctlinenr[args.generated_file][args.line_number])

sys.exit(main())
