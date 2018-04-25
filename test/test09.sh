#!/bin/bash
#
# Test configuration files.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -CWlyza 3 3 'Lorem ipsum dolor|malesuada' test09.txt
