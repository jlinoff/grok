# Utilities used by the macports tools.
#
# Note that the bash shell operations have been dumbed down to make
# sure that they are compatible with the very old version of bash
# used by Mac OSX.
#

# ================================================================
# Functions
# ================================================================
# Print an info message with context (caller line number)
function info() {
    local Msg="$*"
    echo -e "INFO:${BASH_LINENO[0]}: $Msg"
}

# Print a warning message with context (caller line number)
function warn() {
    local Msg="$*"
    echo -e "WARNING:${BASH_LINENO[0]}: $Msg"
}

# Print an error message and exit.
function err() {
    local Msg="$*"
    echo -e "ERROR:${BASH_LINENO[0]}: $Msg"
    exit 1
}

# Run a command with decorations.
function runcmd() {
    local Cmd="$*"
    local LineNum=${BASH_LINENO[0]}
    echo
    echo "INFO:${LineNum}: cmd.run=$Cmd"
    eval "$Cmd"
    local st=$?
    echo "INFO:${LineNum}: cmd.status=$st"
    if (( st )) ; then
        echo "ERROR:${LineNum}: command failed"
        exit 1
    fi
}

# Run a command but allow the user to specify an acceptable
# return code range.
function runcmdst() {
    local Lower=$1
    local Upper=$2
    shift
    shift
    local Cmd="$*"
    local LineNum=${BASH_LINENO[0]}
    echo
    echo "INFO:${LineNum}: cmd.run=$Cmd"
    eval "$Cmd"
    local st=$?
    echo "INFO:${LineNum}: cmd.status=$st OK=[$Lower..$Upper]"
    if (( st < Lower )) || (( st > Upper )) ; then
        echo "ERROR:${LineNum}: command failed"
        exit 1
    fi
}


# ================================================================
# Pre-flight setup.
# ================================================================
Name=$(basename "$0" | cut -d. -f1)
PUT=../bin/grok


