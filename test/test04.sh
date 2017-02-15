#!/bin/bash
#
#
# Check file name matching include rules.
#
#    test04-include.txt
#    test04-exclude.txt
#    test04.txt
#
# Use -M 1 to make it deterministic.
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

echo "test04-include.txt" >test04-include.txt
echo "test04-exclude.txt" >test04-exclude.txt
echo "test04.txt" >test04.txt
$PUT -M 1 -s -v -l -a 'test04' -i 'test04-include' test04*.txt

rm -f test04*.txt
