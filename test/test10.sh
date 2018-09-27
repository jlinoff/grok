#!/bin/bash
#
# Test options parsing for trailing condensed options.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

# Strip off the version info so that this test does not have to be
# updated everytime the version changes.
$PUT -CWyza 2 4 'Lorem ipsum dolor|malesuada' test10.txt -lV | sed -e 's/version.*$/version/'
