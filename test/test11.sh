#!/bin/bash
#
# Test options for deleting accepted patterns.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

# Strip off the version info so that this test does not have to be
# updated everytime the version changes.
$PUT -CWl -a '"([^\\\"]|.)*"' -a "'([^\\']|.)*'" -d '^\s*#' test11.txt
