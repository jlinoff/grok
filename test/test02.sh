#!/bin/bash
#
# A more complex example.
# Look for main in this project.
#
# Use -M 1 to make it deterministic.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -M 1 -s -v -l -e '.*\.log$' -p '/src/github.com$|/src/golang.org$|/test$|/tmp$|\.git$' -a '\bmain\b' .. 2>&1 | grep -v ' - files tested: '

