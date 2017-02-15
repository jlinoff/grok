#!/bin/bash
#
# Run tests.
#
# Note that the bash shell operations have been dumbed down to make
# sure that they are compatible with the very old version of bash
# used by Mac OSX.
#

# ================================================================
# Functions
# ================================================================
function Fail() {
    (( Failed++ ))
    (( Total++ ))
    local Tid="$1"
    local Memo="$2"
    printf "test:%03d:%s:failed %s\n" $Total "$Tid" "$Memo"
}

function Pass() {
    (( Passed++ ))
    (( Total++ ))
    local Tid="$1"
    local Memo="$2"
    printf "test:%03d:%s:passed %s\n" $Total "$Tid" "$Memo"
}

# ================================================================
# Main
# ================================================================
Passed=0
Failed=0
Total=0

for t in $(ls -1 test[0-9][0-9]*.sh) ; do
    Test=$(echo "$t" | sed -e 's/\.sh//')
    n=0
    Log=$Test.log
    ./$t > $Log 2>&1
    st=$?
    if (( $st )) ; then
        Fail $Test "Run failed with status $st - see $Log"
    else
        Pass $Test "Run passed"
        (( n++ ))
    fi

    # Filter out the date info.
    DiffLog=$Test.difflog
    
    #
    cat $Test.log | \
        sed -E \
            -e 's@^[ ]+[0-9]+ \| @  LINENO \| @' \
            -e 's@INFO[ ]+[0-9]+ -@INFO  LINENO -@' \
            -e 's@^[0-9]+/[0-9]+/[0-9]+ [0-9]+:[0-9]+:[0-9]+@YYYY/MM/DD hh:mm:ss@' \
        | \
        grep -v 'summary: files tested : ' | \
        grep -v 'version: grok' \
             > $Test.log.filter
    cat $Test.gold | \
        sed -E \
            -e 's@^[ ]+[0-9]+ \| @  LINENO \| @' \
            -e 's@INFO[ ]+[0-9]+ -@INFO  LINENO -@' \
            -e 's@^[0-9]+/[0-9]+/[0-9]+ [0-9]+:[0-9]+:[0-9]+@YYYY/MM/DD hh:mm:ss@' \
        | \
        grep -v 'summary: files tested : ' | \
        grep -v 'version: grok' \
             > $Test.gold.filter

    diff $Test.log.filter $Test.gold.filter > $DiffLog 2>&1
    st=$?
    if (( $st )) ; then
        Fail $Test "Diff failed with status $st - see $DiffLog"
    else
        Pass $Test "Diff passed"
        (( n++ ))
    fi

    if (( n == 2 )) ; then
        # Everything passed - clean up.
        rm -f $Log $DiffLog $Test.*.filter $Test.log
    fi
done

echo
printf "test:total:passed  %3d\n" $Passed
printf "test:total:failed  %3d\n" $Failed
printf "test:total:summary %3d\n" $Total

echo
if (( Failed )) ; then
    echo "FAILED"
else
    echo "PASSED"
fi
exit $Failed
