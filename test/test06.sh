#!/bin/bash
#
#
# Check matching with no file name matching rules.
#
#    test06-include.txt
#    test06-exclude.txt
#    teest06.txt
#
# Use -M 1 to make it deterministic.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

echo "test06-include.txt" >test06-include.txt
echo "test06-exclude.txt" >test06-exclude.txt
echo "test06.txt" >test06.txt
$PUT -M 1 -s -v -l -a 'test06' test06*.txt

rm -f test06*.txt
