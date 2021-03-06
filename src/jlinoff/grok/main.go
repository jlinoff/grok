package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// program version
var version = "v0.9.1"

// Local reporting stats
type findStats struct {
	FilesTested  int64
	FilesMatched int64
	LinesMatched int64
}

// Locking semaphore for printing.
var mutex *sync.Mutex

// Maximum running goroutines.
var maxgo chan bool

// Need to test golint.
func main() {
	opts := loadCliOptions()
	infov(opts, "version: %v %v", filepath.Base(os.Args[0]), version)
	infov(opts, "cmdline: %v", opts.CmdLine)

	// Setup concurrency.
	mutex = &sync.Mutex{}
	maxgo = make(chan bool, opts.MaxJobs)

	// Start work.
	fs := findStats{}
	for _, dir := range opts.Dirs {
		walk(opts, dir, &fs, 0)
	}

	// Wait for the jobs to finish.
	// We will not be able to assign to all of them until all of the goroutines
	// have finished.
	for i := 0; i < cap(maxgo); i++ {
		maxgo <- true
	}

	// Output the summary information.
	if opts.Summary {
		fmt.Println("")
		fmt.Printf("summary: files tested : %8s\n", commaize(fs.FilesTested))
		fmt.Printf("summary: files matched: %8s\n", commaize(fs.FilesMatched))
		fmt.Printf("summary: lines matched: %8s\n", commaize(fs.LinesMatched))
	}

	infov(opts, "files tested:  %8s", commaize(fs.FilesTested))
	infov(opts, "files matched: %8s", commaize(fs.FilesMatched))
	infov(opts, "lines matched: %8s", commaize(fs.LinesMatched))
	infov(opts, "done")
}

// walk the directory tree looking for files that match.
func walk(opts cliOptions, path string, fs *findStats, depth int) {
	infov2(opts, "checking: %v %v '%v'", depth, opts.MaxDepth, path)
	if opts.MaxDepth >= 0 && depth > opts.MaxDepth {
		return
	}

	// If this is a file, process it.
	// If it is a directory, look at all of the entries.
	stat, err := os.Stat(path)
	if err != nil {
		// Normally this is just a bad link.
		warning(opts, "%v", err)
		return
	}

	if stat.IsDir() {
		if pruneDir(opts, path) {
			infov2(opts, "pruning '%v'", path)
			return
		}

		entries, err := ioutil.ReadDir(path)
		if err != nil {
			warning(opts, "cannot read directory: '%v' - %v", path, err)
			return
		}

		for _, entry := range entries {
			newPath := filepath.Join(path, entry.Name())
			stat, err = os.Stat(newPath)
			if err != nil {
				// Normally this is just a bad link.
				warning(opts, "%v", err)
			} else {
				if stat.IsDir() {
					walk(opts, newPath, fs, depth+1)
				} else {
					checkFileParallel(opts, newPath, stat, fs)
				}
			}
		}
	} else {
		checkFileParallel(opts, path, stat, fs)
	}
}

// pruneDir returns true if the directory path should be pruned.
func pruneDir(opts cliOptions, path string) bool {
	if len(opts.PruneOrPatterns) > 0 {
		for _, p := range opts.PruneOrPatterns {
			if p.MatchString(path) {
				return true // match was found, prune it
			}
		}
	}
	return false // by default all directories are accepted
}

// checkFileParallel sets up the channel for parallel execution
func checkFileParallel(opts cliOptions, path string, stat os.FileInfo, fs *findStats) {
	fs.FilesTested++
	maxgo <- true // reserve the slot
	go func(opts cliOptions, path string, stat os.FileInfo, fs *findStats) {
		checkFile(opts, path, stat, fs)
		<-maxgo // give up the slot
	}(opts, path, stat, fs)
}

// checkFile checks to see whether this file matches.
func checkFile(opts cliOptions, path string, stat os.FileInfo, fs *findStats) {
	// This is a file that we need to check.
	infov2(opts, "checking file: %v", path)

	// See if the file is out of date.
	if validTimestamp(opts, path, stat) == false {
		infov2(opts, "rejecting file by timestamp: '%v'", path)
		return
	}

	// Test the include/exclude and/or patterns.
	if matchFileName(opts, path) == false {
		infov2(opts, "rejecting file by name: '%v'", path)
		return
	}

	// Check to see if this is binary file.
	if opts.Binary == false && isBinary(opts, path, stat) {
		infov2(opts, "rejecting binary file: '%v'", path)
		return
	}

	// Create the AND tables for accept, delete and reject.
	aa := []bool{}
	da := []bool{}
	ra := []bool{}
	if len(opts.AcceptAndPatterns) > 0 {
		aa = make([]bool, len(opts.AcceptAndPatterns))
	}
	if len(opts.DeleteAndPatterns) > 0 {
		da = make([]bool, len(opts.DeleteAndPatterns))
	}
	if len(opts.RejectAndPatterns) > 0 {
		ra = make([]bool, len(opts.RejectAndPatterns))
	}

	// Read the file, look for matching patterns on each line.
	// If a match is found, print the line.
	// The key here is recognized that the OR accept/reject are
	// stateless but that the AND accept/reject are not.
	// For the AND conditions we keep track of all of the unique
	// matches.
	lines := readLines(opts, path)
	matchedLines := []int{}
	fileRejected := false
	fileAllAndAccepted := false
	fileAllAndDeleted := false
	fileAnyOrAccepted := false
	var aa1 bool // accept any for an AND condition (partial match)
	var da1 bool // delete accept any for an AND condition (partial match)

	before := make([][]string, 0)
	after := make([][]string, 0)
	for i, line := range lines {
		infov3(opts, "line: %04d %v : %v", i+1, path, line)

		// Check reject patterns.
		ra1, _ := checkAndConditions(line, opts.RejectAndPatterns, &ra)
		ro1 := checkOrConditions(line, opts.RejectOrPatterns)

		infov3(opts, "   reject all : %v", ra1)
		infov3(opts, "   reject any : %v", ro1)

		// Any rejection, abort this file.
		if ra1 == true || ro1 == true {
			fileRejected = true
			break
		}

		// Check accept AND patterns.
		// That can occur for a partial match on the AND conditions or
		// any match for the OR conditions.
		fileAllAndAccepted, aa1 = checkAndConditions(line, opts.AcceptAndPatterns, &aa)
		ao1 := checkOrConditions(line, opts.AcceptOrPatterns)

		infov3(opts, "   accept all : %v", fileAllAndAccepted)
		infov3(opts, "   accept any : %v %v", ao1, aa1)

		// Delete accepted patterns if there are delete matches.
		if fileAllAndAccepted == true || aa1 == true || ao1 == true {
			fileAllAndDeleted, da1 = checkAndConditions(line, opts.DeleteAndPatterns, &da)
			do1 := checkOrConditions(line, opts.DeleteOrPatterns)
			infov3(opts, "   delete all : %v", fileAllAndDeleted)
			infov3(opts, "   delete any : %v %v", do1, da1)
			if fileAllAndDeleted == true || do1 == true {
				// The delete won out.
				fileAllAndAccepted = false
				aa1 = false
				ao1 = false
			}
		}

		// Set the accept any flag.
		if fileAnyOrAccepted == false {
			fileAnyOrAccepted = ao1
		}

		infov3(opts, "   fileAnyOrAccepted  : %v", fileAnyOrAccepted)
		infov3(opts, "   fileAllAndAccepted : %v", fileAllAndAccepted)

		// Any partial matches are collected for later.
		if aa1 == true || ao1 == true {
			if len(line) > 0 {
				matchedLines = append(matchedLines, i)

				// Before context.
				if opts.Before > 0 {
					var context []string
					for k := i - opts.Before; k < i && k < len(lines); k++ {
						if k >= 0 {
							context = append(context, lines[k])
						}
					}
					before = append(before, context)
				}

				// After context.
				if opts.After > 0 {
					var context []string
					for k := i + 1; k < (i+opts.After+1) && k < len(lines); k++ {
						if k >= 0 {
							context = append(context, lines[k])
						}
					}
					after = append(after, context)
				}

			}
		}
	}

	matched := false
	if fileRejected == false && (fileAllAndAccepted == true || fileAnyOrAccepted) {
		mutex.Lock()
		matched = true
		fs.FilesMatched++
		fs.LinesMatched += int64(len(matchedLines))
		if opts.Lines != RawLines {
			// Do not print the file name for raw lines.
			if opts.Colorize {
				fmt.Printf("\033[1m%v\033[0m\n", path)
			} else {
				fmt.Printf("%v\n", path)
			}
		}
		if opts.Lines != NoLines {
			for m, i := range matchedLines {
				lineno := i + 1
				line := lines[i]

				// Before
				if opts.Before > 0 {
					if opts.Colorize {
						fmt.Printf("%8s \033[38;5;245m|----------------------------------------------------------------\033[0m\n", "")
						for _, c := range before[m] {
							fmt.Printf("\033[38;5;245m%8s |-%v\033[0m", "", c)
							printNewline(c)
						}
					} else {
						fmt.Printf("%8s |----------------------------------------------------------------\n", "")
						for _, c := range before[m] {
							fmt.Printf("%8s |-%v", "", c)
							printNewline(c)
						}
					}
				}

				// Line.
				if opts.Colorize {
					if opts.Lines == DecoratedLines {
						fmt.Printf("\033[38;5;245m%8d | \033[0m%v", lineno, colorizeLine(opts, line))
					} else if opts.Lines == RawLines {
						fmt.Printf("%v", colorizeLine(opts, line))
					}
					printNewline(line)
				} else {
					if opts.Lines == DecoratedLines {
						fmt.Printf("%8d | %v", lineno, line)
					} else if opts.Lines == RawLines {
						fmt.Printf("%v", line)
					}
					printNewline(line)
				}

				// After.
				if opts.After > 0 {
					if opts.Colorize {
						for _, c := range after[m] {
							fmt.Printf("\033[38;5;245m%8s |+%v\033[0m", "", c)
							printNewline(c)
						}
						fmt.Printf("%8s \033[38;5;245m|++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\033[0m\n", "")
					} else {
						for _, c := range after[m] {
							fmt.Printf("%8s |+%v", "", c)
							printNewline(c)
						}
						fmt.Printf("%8s |++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n", "")
					}
				}
			}
		}
		mutex.Unlock()
	}

	infov2(opts, "read %v lines, %v bytes, matched=%v", len(lines), stat.Size(), matched)
	return
}

// printNewline prints a new line if it is needed.
func printNewline(line string) {
	if len(line) == 0 {
		fmt.Printf("\n")
	} else if line[len(line)-1] != '\n' {
		fmt.Printf("\n")
	}
}

// colorizeLine colorizes a line.
func colorizeLine(opts cliOptions, line string) string {
	// At this point we know that we have a match.
	// TODO: find the longest match.
	// For now, just grab the first match.
	// opts.AcceptAndPatterns
	// opts.AcceptOrPatterns
	if len(opts.AcceptOrPatterns) > 0 {
		for _, re := range opts.AcceptOrPatterns {
			line = re.ReplaceAllStringFunc(line, func(arg1 string) string {
				return "\033[31;1m" + arg1 + "\033[0m"
			})
		}
	}
	if len(opts.AcceptAndPatterns) > 0 {
		for _, re := range opts.AcceptAndPatterns {
			line = re.ReplaceAllStringFunc(line, func(arg1 string) string {
				return "\033[31;1m" + arg1 + "\033[0m"
			})
		}
	}
	return line
}

// Read the lines from the file.
func readLines(opts cliOptions, path string) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		warning(opts, "unable to open file: %v", err)
		return
	}
	defer file.Close()
	s := bufio.NewScanner(file)
	sbuf := make([]byte, opts.ScanBufInitSize)
	s.Buffer(sbuf, opts.ScanBufMaxSize)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		warning(opts, "scanner error: %v", err)
	}
	return
}

// checkOrCondition, return True if anything matches.
func checkOrConditions(data string, ps []*regexp.Regexp) (matchedAny bool) {
	if len(ps) > 0 {
		for _, p := range ps {
			if p.MatchString(data) {
				matchedAny = true
				break
			}
		}
	} else {
		matchedAny = false
	}
	return
}

// checkAndCondition, return true if everything matches
func checkAndConditions(data string, ps []*regexp.Regexp, array *[]bool) (matchedAll bool, matchedAny bool) {
	if len(ps) > 0 {
		n := 0
		for i, p := range ps {
			if (*array)[i] == false {
				if p.MatchString(data) {
					(*array)[i] = true // update
					matchedAny = true
					n++
				}
			} else {
				n++
			}
		}
		if n == len(*array) {
			matchedAll = true
		}
	} else {
		matchedAll = false
		matchedAny = false
	}
	return
}

// matchesFileName test whether a file name matches all of the related
// criteria.
func matchFileName(opts cliOptions, path string) (match bool) {
	var p *regexp.Regexp

	// Exclude rules have priority.
	// If a valid exclude is found, it overrides the include rules.
	match = false

	// Any of the exclude OR patterns must match to exclude this file.
	if len(opts.ExcludeOrPatterns) > 0 {
		for _, p := range opts.ExcludeOrPatterns {
			if p.MatchString(path) == true {
				return
			}
		}
	}

	// All of the exclude AND patterns must match to exclude this file.
	if len(opts.ExcludeAndPatterns) > 0 {
		all := true
		for _, p = range opts.ExcludeAndPatterns {
			if p.MatchString(path) == false {
				all = false
				break
			}
		}
		if all {
			// Any rejection short circuits the logic.
			return
		}
	}

	// At this point there are no explicit rejections
	// but we need to check for explicit includes.
	match = true

	// Any of the include OR patterns must match to include this file.
	if len(opts.IncludeOrPatterns) > 0 {
		for _, p = range opts.IncludeOrPatterns {
			if p.MatchString(path) == true {
				return
			}
		}
	}

	// All of the include AND patterns must match to include this file.
	if len(opts.IncludeAndPatterns) > 0 {
		all := true
		for _, p = range opts.IncludeAndPatterns {
			if p.MatchString(path) == false {
				all = false
				break
			}
		}
		if all {
			return
		}
	}

	// If we made it to this point, no patterns were matched so
	// we apply a heuristic.
	// If only include patterns were defined, then exclude this file.
	// If only exclude patterns were defined, then include this file.
	// If both were defined, reject the file.
	if len(opts.IncludeAndPatterns) > 0 || len(opts.IncludeOrPatterns) > 0 {
		// Include patterns specified, never match.
		match = false
	} else {
		// No include patterns, always match.
		match = true
	}
	return
}

// matches
func matches(data string, patterns []*regexp.Regexp, def bool) bool {
	if len(patterns) == 0 {
		return def
	}
	for _, p := range patterns {
		m := p.MatchString(data)
		//debug("checking matched=%v, pattern='%v', data='%v'", m, p, data)
		if m {
			return true
		}
	}
	return false
}

// matchedsAny checks to see whether the data matches any of
// the patterns. It returns the index of the match so that
// it can be used for an AND operation.
func matchesAny(data string, patterns []*regexp.Regexp) (indexes []int) {
	indexes = []int{}
	if len(patterns) == 0 {
		return
	}
	for i, p := range patterns {
		if p.MatchString(data) {
			indexes = append(indexes, i)
		}
	}
	return
}

// isBinary determines whether a file is binary.
// It is a bit of a hack.
func isBinary(opts cliOptions, path string, stat os.FileInfo) bool {
	file, err := os.Open(path)
	if err != nil {
		infov2(opts, "unable to open file: %v", err)
		return true // skip files that we cannot open
	}
	defer file.Close()

	// Read the first N bytes.
	// If the file is smaller than the specified binary test size, then use that
	// size instead.
	size := int64(opts.BinarySize)
	if size > stat.Size() {
		size = stat.Size()
	}
	var buf = make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		warning(opts, "binary test: %v - %v", path, err)
		return true // skip files with read errors
	}

	n := 0 // number of new lines
	for i, b := range buf {
		switch b {
		case 0:
			// The last byte (EOF) is always NULL.
			if int64(i) < stat.Size() {
				return true
			}
		case '\n':
			n++
		}
	}

	// The first N bytes must contain at least 1 newline and no NULLs.
	if n == 0 {
		return true
	}

	// Passed all of the binary tests, assume that it is a text file.
	return false
}

// validTimestamp returns true if the file is in range.
func validTimestamp(opts cliOptions, path string, stat os.FileInfo) bool {
	if opts.NewerThanFlag == true {
		f := stat.ModTime().After(opts.NewerThan) || stat.ModTime().Equal(opts.NewerThan)
		// Don't exit if it is true, there may be an older-than check.
		if f == false {
			return f
		}
	}
	if opts.OlderThanFlag == true {
		return stat.ModTime().Before(opts.OlderThan) || stat.ModTime().Equal(opts.OlderThan)
	}
	return true // no timestamp is a valid timestamp
}
