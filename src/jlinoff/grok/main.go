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
var version = "v0.3"

// Local reporting stats
type findStats struct {
	FilesTested  int64
	FilesMatched int64
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
	}

	infov(opts, "files tested:  %8s", commaize(fs.FilesTested))
	infov(opts, "files matched: %8s", commaize(fs.FilesMatched))
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

	// Create the AND tables for accept and reject.
	aa := []bool{}
	ra := []bool{}
	if len(opts.AcceptAndPatterns) > 0 {
		aa = make([]bool, len(opts.AcceptAndPatterns))
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
	fileAnyOrAccepted := len(opts.AcceptAndPatterns) > 0
	aa1 := false // accept any for an AND condition (partial match)

	for i, line := range lines {
		ra1, _ := checkAndConditions(line, opts.RejectAndPatterns, &ra)
		ro1 := checkOrConditions(line, opts.RejectOrPatterns)

		// Any rejection, aborts.
		if ra1 == true || ro1 == true {
			fileRejected = true
			break
		}

		// Now see if the line is accepted.
		// That can occur for a partial match on the AND conditions or
		// any match for the OR conditions.
		fileAllAndAccepted, aa1 = checkAndConditions(line, opts.AcceptAndPatterns, &aa)
		ao1 := checkOrConditions(line, opts.AcceptOrPatterns)
		if fileAnyOrAccepted == false {
			fileAnyOrAccepted = ao1
		}

		// Any partial matches are collected for later.
		if aa1 == true || ao1 == true {
			if len(line) > 0 {
				matchedLines = append(matchedLines, i)
			}
		}
	}

	matched := false
	if fileRejected == false && (fileAllAndAccepted == true || fileAnyOrAccepted) {
		mutex.Lock()
		matched = true
		fs.FilesMatched++
		fmt.Printf("%v\n", path)
		if opts.Lines {
			for _, i := range matchedLines {
				lineno := i + 1
				line := lines[i]
				fmt.Printf("%8d | %v", lineno, line)
				if line[len(line)-1] != '\n' {
					fmt.Println("")
				}
			}
		}
		mutex.Unlock()
	}

	infov2(opts, "read %v lines, %v bytes, matched=%v", len(lines), stat.Size(), matched)
	return
}

// Read the lines from the file.
func readLines(opts cliOptions, path string) (lines []string) {
	file, err := os.Open(path)
	if err != nil {
		warning(opts, "unable to open file: %v", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		lines = append(lines, line)
		if err != nil {
			break // EOF
		}
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
