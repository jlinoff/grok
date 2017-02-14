#!/bin/bash
#
# A more complex example.
# Look for main
# and FOOBARSPAM in this project.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -s -v -l -e '.*\.log$' -e '/test/' -e '/tmp/' -A '\bmain\b' -A '\bFOOBARSPAM\b' ..

