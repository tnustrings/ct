# __main__ kicks off ct called from command line and python -m

import sys
from ct import tangle

# main kicks off tangling
def main():
    tangle.main(open(sys.argv[1]))

sys.exit(main())
