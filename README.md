# grok
[![Releases](https://img.shields.io/github/release/jlinoff/grok.svg?style=flat)](https://github.com/jlinoff/grok/releases)

Go program that searches directory trees for files that match regular expressions. 
It uses parallelism via goroutines to improve performance.

## Installation
Here is how you install it. You must have `go` and `make` in your path.

```bash
$ cd ~/work
$ git clone https://github.com/jlinoff/grok.git
$ cd grok
$ make
```

Now you can test it by going to a project directory and running it.

```bash
$ cd ~/projects/myproject
$ ~/work/grok/bin/grok -l -s -a `\bmain\b` .
```

For detailed information about the available options, use the `-h` option.

## Overview
I developed this tool to allow me to find symbols in files in directory trees
reasonably quickly which helps me grok the structure of the source code.
By default it searches in the current directory but you can explicitly specify
the directories or files that you want to search.

It is similar to doing a `find/grep` but the regular expressions
are more powerful, the file name appears before the file content and
multiple expressions can be search for simultaneously. Note that the search
capability is much less powerful than the capabilities provided by `find`.

You can specify whether to keep a file based on whether the file name
matches or does not match a set of regular expressions, whether the file is
older or newer than a date or whether the file content matches or does not
match regular expressions. You can also limit the search by directory depth.

These ideas are summarized in the following table. The term REs refers to
regular expressions.

| #   | Test     | Target    | Options  | Action |
| --: | -------- | --------- | -------- | ------ |
| 1   | accept   | contents  | -a, -A   | Accept a file if the contents match REs. |
| 2   | reject   | contents  | -r, -R   | Reject a file if the contents match REs. |
| 3   | include  | name      | -i, -I   | Include a file if the file name matches REs. |
| 4   | exclude  | name      | -e, -E   | Exclude a file if the file name matches REs. |
| 5   | newer    | date/time | -n       | Accept a file if it is newer than a date/time. |
| 6   | older    | date/time | -o       | Accept a file if it is older than a date/time. |
| 7   | maxdepth | depth     | -m       | Exclude files deeper than the depth. |
| 8   | prune    | name      | -p       | Exclude a directory if the path matched REs. |

You can specify whether a file must match all criteria (AND) or any criteria
(OR). In the table above, you can see that with the options that are lower and upper case.

A simple example should make all this a bit clearer. You want to search your
python, java and C source files to see which ones do not have a copyright
notice. The copy right notice has a very specific form:
    
    Copyright (c) YEARS by Acme Inc., all rights reserved

The YEARS is a list of years that is a comma or dash separated list of
years where each year is a 4 digit integer. This would be a valid list
of years: 2004-2015, 2017. Spaces are allowed.

The C files have .c and .h extensions. The java files have a .java extension
and the python files have a .py extension.

Here is the grok command you might use.

```bash
$ grok \
    -s \
    -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
    -i '\.[ch]$|\.java$|\.py$'
```

No directory was specified so the current directory tree will be searched to the bottom.

The `-s` tells the program to print the summary statistics.

The `-r` says to reject any file that contains the valid copyright notice so
that we only print the ones with the valid copyright notice.

The `-i` says to only include the files that have the specified extensions.

The regular expression syntax is the same used by go. It is described here:
https://github.com/google/re2/wiki/Syntax.

## Date/Time Specifications
The date/time specification is used by the `-n` and `-o` options to specify a
relative date time. A specification consists of a positive integer with a
suffix to indicate seconds, minutes, hours, days and weeks. To analyze all
files that have been modified in the last week you would specify `-n 1w` or
`-n 7d`.

 The table below lists the suffixes.
 
| Suffix | Duration | Example  |
| :----: | -------- | -------- |
| s      | seconds  | `-n 30s` |
| m      | minutes  | `-n 20m` |
| h      | hours    | `-n 12h` |
| d      | days     | `-n 3d`  |
| w      | weeks    | `-n 2w`  |
 
If no suffix is specified, seconds are assumed.

You can search time windows by using both options. Here is an example that
shows how to search files that are newer than 4 weeks but older than 2 weeks: `-n 4w -o 2w`.

This is very useful when you only want to search a specific time window.

## Examples
This section shows a few more examples that will help you understand how to use the tool.
Note that for most general searches, you will primarily use the `-a` option to match contents
along with the `-e` option to skip log files, tmp files, etc.

### Example 1
Get help.

```bash
$ grok -h
```

### Example 2
Search the current directory tree for C, java and python files
that do not have a specific copyright notice.
Note that we reject files that contain the valid copyright
notice so that we can fix the ones that don't have it.

```bash
$ grok \
    -s \
    -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
    -i '\.[ch]$|\.java$|\.py$' \
    tool1/src tool1/include tool2/src tool2/include
```

Use `-s` to print a summary.

Use `-r` to reject files with the valid copyright notice.

Use `-i` to define the files to search.

### Example 3
Same as the previous search but only look at files that have
changed in the past 4 weeks.

```bash
$ grok \
   -n 4w \
   -s \
   -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
   -i '\.[ch]$|\.java$|\.py$' \
   tool1/src tool1/include tool2/src tool2/include
```

Use `-n` to specify that we only want to look at files newer than 4 weeks.

### Example 4
Find all source files that have main and reference a macro called FOOBAR.

```bash
$ grok \
    -s \
    -l \
    -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include \
    -A '\bmain\b' -A '\bFOOBAR\b'
```

Use `-l` to show the matching lines in the files.

Use `-A` to make sure that all of the patterns occur in the same file to designate a match.

### Example 5
Find which files use a constant called FOOBAT_SPAM. Ignore generated files.

```bash
$ grok \
   -s \
   -l \
   -a '\bFOOBAR_SPAM\b` \
   -e '\.log$|\.tmp$|\.o$|\.py[co]'
```

Use `-a` to specify the word to search for.

Use `-e` to exclude .log, .tmp, .o, .pyc and .pyo files.

## Epilogue
I hope that you find this tool as useful as I have.

Comments or suggestions for improving/fixing it are greatly appreciated.

