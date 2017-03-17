#!/bin/bash
#
# Test configuration files.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

$PUT -c test07.conf test07.sh

