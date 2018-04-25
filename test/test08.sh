#!/bin/bash
#
# Test configuration files.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -CWla waldo test08.txt
