#!/bin/bash
#
# Test configuration files.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -CWlyza 2 4 'Lorem ipsum dolor|malesuada' test09.txt
