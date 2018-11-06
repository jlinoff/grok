// CLI options
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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
	After              int              // -x after
	Before             int              // -y before
	Binary             bool             // -b
	BinarySize         int              // -B
	CmdLine            string
	Colorize           bool // --color
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
	PruneOrPatterns    []*regexp.Regexp // -p
	RejectAndPatterns  []*regexp.Regexp // -R
	RejectOrPatterns   []*regexp.Regexp // -r
	ScanBufInitSize    int              // -S, --scan-buf-params
	ScanBufMaxSize     int              // -S, --scan-buf-params
	Summary            bool             // -s
	Verbose            int              // -v
	Warnings           bool             // --no-warnings
}

func loadCliOptions() (opts cliOptions) {
	opts.Verbose = 0
	opts.MaxDepth = -1 // all files
	opts.BinarySize = 1024
	opts.CmdLine = cliCmdLine()
	opts.Warnings = true
	opts.MaxJobs = runtime.NumCPU()
	opts.ScanBufInitSize = 1024 * 1024
	opts.ScanBufMaxSize = 10 * opts.ScanBufInitSize

	// Used to detect nested conf files.
	confMap := map[string]string{}

	// Make a local copy of the arguments because we have to modify
	// it.
	cache := make([]string, 0)
	args := append([]string{}, os.Args...)
	for i := 1; i < len(args) || len(cache) != 0; i++ {
		var arg string
		if len(cache) > 0 {
			// There are still single arguments in the cache,
			// handle them first.
			i--            // backup for the next round.
			arg = cache[0] // populate arg from the cache
			cache = cache[1:]
		} else {
			arg = args[i]
			if len(arg) > 2 && arg[0] == '-' && arg[1] != '-' {
				// Update the cache.
				// This allows spefications like:
				//   -CWla 'pattern'
				for j := 2; j < len(arg); j++ {
					newArg := fmt.Sprintf("-%c", arg[j])
					cache = append(cache, newArg)
				}
				arg = arg[:2] // grab the first option
			}
		}
		switch arg {
		case "-a", "--accept":
			opts.AcceptOrPatterns = append(opts.AcceptOrPatterns, cliGetNextArgRegexp(&i, args))
		case "-A", "--Accept", "--ACCEPT":
			opts.AcceptAndPatterns = append(opts.AcceptAndPatterns, cliGetNextArgRegexp(&i, args))
		case "-z", "--after":
			opts.After = cliGetNextArgInt(&i, args)
		case "-y", "--before":
			opts.Before = cliGetNextArgInt(&i, args)
		case "-b", "--binary":
			opts.Binary = true
		case "-B", "--binary-size":
			opts.BinarySize = cliGetNextArgInt(&i, args)
		case "-c", "--conf":
			readOptsConfFile(&i, &args, confMap)
		case "-C", "--color", "--colorize":
			opts.Colorize = true
		case "-e", "--exclude":
			opts.ExcludeOrPatterns = append(opts.ExcludeOrPatterns, cliGetNextArgRegexp(&i, args))
		case "-E", "--Exclude", "--EXCLUDE":
			opts.ExcludeAndPatterns = append(opts.ExcludeAndPatterns, cliGetNextArgRegexp(&i, args))
		case "-h", "--help":
			help()
		case "-i", "--include":
			opts.IncludeOrPatterns = append(opts.IncludeOrPatterns, cliGetNextArgRegexp(&i, args))
		case "-I", "--Include", "--INCLUDE":
			opts.IncludeAndPatterns = append(opts.IncludeAndPatterns, cliGetNextArgRegexp(&i, args))
		case "-l", "--lines":
			opts.Lines = true
		case "-m", "--max-depth":
			opts.MaxDepth = cliGetNextArgInt(&i, args)
		case "-M", "--max-jobs":
			opts.MaxJobs = cliGetNextArgInt(&i, args)
			if opts.MaxJobs < 1 {
				opts.MaxJobs = 1
			}
		case "-n", "--newer-than":
			opts.NewerThanFlag = true
			opts.NewerThan = cliGetNextArgDatetime(&i, args)
		case "-o", "--olderthan-than":
			opts.OlderThanFlag = true
			opts.OlderThan = cliGetNextArgDatetime(&i, args)
		case "-p", "--prune":
			opts.PruneOrPatterns = append(opts.PruneOrPatterns, cliGetNextArgRegexp(&i, args))
		case "-r", "--reject":
			opts.RejectOrPatterns = append(opts.RejectOrPatterns, cliGetNextArgRegexp(&i, args))
		case "-R", "--Reject", "--REJECT":
			opts.RejectAndPatterns = append(opts.RejectAndPatterns, cliGetNextArgRegexp(&i, args))
		case "-s", "--summary":
			opts.Summary = true
		case "-S", "--scan-buf-params":
			opts.ScanBufInitSize = cliGetNextArgInt(&i, args)
			opts.ScanBufMaxSize = cliGetNextArgInt(&i, args)
		case "-v", "--verbose":
			opts.Verbose++
		case "-vv", "-vvv", "-vvvv":
			opts.Verbose += len(arg) - 1
		case "-V", "--version":
			fmt.Printf("%v version %v\n", filepath.Base(os.Args[0]), version)
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
func cliGetNextArgDatetime(i *int, args []string) time.Time {
	j := *i
	now := time.Now()
	arg := cliGetNextArg(i, args)

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
	fatal("invalid date format for %v: '%v'", args[j], arg)
	return now
}

// cliGetNextArgRegexp
func cliGetNextArgRegexp(i *int, args []string) *regexp.Regexp {
	j := *i
	arg := cliGetNextArg(i, args)
	re, err := regexp.Compile(arg)
	if err != nil {
		fatal("could not compile regexp for %v: %v", args[j], err)
	}
	return re
}

// cliGetNextArgInt
func cliGetNextArgInt(i *int, args []string) int {
	j := *i
	arg := cliGetNextArg(i, args)
	val, err := strconv.Atoi(arg)
	if err != nil {
		fatal("not an integer for %v: %v", args[j], arg)
	}
	return val
}

// cliGetNextArg gets the next command line argument.
func cliGetNextArg(i *int, args []string) string {
	j := *i
	*i++
	if *i >= len(args) {
		fatal("missing argument for option %v", args[j])
	}
	return args[*i]
}

// quoteString
func quote(arg string) (result string) {
	result = arg
	qs := []string{" ", "\t", "\"", "'", "\\", "$", "*", "+", "^"}
	for _, x := range qs {
		if strings.Index(arg, x) >= 0 {
			// Quote it			arg = strings.Replace(arg, "'", "\\'", -1)
			result = "'" + arg + "'"
			break
		}
	}
	return
}

// get the command line
func cliCmdLine() (cli string) {
	cli = os.Args[0]
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		cli += " "
		cli += quote(arg)
	}
	return
}

// getCanonicalPath
func getCanonicalPath(path string) string {
	s, e := filepath.EvalSymlinks(path)
	if e != nil {
		fatal("%v", e)
	}
	a, e := filepath.Abs(s)
	if e != nil {
		fatal("%v", e)
	}
	return a
}

// readOptsConfFile - reads the options configuration file and inserts them
// into the args array.
func readOptsConfFile(i *int, args *[]string, confMap map[string]string) {
	conf := cliGetNextArg(i, *args) // conf file path
	newargs := []string{}
	path := getCanonicalPath(conf)
	if _, found := confMap[path]; found {
		// It was found! This could be an infinite recursive descent.
		fatal("nested reference to file '%v' found in conf file '%v'", conf, confMap[path])
	}
	confMap[path] = conf
	ifp, err := os.Open(conf)
	if err != nil {
		fatal("conf file read failed %v: %v", conf, err)
	}
	defer ifp.Close()
	s := bufio.NewScanner(ifp)
	for s.Scan() {
		line := strings.Trim(s.Text(), " \t")
		if len(line) == 0 || line[0] == '#' {
			// skip blank lines and lines that start with #.
			continue
		}
		// Split on the first white space.
		// Keep the rest intact.
		// Examples --arg "foo bar  spam"
		// want: ["--arg", "foo bar  spam"]
		flds := strings.Fields(line) // ignore all white space
		opt := flds[0]               // this is the option
		newargs = append(newargs, opt)

		// Now get the rest of the line
		// This assumes that all options have, at most, a single argument.
		arg := line[len(opt):]
		arg = strings.Trim(arg, " \t")
		if len(arg) > 0 {
			if len(arg) > 1 && arg[0] == arg[len(arg)-1] {
				if arg[0] == '"' || arg[0] == '\'' {
					arg = arg[1 : len(arg)-1] // trim the leading and trailing quotes
				}
			}
			newargs = append(newargs, arg)
		}
	}
	if err := s.Err(); err != nil {
		fatal("conf file read failed: %v: %v", conf, err)
	}

	// Update the slice by reference.
	if len(newargs) > 0 {
		*args = append((*args)[:*i+1], append(newargs, (*args)[*i+1:]...)...)
	}
}
