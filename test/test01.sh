#!/bin/bash
#
# Simple search of a single file.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -l -v -s -a 'test-utils' test01.sh

