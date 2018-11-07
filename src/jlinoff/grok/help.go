// Help message.
package main

import (
	"os"
	"path/filepath"
	"fmt"
	"runtime"
)

var helpMsg = `
USAGE
    %[1]v [OPTIONS] [<DIRS>]

DESCRIPTION
    Searches directory trees for files that match regular expressions
    using parallelism to improve performance. It is written in go.

    It was developed to allow me to find symbols in files in directory
    trees reasonably quickly which helps me grok the structure of
    the source code.  By default it searches in the current directory
    but you can explicitly specify the directories or files that
    you want to search.

    It is similar to doing a find/grep but the regular expressions
    are more powerful, the file name appears before the file content
    and multiple expressions can be search for simultaneously. Note
    that the search capability is much less powerful than the
    capabilities provided by find.

    You can specify whether to keep a file based on whether the
    file name matches or does not match a set of regular expressions,
    whether the file is older or newer than a date or whether the
    file content matches or does not match regular expressions. You
    can also limit the search by directory depth or directory path
    name.

    These ideas are summarized in the following table. The term REs
    refers to regular expressions.

        Test        Target     Action
        ==========  =========  =============================================
        accept      contents   Accept a file if the contents match REs.
        reject      contents   Reject a file if the contents match REs.

        include     name       Include a file if the file name matches REs.
        exclude     name       Exclude a file if the file name matches REs.

        newer       date/time  Accept a file if it is newer than a date/time.
        older       date/time  Accept a file if it is older than a date/time.

        maxdepth    depth      Exclude files deeper than the depth.
        prune       name       Exclude a directory if the path matched REs.

    You can specify whether a file must match all criteria (AND)
    or any criteria (OR). In the table above, you can see that with
    the options that are lower and upper case.

    A simple example should make all this a bit clearer. You want
    to search your python, java and C source files to see which
    ones do not have a copyright notice. The copy right notice has
    a very specific form:

        "Copyright (c) YEARS by Acme Inc., all rights reserved"

    The YEARS is a list of years that is a comma or dash separated
    list of years where each year is a 4 digit integer. This would
    be a valid list of years: 2004-2015, 2017. Spaces are allowed.

    The C files have .c and .h extensions. The java files have a
    .java extension and the python files have a .py extension.

    Here is the %[1]v command you might use.

        $ %[1]v \
            -s \
            -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
            -i '\.[ch]$|\.java$|\.py$'

    No directory was specified so the current directory tree will
    be searched to the bottom.

    The -s tells the program to print the summary statistics.

    The -r says to reject any file that contains the valid copyright
    notice so that we only print the ones with the valid copyright
    notice.

    The -i says to only include the files that have the specified
    extensions.

    The regular expression syntax is the same used by go. It is
    described here: https://github.com/google/re2/wiki/Syntax.

DATE/TIME SPECIFICATION
    The date/time specification is used by the -n and -o options
    to specify a relative date time. A specification consists of a
    positive integer with a suffix to indicate seconds, minutes,
    hours, days and weeks. To analyze all files that have been
    modified in the last week you would specify -n 1w or -n 7d.

    The table below lists the suffixes.

        S  Duration  Example
        =  ========  =======
        s  seconds   -n 30s
        m  minutes   -n 20m
        h  hours     -n 12h
        d  days      -n 3d
        w  weeks     -n 2w

    If no suffix is specified, seconds are assumed.

    You can search time windows by using both options. Here is an
    example that shows how to search files that are newer than 4
    weeks but older than 2 weeks:

        -n 4w -o 2w

    This is very useful when you only want to search a specific
    time window.

OPTIONS
    -a REGEXP, --accept REGEXP
                       Accept if the contents match the regular
                       expression.

                       If multiple accept criterion are specified,
                       only one of them has to match (an OR operation).

                       Here is an example that will accept a file
                       if it contains either foo or bar:
                           $ %[1]v -a foo -a bar  # same as -a 'foo|bar'
                           test/fooonly
                           test/baronly
                           test/foobar

    -A REGEXP, --Accept REGEXP, --ACCEPT REGEXP
                       Accept if the contents match the regular
                       expression.

                       If multiple accept criterion are specified,
                       all of them have to match (an AND operation).

                       Here is an example that will only accept a
                       file if it contains both foo and bar:
                           $ %[1]v -A foo -A bar .
                           test/foobar

    --after NUM, -z NUM
                       Print NUM lines after the match.

    --before NUM, -y NUM
                       Print NUM lines before the match.

    -b, --binary       Search binary files.

                       By default binary files are skipped.
                       The binary test is not very sophisticated.

                       It may designate UTF based document files as
                       binary.

    -B INT, --binary-size INT
                       Number of bytes to read to determine whether
                       this is a binary file.
                       The default is 512.

    -c CONF, --conf CONF
                       Read a conf file and insert the arguments
                       directly into the command line. This is
                       convenient for storing and re-using common
                       options. The syntax is simple, each line is
                       a single option. Blank lines and lines that
                       start with a hash are ignore.

                       Here is an example that prunes .git and .repo
                       files:
                           # Common options to prune common directories and
                           # disable warnings.
                           -p '\.git$|\.repo$'
                           -W

                       When this file is read, it is exactly like
                       specifying those options on the command line.

                       Nested conf files can be specified but the
                       program aborts if it finds nested references
                       to the same conf file.

    --color            Colorize the output using ANSI escape sequences.

    -d REGEXP, --delete REGEXP
                       Delete accepted entries if the contents match
                       the regular expression. This is extremely
                       useful because it allows single line matches to
                       be excluded.

                       If multiple delete criterion are specified,
                       only one of them has to match (an OR operation).

                       Here is an example that will accept a file
                       if it contains either foo or bar but not spam
                       on the same line:
                           $ %[1]v -a foo -a bar -d spam
                           test/fooonly
                           test/baronly
                           test/foobar

    -D REGEXP, --Delete REGEXP, --DELETE REGEXP
                       Delete accepted entries if the contents match
                       the regular expression.

                       If multiple delete criterion are specified,
                       all of them have to match (an AND operation).

                       Here is an example that will only accept a
                       file if it contains both foo and bar but not
                       spam and wombat on the same line:
                           $ %[1]v -A foo -A bar -D spam -D wombat
                           test/foobar

    -e REGEXP, --exclude REGEXP
                       Exclude file if the name matches the regular
                       expression.

                       If multiple exclude criterion are specified,
                       only one of them has to match (an OR operation).

                       Here is an example that will exclude a file
                       if its name contains either foo or bar:
                           $ %[1]v -e foo -e bar

    -E REGEXP, --Exclude REGEXP
                       Exclude file if the name matches the regular
                       expression.

                       If multiple exclude criterion are specified,
                       all of them have to match (an AND operation).

                       Here is an example that will exclude a file
                       if its name contains both foo and bar:
                           $ %[1]v -E foo -E bar
                           test/fooonly
                           test/baronly

    -h, --help         On-line help.

    -i REGEXP, --include REGEXP
                       Include file if the name matches the regular
                       expression.

                       If multiple include criterion are specified,
                       only one of them has to match (an OR operation).

                       Here is an example that will include a file
                       if its name contains either foo or bar:
                           $ %[1]v -i foo -i bar
                           test/fooonly
                           test/baronly
                           test/foobar
                           test/nofoobar

    -I REGEXP, --Include REGEXP
                       Include file if the name matches the regular
                       expression.

                       If multiple include criterion are specified,
                       all of them have to match (an AND operation).

                       Here is an example that will include a file
                       if its name contains both foo and bar:
                           $ %[1]v -i foo -i bar
                           test/foobar
                           test/nofoobar

    -l, --lines        Show the lines that match.
                       If this is not specified, only the file names
                       are shown. It is useful when using the tool
                       interactively.

    -m INT, --max-depth INT
                       The maximum depth in the directory tree.
                       The top level is 0.
                       The default is no maximum (0). All
                       subdirectories are processed.

    -M INT, --max-jobs INT
                       Maximum number of jobs (goroutines) to run
                       in parallel. Each job is a file analysis.
                       The default is %[2]v.

    -n DATE/TIME, --newer-than DATE/TIME
                       Only consider files that are newer than the
                       date/time specification. The specification
                       has a lot of options to make it simpler to
                       use. See the DATE/TIME SPECIFICATION section
                       for more details.

                       Here is an example that looks for files that
                       were modified in the last day:
                           $ %[1]v -n 1d

    -o DATE/TIME, --older-than DATE/TIME
                       Only consider files that are older than the
                       date/time specification. The specification
                       has a lot of options to make it simpler to
                       use. See the DATE/TIME SPECIFICATION section
                       for more details.

                       Here is an example that looks for files that
                       have not been modified in the last week:
                           $ %[1]v -o 1w

    -p REGEXP, --prune REGEXP
                       Prune a directory if the path matches the
                       regular expression. By default all directories
                       are searched.

                       The prune option can be used to significantly
                       speed up analysis. It is typically used to
                       ignore git repositories or directories that
                       contain generated files like lib, bin or
                       tmp.

                       Here is an example:
                           $ %[1]v -p '\.git$|^lib$|^bin$|^tmp$'

                       Multi-level directories can be specified as
                       well. Here is an example of that:
                           $ %[1]v -p 'project1/lib|project1/bin|project1/tools'

    -r REGEXP, --reject REGEXP
                       Reject if the contents match the regular expression.
                       If multiple reject criterion are specified,
                       only one of them has to match (an OR operation).

                       Here is an example that will reject a file
                       if it contains either foo or bar:
                           $ %[1]v -r foo -r bar .
                           test/nofoobar

    -R REGEXP, --Reject REGEXP
                       Reject if the contents match the regular expression.
                       If multiple reject criterion are specified,
                       all of them have to match (an AND operation).

                       Here is an example that will only reject a
                       file if it contains both foo and bar:
                           $ %[1]v -R foo -R bar .
                           test/fooonly
                           test/baronly
                           test/nofoobar

    -s, --summary      Print the summary report.

    -S INIT MAX --scan-buf-params INIT MAX
                       Set the internal scan buffer parameters to
                       handle long lines like those in some log
                       files. The defaults are 1048576 (1MB) and
                       10485760 (10MB). These values normally do
                       not need to be set.

    -v, --verbose      Increase the level of verbosity.
                       Can use -vv and -vvv as shorthand.

    -V, --version      Print the program version and exit.

    -W, --no-warning   Do not print warnings.

EXAMPLES
    # Example 1: help
    $ %[1]v -h

    # Example 2: Search the current directory tree for C, java and python files
    #            that do not have a specific copyright notice.
    #            Note that we reject files that contain the valid copyright
    #            notice so that we can fix the ones that don't have it.
    $ %[1]v -s \
        -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include

    # Example 3: Same as the previous search but only look at files that have
    #            changed in the past 4 weeks.
    $ %[1]v -n 4w -s \
        -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include

    # Example 4: Find all source files that have main and reference a macro
    #            called FOOBAR.
    $ %[1]v -s -l \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include \
        -A '\bmain\b' -A '\bFOOBAR\b'

    # Example 5: Find which files use a constant called FOOBAR_SPAM.
    #            Ignore generated files.
    $ %[1]v -s -l \
       -a '\bFOOBAR_SPAM\b \
       -e '\.log$|\.tmp$|\.o$|\.py[co]'

    # Example 6: Find which files use a constant called FOOBAR_SPAM.
    #            Ignore generated files and prune generated directories.
    $ %[1]v -s -l \
       -a '\bFOOBAR_SPAM\b \
       -e '\.log$|\.tmp$|\.o$|\.py[co]' \
       -p '\.git$|^lib$|^bin$|^tmp$'

    # Example 7: Find which files use constants called foobar_spam and
    #            wombat_zoo in a case-insensitive manner. Note that the (?i)
    #            applies to all of the OR alternatives.
    #            Ignore generated files and prune generated directories.
    $ %[1]v -s -l \
       -a '(?i)\bfoobar_spam\b|\bwombat_zoo\b' \
       -e '\.log$|\.tmp$|\.o$|\.py[co]' \
       -p '\.git$|^lib$|^bin$|^tmp$'

    # Example 8: Colorize the matches.
    $ %[1]v -C -a 'class|struct' -W -l

    # Example 9: Show before and after context.
    $ %[1]v -C --before 1 --after 5 -a 'def foo.*$' -W -l

    # Example 10: Colorize, before and after using shorthand.
    $ %[1]v -CWlyza 1 5 'def foo.*$'

    # Example 11: Use another tool (find) to pre-select files.
    $ %[1]v -CWlyza 1 5 'def foo.*$' $(find . -type f -name '*.c' -o -name '*.h')

    # Example 12: Find all double and single quoted strings in C source files.
    #             It is not perfect, it will find quoted strings in
    #             long comments but it is a decent start.
    #             Note the use of the -d option to prune single
    #             line comments, pragmas and other artifacts. This
    #             allows lines like this to be ignored:
    #                // this contains "a quoted string"
    $ %[1]v -CWl -a '"([^\\"]|.)*"' -a "'([^\\']|.)*'" -d '^\s*/\*|^\s*\*|^\s*#|^\s*//|\*/' -i '\.[ch]$'

COPYRIGHT:
   Copyright (c) 2017 Joe Linoff, all rights reserved

LICENSE:
   MIT Open Source

PROJECT:
   https://github.com/jlinoff/%[1]v
`

func help() {
	base := filepath.Base(os.Args[0])
	ncpus := runtime.NumCPU()
	fmt.Printf(helpMsg, base, ncpus)
	os.Exit(0)
}
