# splitt0r

Split one file into multiple files based on delimiter

For example, take this input file:

```
=====
aaa bbb ccc
ddd eee fff
=====
ggg hhh iii
=====
jjj kkk lll
```

splitt0r will read this file and create three new files, `aaa.txt`, `ggg.txt` and `jjj.txt`, each containing the respective lines of text from the original file.

## Install

Prerequisites:
  - Git
  - Golang

If you haven't already, install Git and Golang on your system. On
Ubuntu/Debian this would be:

```
sudo apt-get install git golang
```

Then set up Go:
  - Create a directory for your `$GOPATH`, for example `~/gocode`
  - Set the `$GOPATH` environment variable accordingly: `export GOPATH=~/gocode`
  - Add the `bin` directory to your `$PATH`, for example: ` export PATH=$PATH:~/gocode/bin`

Now you can install splitt0r using `go get`:

```
go get github.com/thomasheller/splitt0r
```

## Usage

splitt0r supports three modes of operation:
  - `-stats` prints a few statistics about the input
  - `-print` prints all titles found in the input (see below for what's a title)
  - `-write` actually writes the split files into the output directory

If you don't pass any of the options, `-stats` is implied.

### Input

You can specify an input filename using `-file FILENAME`.
If you don't specify a filename, splitt0r will read from STDIN.

If you'd like splitt0r to recognize something different from `=====` as the delimiter,
specify the delimiter character using `-char CHAR`.
Note that `CHAR` must be exactly one character.

splitt0r assumes that the delimiter character appears at least 5 times.
If the input line is shorter, it is considered part of the content.
You can set this to any positive number (integer) using `-len NUMBER`.

### Output

By default, splitt0r will put all files in a subdirectory called `output`.
You can set this to something else using `-outdir DIRECTORY`.
The directory must be empty when splitt0r is started.
If the directory doesn't exist, splitt0r will create it for you.

All output filenames will be in the format `TITLE.txt` (regarding `TITLE`, see below).
You can change the filename extension using `-outext EXTENSION`.
Note that `EXTENSION` must include the leading dot (unless you don't want a dot), for example `.foo`.

### Titles (Filenames) and Duplicates

splitt0r will use the first word that appears after a delimiter line as the filename ("title") for the output (split) file.

There is a special mode called `-wiki` which parses the content according to MediaWiki markup rules.
In this case, the first word that is formatted either *italic* (`''italic''`), **bold** (`'''bold'''`) or ***bold-italic*** (`''''bold-italic''''`) is used as the filename ("title"), whichever comes first.

#### Duplicates

If splitt0r finds the same title more than once, it will proceed as follows:
  - The first file is put in the regular output directory, for example:  
  `output/TITLE.txt`
  - The second file is put in the `dupes` subdirectory and splitt0r appends an index, for example:  
  `output/dupes/TITLE (2).txt`
  - The third file would be:  
  `output/dupes/TITLE (3).txt`
  - and so forth...

This applies to `-write` mode. If you use `-print` to get a list of all titles, splitt0r will print the titles regardless of how often they appear in the input -- no indices will be appended. This is by design. In `-stats` mode, splitt0r will tell you how many duplicates appeared.

### Details

  - It doesn't matter how the input begins -- i.e., the first line doesn't need to be a delimiter line. splitt0r will ignore delimiter lines and empty lines until it finds the first line of actual content.
  - If there are two or more delimiter lines with no lines or just empty lines in between them, they will be ignored. splitt0r won't produce empty output files.
  - splitt0r will remove any empty lines that occur right before or right after the actual content. That means, you can have as many empty lines around your delimiter lines as you wish. If your content contains empty lines, they will be preserved.
  - If there is any whitespace (spaces, tabs...) *after* the delimiter characters, splitt0r will still consider the line a delimiter line. For example, `=====<SPACE><SPACE>` is a valid delimiter line. The opposite is not true: If there is any whitespace *before* the delimiter characters, the line is not recognized as a delimiter -- so `<SPACE><SPACE>=====` will be considered part of the content.
  - A line that consists solely of whitespace (spaces, tabs...) is considered empty. Nonetheless, splitt0r will preserve all whitespace characters in the output files: As noted above, it will remove empty lines that occur around your content and delimiter lines, but all empty lines inside your content will be copied verbatim, including potential whitespace characters.
  - A whitespace character is any character that Golang's [`unicode.IsSpace`](https://golang.org/pkg/unicode/#IsSpace) recognizes as a whitespace character.
