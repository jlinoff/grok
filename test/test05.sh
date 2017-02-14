#!/bin/bash
#
#
# Check file name matching exclude rules.
#
#    test05-include.txt
#    test05-exclude.txt
#    test04.txt
#
#

# ================================================================
# Includes
# ================================================================
Location="$(cd $(dirname $0) && pwd)"
source $Location/test-utils.sh

echo "test05-include.txt" >test05-include.txt
echo "test05-exclude.txt" >test05-exclude.txt
echo "test05.txt" >test05.txt
$PUT -s -v -l -a 'test05' -e 'test05-include' test05*.txt

rm -f test05*.txt
