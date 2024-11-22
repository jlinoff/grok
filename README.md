# grok
[![Releases](https://img.shields.io/github/release/jlinoff/grok.svg?style=flat)](https://github.com/jlinoff/grok/releases)

Grep-like tool that searches for files that match regular expressions using concurrency to improve performance.

Written in go.

## Installation
There are three different types of installation described here: download the executable for linux,
download the executable for Mac and build from source.

### Install via Go

```
go install github.com/jlinoff/grok/src/jlinoff/grok@latest
```

### Download the Executable on Linux
If you simply want the executable and are on a linux-64 platform, do this.

```bash
$ curl -L https://github.com/jlinoff/grok/releases/download/v0.8.3/grok-linux-amd64 --out grok
$ chmod a+x grok
$ ./grok -h
```

### Download the Executable on Mac
If you simply want the executable and are on a recent MacOSX platform, do this.

```bash
$ curl -L https://github.com/jlinoff/grok/releases/download/v0.8.3/grok-darwin-amd64 --out grok
$ chmod a+x grok
$ ./grok -h
```

### Download and Build from Source
Here is how to download and build from source. You must have `go` and `make` in your path.

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
| 1   | accept   | contents  | -a, -A   | Accept a file if any of the line contents match REs. |
| 2   | reject   | contents  | -r, -R   | Reject a file if any of the line contents match REs. |
| 3   | include  | name      | -i, -I   | Include a file if the file name matches REs. |
| 4   | exclude  | name      | -e, -E   | Exclude a file if the file name matches REs. |
| 5   | newer    | date/time | -n       | Accept a file if it is newer than a date/time. |
| 6   | older    | date/time | -o       | Accept a file if it is older than a date/time. |
| 7   | maxdepth | depth     | -m       | Exclude files deeper than the depth. |
| 8   | prune    | name      | -p       | Exclude a directory if the path matched REs. |
| 9   | delete   | contents  | -d, -D   | Delete an accepted line if contents match REs. |

You can specify whether a file must match all criteria (AND) or any criteria
(OR). In the table above, you can see that with the options that are lower and upper case.

You can also specify information to help you get some context information.

| Short Option | Long Option | Description |
| ------------ | ----------- | ----------- |
| -C           | --color     | Colorize the regular expression matches. |
| -y N         | --before N  | Print the N lines before the line that has a match. |
| -z N         | --after N   | Print the N lines after the line that has a match. |

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
that we only print the ones _without_ the valid copyright notice.

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
Prune directories that contain generated files or repository files.

```bash
$ grok -s \
    -p '\.git$|lib$|bin$|tmp$' \
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
Prune directories that contain generated files or repository files.

```bash
$ grok -n 4w -s \
   -p '\.git$|lib$|bin$|tmp$' \
   -r 'Copyright (c) ([0-9]{4}\s*[,-]\s*)*[0-9]{4} by Acme Inc., all rights reserved' \
   -i '\.[ch]$|\.java$|\.py$' \
   tool1/src tool1/include tool2/src tool2/include
```

Use `-n` to specify that we only want to look at files newer than 4 weeks.

### Example 4
Find all source files that have main and reference a macro called FOOBAR.
Prune directories that contain generated files or repository files.

```bash
$ grok -s -l \
    -p '\.git$|lib$|bin$|tmp$' \
    -i '\.[ch]$|\.java$|\.py$' tool1/src tool1/include tool2/src tool2/include \
    -A '\bmain\b' -A '\bFOOBAR\b'
```

Use `-l` to show the matching lines in the files.

Use `-A` to make sure that all of the patterns occur in the same file to designate a match.

### Example 5
Find which files use a constant called FOOBAR_SPAM. Ignore generated files.
Prune directories that contain generated files or repository files.

```bash
$ grok -s -l \
   -p '\.git$|lib$|bin$|tmp$' \
   -a '\bFOOBAR_SPAM\b` \
   -e '\.log$|\.tmp$|\.o$|\.py[co]'
```

Use `-a` to specify the word to search for.

Use `-e` to exclude .log, .tmp, .o, .pyc and .pyo files.

### Example 6
Search the current directory tree to find all references to packages in /opt.

1. Ignore references to /opt/mystuff because we know that those are valid.
2. Ignore generated files in lib, bin, tmp and log directories.
3. Ignore the .git repository.
4. Ignore log and tmp files.

```bash
$ grok -W -s -l \
    -p '^\.git$|lib$|bin$|tmp$|log$' \
    -e '\.log$|~$|\.tmp$' \
    -r '/opt/mystuff' \
    -a '/opt/' \
    > /tmp/opt.log 2>&1
$ # Now generate the report.
$ grep '^  ' /tmp/opt.log | sed -e 's@.*/opt/@/opt/@' | awk -F/ '{if (NF > 2) {printf("/%s/%s\n",$2,$3)}}' | grep -v '/opt/$' | sort -fu | cat -n
     1	/opt/boost
     2	/opt/gcc
     3  /opt/macports
     4  /opt/openssl
```

Note that the `/tmp/opt.log` file is not strictly necessary. You could simply pipe the
results in the grep command sequence but I normally use a log file to allow me to refine
the filtering.

### Example 7
Put common options in a conf file (v0.6.0 or later).

The conf file recognizes one option with an optional argument per line. Blank lines, white space only lines and lines starting
with `#` as the first non-whitespace character are ignored. Nest references to the same conf file are detected and will cause a fatal
error to avoid infinite recursion.

```
$ cat >my.conf <<EOF
# Common options
# Prune git, repo source code control subdirectories.
# Prune the bin directory that only has generated files.
-p '\.git$|\.repo$|bin$'

# Ignore warnings.
-W

# Always print the summary report.
-s
EOF
$ grok -c myconf.conf -a '\bFOO_BAR_SPAM\b' -l
```

### Example 8
Use some of the more interesting functions to colorize and add before/after context lines.
```bash
$ grok -CWlyza 5 4 'error:|fatal:' /foo/bar /spam/wombat/blatz
```

Note that `y` is the same as `--before` and `z` is the same `--after`. These options and `-C` were added in v0.7.0.

### Example 9
Find all of the singly and doubly quoted strings that are fully contained on a single line.
Ignore lines that start with a `#` or `//`.
Note the use of the `-d` option that was introduced in v0.9.
```bash
$ grok -CWl -a '"([^\\\"]|.)*"' -a "'([^\\']|.)*'" -d '^\s*#|^\s*//'
```

## Epilogue
I hope that you find this tool as useful as I have.

Comments or suggestions for improving/fixing it are greatly appreciated.
