# __main__ kicks off ct from command line and python -m

import sys
from ct import ct, tex, org
import argparse

# main kicks off tangling
def main():

    # parse the command line argument(s)
    parser = argparse.ArgumentParser()

    # the .ct file
    parser.add_argument("ct_file", help="codetext file")
    # optional generated file
    parser.add_argument("generated_file", help="generated file line number in ct", nargs="?") 
    # optional line number in generated file
    parser.add_argument("line_number", help="generated file line number", nargs="?")
    parser.add_argument("--tex", help="print doc as latex", action="store_true") 
    parser.add_argument("--from-org", help="input is a .org file", action="store_true")
    parser.add_argument("--mdtotex", help="for latex doc generation. a command to convert markdown between codechunks to tex, e.g. 'pandoc -f markdown -t latex'", nargs=1)
    parser.add_argument("--shell", help="run mdtotex command as shell script.", action="store_true")

    args = parser.parse_args()

    f = args.ct_file
    text = open(f).read()

    # is text in org format? convert it to ct format
    if args.from_org:
        print("from org")
        text = org.orgtoct(text)
        print(text)

    # normal compilation
    if args.generated_file is None:
        if args.tex == True:
            # run ct and print tex
            # args.mdtotex seems to be a string array
            tex.printtex(text, mdtotex=args.mdtotex, shell=args.shell) # todo maybe print(tex.totex(text))
        else:
            # run ct and write files
            ct.ctwrite(text)
    elif len(sys.argv) == 4: # todo change this to if args.line_number is not None
        # map from line number in generated source to original line number in ct

        # run ct without writing files
        ct.ct(text)

        # print the original line number
        print(ct.ctlinenr[args.generated_file][args.line_number-1]) # line numbers are 0-indexed

sys.exit(main())
