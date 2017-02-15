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

$PUT -M 1 -s -v -l -e '\.gold$|\.filter$|log$' -p '/src$|/tmp$|\.git$' -A '\bmain\b' -A '\bFOOBARSPAM\b' .. 2>&1 | grep -v ' - files tested: '

