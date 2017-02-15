#!/bin/bash
#
# A more complex example.
# Look for main in this project.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -s -v -l -e '.*\.log$' -p '/test$|/tmp$|\.git$' -a '\bmain\b' ..

