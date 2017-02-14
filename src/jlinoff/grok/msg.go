package main

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
)

// commaize - CITATION: http://stackoverflow.com/questions/13020308/how-to-fmt-printf-an-integer-with-thousands-comma
func commaize(n int64) string {
	in := strconv.FormatInt(n, 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

// getFunctionName gets the function name of the caller.
func getFunctionName(caller int) string {
	pc, _, _, _ := runtime.Caller(caller)
	fct := runtime.FuncForPC(pc).Name()
	return fct
}

func _msg(p string, f string, a ...interface{}) {
	_, _, lineno, _ := runtime.Caller(2)
	msg := fmt.Sprintf(f, a...)
	log.Printf("%-7s %5d - %v\n", p, lineno, msg)
}

func infov2(opts cliOptions, f string, a ...interface{}) {
	if opts.Verbose > 1 {
		_msg("INFO", f, a...)
	}
}

func infov3(opts cliOptions, f string, a ...interface{}) {
	if opts.Verbose > 2 {
		_msg("INFO", f, a...)
	}
}

func infov(opts cliOptions, f string, a ...interface{}) {
	if opts.Verbose > 0 {
		_msg("INFO", f, a...)
	}
}

func info(f string, a ...interface{}) {
	_msg("INFO", f, a...)
}

func debug(f string, a ...interface{}) {
	_msg("DEBUG", f, a...)
}

func warning(opts cliOptions, f string, a ...interface{}) {
	if opts.Warnings {
		_msg("WARNING", f, a...)
	}
}

func fatal(f string, a ...interface{}) {
	_, _, lineno, _ := runtime.Caller(1)
	msg := fmt.Sprintf(f, a...)
	log.Fatalf("FATAL   %5d - %v\n", lineno, msg)
}
