// CLI options
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// program version
var version = "v0.1"

// datetime
type cliDatetime struct {
	Re    *regexp.Regexp
	Scale time.Duration
	Value time.Time
}

// command line options
type cliOptions struct {
	AcceptAndPatterns  []*regexp.Regexp // -A
	AcceptOrPatterns   []*regexp.Regexp // -a
	Binary             bool             // -b
	BinarySize         int              // -B
	CmdLine            string
	Dirs               []string
	ExcludeAndPatterns []*regexp.Regexp // -E
	ExcludeOrPatterns  []*regexp.Regexp // -i
	IncludeAndPatterns []*regexp.Regexp // -I
	IncludeOrPatterns  []*regexp.Regexp // -i
	Lines              bool             // -l
	MaxDepth           int              // -m
	MaxJobs            int              // -M
	NewerThan          time.Time        // -n
	NewerThanFlag      bool
	OlderThan          time.Time //-o
	OlderThanFlag      bool
	RejectAndPatterns  []*regexp.Regexp // -R
	RejectOrPatterns   []*regexp.Regexp // -r
	Summary            bool             // -s
	Verbose            int              // -v
	Warnings           bool             // --no-warnings
}

func loadCliOptions() (opts cliOptions) {
	opts.Verbose = 0
	opts.MaxDepth = -1 // all files
	opts.BinarySize = 512
	opts.CmdLine = cliCmdLine()
	opts.Warnings = true
	opts.MaxJobs = runtime.NumCPU()
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-a", "--accept":
			opts.AcceptOrPatterns = append(opts.AcceptOrPatterns, cliGetNextArgRegexp(&i, arg))
		case "-A", "--Accept", "--ACCEPT":
			opts.AcceptAndPatterns = append(opts.AcceptAndPatterns, cliGetNextArgRegexp(&i, arg))
		case "-b", "--binary":
			opts.Binary = true
		case "-B", "--binary-size":
			opts.BinarySize = cliGetNextArgInt(&i, arg)
		case "-e", "--exclude":
			opts.ExcludeOrPatterns = append(opts.ExcludeOrPatterns, cliGetNextArgRegexp(&i, arg))
		case "-E", "--Exclude", "--EXCLUDE":
			opts.ExcludeAndPatterns = append(opts.ExcludeAndPatterns, cliGetNextArgRegexp(&i, arg))
		case "-h", "--help":
			help()
		case "-i", "--include":
			opts.IncludeOrPatterns = append(opts.IncludeOrPatterns, cliGetNextArgRegexp(&i, arg))
		case "-I", "--Include", "--INCLUDE":
			opts.IncludeAndPatterns = append(opts.IncludeAndPatterns, cliGetNextArgRegexp(&i, arg))
		case "-l", "--lines":
			opts.Lines = true
		case "-m", "--max-depth":
			opts.MaxDepth = cliGetNextArgInt(&i, arg)
		case "-M", "--max-jobs":
			opts.MaxJobs = cliGetNextArgInt(&i, arg)
			if opts.MaxJobs < 1 {
				opts.MaxJobs = 1
			}
		case "-n", "--newer-than":
			opts.NewerThanFlag = true
			opts.NewerThan = cliGetNextArgDatetime(&i, arg)
		case "-o", "--olderthan-than":
			opts.OlderThanFlag = true
			opts.OlderThan = cliGetNextArgDatetime(&i, arg)
		case "-r", "--reject":
			opts.RejectOrPatterns = append(opts.RejectOrPatterns, cliGetNextArgRegexp(&i, arg))
		case "-R", "--Reject", "--REJECT":
			opts.RejectAndPatterns = append(opts.RejectAndPatterns, cliGetNextArgRegexp(&i, arg))
		case "-s", "--summary":
			opts.Summary = true
		case "-v", "--verbose":
			opts.Verbose++
		case "-vv", "-vvv", "-vvvv":
			opts.Verbose += len(arg) - 1
		case "-V", "--version":
			fmt.Printf("%v\n", version)
			os.Exit(0)
		case "-W", "--no-warnings":
			opts.Warnings = false
		default:
			// Everything that is not an option must be a valid directory or file.
			_, err := os.Stat(arg)
			if err == nil {
				opts.Dirs = append(opts.Dirs, arg)
			} else {
				// The directory/file does not exist.
				// If there is a leading "-", assume that the user specified an invalid
				// option.
				if strings.HasPrefix(arg, "-") {
					fatal("unrecognized option '%v'", arg)
				}
				fatal("%v", err)
			}
		}
	}

	if len(opts.Dirs) == 0 {
		opts.Dirs = append(opts.Dirs, ".")
	}
	return
}

// cliGetNextArgDatetime
func cliGetNextArgDatetime(i *int, opt string) time.Time {
	now := time.Now()
	arg := cliGetNextArg(i)

	// Map of valid arguments.
	// This could easily be simple static structure but since this is a rare
	// operation, this approach is okay.
	m := map[string]cliDatetime{}
	m["default"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)$"), Scale: time.Second}
	m["second"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)[s]$"), Scale: time.Second}
	m["minute"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)[m]$"), Scale: time.Minute}
	m["hour"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)[h]$"), Scale: time.Minute}
	m["day"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)[d]$"), Scale: time.Hour * time.Duration(24)}
	m["week"] = cliDatetime{Re: regexp.MustCompile("^([0-9]+)[w]$"), Scale: time.Hour * time.Duration(24) * time.Duration(7)}

	for _, rec := range m {
		result := rec.Re.FindStringSubmatch(arg)
		if len(result) > 1 {
			val, _ := strconv.Atoi(result[1])
			td := time.Duration(val) * rec.Scale
			stamp := now.Add(-td)
			return stamp
		}
	}

	// Not found.
	fatal("invalid date format for %v: '%v'", opt, arg)
	return now
}

// cliGetNextArgRegexp
func cliGetNextArgRegexp(i *int, opt string) *regexp.Regexp {
	arg := cliGetNextArg(i)
	re, err := regexp.Compile(arg)
	if err != nil {
		fatal("could not compile regexp for %v: %v", opt, err)
	}
	return re
}

// cliGetNextArgInt
func cliGetNextArgInt(i *int, opt string) int {
	arg := cliGetNextArg(i)
	val, err := strconv.Atoi(arg)
	if err != nil {
		fatal("not an integer for %v: %v", opt, arg)
	}
	return val
}

// cliGetNextArg gets the next command line argument.
func cliGetNextArg(i *int) string {
	opt := os.Args[*i]
	*i++
	if *i >= len(os.Args) {
		fatal("missing argument for option %v", opt)
	}
	return os.Args[*i]
}

// get the command line
func cliCmdLine() (cli string) {
	cli = os.Args[0]
	qs := []string{" ", "\t", "\"", "'", "\\", "$", "*", "+", "^"}
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		cli += " "
		q := false
		for _, x := range qs {
			if strings.Index(arg, x) >= 0 {
				q = true
				break
			}
		}
		if q {
			arg = strings.Replace(arg, "'", "\\'", -1)
			cli += "'" + arg + "'"
		} else {
			cli += arg
		}
	}
	return
}

func help() {
	msg := `
USAGE
    %[1]v [OPTIONS] [<DIRS>]

DESCRIPTION
    Go program that searches directory trees for files that match regular
    expressions. It uses parallelism via goroutines to improve performance.

    It was developed to allow me to find symbols in files in directory trees
    reasonably quickly which helps me grok the structure of the source code.
    By default it searches in the current directory but you can explicitly
    specify the directories or files that you want to search.

    It is similar to doing a find/grep but the regular expressions
    are more powerful, the file name appears before the file content and
    multiple expressions can be search for simultaneously. Note that the search
    capability is much less powerful than the capabilities provided by find.

    You can specify whether to keep a file based on whether the file name
    matches or does not match a set of regular expressions, whether the file is
    older or newer than a date or whether the file content matches or does not
    match regular expressions. You can also limit the search by directory depth.

    These ideas are summarized in the following table.

        Test        Target     Action
        ==========  =========  =============================================
        accept      contents   Accept a file if the contents match REs.
        reject      contents   Reject a file if the contents match REs.

        include     name       Include a file if the file name matches REs.
        exclude     name       Exclude a file if the file name matches REs.

        newer       date/time  Accept a file if it is newer than a date/time.
        older       date/time  Accept a file if it is older than a date/time.

        maxdepth    depth      Exclude files deeper than the depth.

    You can specify whether a file must match all criteria (AND) or any criteria
    (OR). In the table above, you can see that with the options that are lower
    and upper case.

    A simple example should make all this a bit clearer. You want to search your
    python, java and C source files to see which ones do not have a copyright
    notice. The copy right notice has a very specific form:

        "Copyright (c) YEARS by Acme Inc., all rights reserved"

    The YEARS is a list of years that is a comma or dash separated list of
    years where each year is a 4 digit integer. This would be a valid list
    of years: 2004-2015, 2017. Spaces are allowed.

    The C files have .c and .h extensions. The java files have a .java extension
    and the python files have a .py extension.

    Here is the %[1]v command you might use.

        $ %[1]v \
            -s \
            -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
            -i '\.[ch]$|\.java$|\.py$'

    No directory was specified so the current directory tree will be searched
    to the bottom.

    The -s tells the program to print the summary statistics.

    The -r says to reject any file that contains the valid copyright notice so
    that we only print the ones with the valid copyright notice.

    The -i says to only include the files that have the specified extensions.

    The regular expression syntax is the same used by go. It is described here:
    https://github.com/google/re2/wiki/Syntax.

DATE/TIME SPECIFICATION
    The date/time specification is used by the -n and -o options to specify a
    relative date time. A specification consists of a positive integer with a
    suffix to indicate seconds, minutes, hours, days and weeks. To analyze all
    files that have been modified in the last week you would specify -n 1w or
    -n 7d.

    The table below lists the suffixes.

        S  Duration  Example
        =  ========  =======
        s  seconds   -n 30s
        m  minutes   -n 20m
        h  hours     -n 12h
        d  days      -n 3d
        w  weeks     -n 2w

    If no suffix is specified, seconds are assumed.

    You can search time windows by using both options. Here is an example that
    shows how to search files that are newer than 4 weeks but older than 2 weeks:

        -n 4w -o 2w

    This is very useful when you only want to search a specific time window.

OPTIONS
    -a REGEXP, --accept REGEXP
                       Accept if the contents match the regular expression.
                       If multiple accept criterion are specified, only one of
                       them has to match (an OR operation).
                       Here is an example that will accept a file if it contains
                       either foo or bar:
                           $ %[1]v -a foo -a bar  # same as -a 'foo|bar'
                           test/fooonly
                           test/baronly
                           test/foobar

    -A REGEXP, --Accept REGEXP
                       Accept if the contents match the regular expression.
                       If multiple accept criterion are specified, all of them
                       have to match (an AND operation).
                       Here is an example that will only accept a file if it
                       contains both foo and bar:
                           $ %[1]v -A foo -A bar .
                           test/foobar

    -b, --binary       Search binary files.
                       By default binary files are skipped.

    -B INT, --binary-size INT
                       Number of bytes to read to determine whether this is a
                       binary file.
                       The default is 512.

    -e REGEXP, --exclude REGEXP
                       Exclude file if the name matches the regular expression.
                       If multiple exclude criterion are specified, only one of
                       them has to match (an OR operation).
                       Here is an example that will exclude a file if its name
                       contains either foo or bar:
                           $ %[1]v -i foo -i bar

    -E REGEXP, --Exclude REGEXP
                       Exclude file if the name matches the regular expression.
                       If multiple exclude criterion are specified, all of them
                       have to match (an AND operation).
                       Here is an example that will exclude a file if its name
                       contains both foo and bar:
                           $ %[1]v -i foo -i bar
                           test/fooonly
                           test/baronly

    -h, --help         On-line help.

    -i REGEXP, --include REGEXP
                       Include file if the name matches the regular expression.
                       If multiple include criterion are specified, only one of
                       them has to match (an OR operation).
                       Here is an example that will include a file if its name
                       contains either foo or bar:
                           $ %[1]v -i foo -i bar
                           test/fooonly
                           test/baronly
                           test/foobar
                           test/nofoobar

    -I REGEXP, --Include REGEXP
                       Include file if the name matches the regular expression.
                       If multiple include criterion are specified, all of them
                       have to match (an AND operation).
                       Here is an example that will include a file if its name
                       contains both foo and bar:
                           $ %[1]v -i foo -i bar
                           test/foobar
                           test/nofoobar

    -l, --lines        Show the lines that match.
                       If this is not specified, only the file names are shown.

    -m INT, --max-depth INT
                       The maximum depth in the directory tree.
                       The top level is 0.

    -M INT, --max-jobs INT
                       Maximum number of jobs (goroutines) to run in parallel.
                       Each job is a file analysis.
                       The default is %[2]v.

    -n DATE/TIME, --newer-than DATE/TIME
                       Only consider files that are newer than the date/time
                       specification. The specification has a lot of options to
                       make it simpler to use. See the DATE/TIME SPECIFICATION
                       section for more details.
                       Here is an example that looks for files that were
                       modified in the last day:
                           $ %[1]v -n 1d

    -o DATE/TIME, --older-than DATE/TIME
                       Only consider files that are older than the date/time
                       specification. The specification has a lot of options to
                       make it simpler to use. See the DATE/TIME SPECIFICATION
                       section for more details.
                       Here is an example that looks for files that have not
                       been modified in the last week:
                           $ %[1]v -n 1w

    -r REGEXP, --reject REGEXP
                       Reject if the contents match the regular expression.
                       If multiple reject criterion are specified, only one of
                       them has to match (an OR operation).
                       Here is an example that will reject a file if it contains
                       either foo or bar:
                           $ %[1]v -r foo -r bar .
                           test/nofoobar

    -R REGEXP, --Reject REGEXP
                       Reject if the contents match the regular expression.
                       If multiple reject criterion are specified, all of them
                       have to match (an AND operation).
                       Here is an example that will only reject a file if it
                       contains both foo and bar:
                           $ %[1]v -R foo -R bar .
                           test/fooonly
                           test/baronly
                           test/nofoobar

    -s, --summary      Print the summary report.

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
    $ %[1]v \
        -s \
        -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include

    # Example 3: Same as the previous search but only look at files that have
    #            changed in the past 4 weeks.
    $ %[1]v \
        -n 4w \
        -s \
        -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include

    # Example 4: Find all source files that have main and reference a macro
    #            called FOOBAR
    $ %[1]v \
        -s \
        -l \
        -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include \
        -A '\bmain\b' -A '\bFOOBAR\b'

COPYRIGHT:
   Copyright (c) 2017 Joe Linoff, all rights reserved

LICENSE:
   MIT Open Source

PROJECT:
   https://github.com/jlinoff/%[1]v
`
	base := filepath.Base(os.Args[0])
	ncpus := runtime.NumCPU()
	fmt.Printf(msg, base, ncpus)
	os.Exit(0)
}