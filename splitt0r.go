package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

type parserState int

const (
	delimiter = iota
	leadingEmpty
	content
	empty
)

type wikiMarkup int

const (
	none = iota
	italic
	bold
	boldItalic
)

var (
	delimiterChar rune
	delimiterLen  int
	wikimode      bool
	doWrite       bool
	doPrint       bool
	doStats       bool
	outputDir     string
	outputExt     string

	dupesDir string

	titles          = map[string]int{}
	articlesCount   int
	linesCount      int
	dupeTitlesCount int
	dupeFilesCount  int
)

func main() {
	filename := flag.String("file", "", "input filename")
	char := flag.String("char", "=", "delimiter char")
	flag.IntVar(&delimiterLen, "len", 5, "minimum number of delimiter chars")
	flag.BoolVar(&wikimode, "wiki", false, "detect titles with MediaWiki markup")
	flag.BoolVar(&doWrite, "write", false, "actually write output files")
	flag.BoolVar(&doPrint, "print", false, "just print titles")
	flag.BoolVar(&doStats, "stats", false, "just print statistics")
	flag.StringVar(&outputDir, "outdir", "output", "output directory name")
	flag.StringVar(&outputExt, "outext", ".txt", "output files extension")

	flag.Parse()

	if delimiterLen <= 0 {
		log.Fatal("Error: delimiter length must be 1 or greater")
	}

	if len(*char) != 1 {
		log.Fatal("Error: delimiter must be a single character")
	}

	useStdin := *filename == ""

	delimiterChar = []rune(*char)[0]

	if !doWrite && !doPrint && !doStats {
		doStats = true
	}

	dupesDir = path.Join(outputDir, "dupes")

	if doWrite {
		prepareOutputDirs()
	}

	var scanner *bufio.Scanner

	if useStdin {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(*filename)
		if err != nil {
			log.Fatalf("Error opening file %s:\n%s", *filename, err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	err := parseFile(scanner)

	if err != nil {
		if useStdin {
			log.Fatalf("Error reading from stdin: %s\n", err)
		} else {
			log.Fatalf("Error reading from file %s: %s\n", *filename, err)
		}
	}

	if doStats {
		printStats()
	}
}

func prepareOutputDirs() {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating output directory %s: %s\n", outputDir, err)
	}

	isEmpty, err := isDirEmpty(outputDir)
	if err != nil {
		log.Fatalf("Error while checking if output directory %s is empty: %s\n", outputDir, err)
	}
	if !isEmpty {
		log.Fatalf("Error: Please make sure the output directory %s is empty\n", outputDir)
	}

	err = os.MkdirAll(dupesDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating duplicates directory %s: %s\n", dupesDir, err)
	}
}

func isDirEmpty(name string) (bool, error) {
	file, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = file.Readdirnames(1)

	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func parseFile(scanner *bufio.Scanner) error {
	state := leadingEmpty
	emptyLines := 0
	var title string
	lines := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case delimiter:
			if isEmpty(line) {
				state = leadingEmpty
			} else if isDelimiter(line) {
				// skip
			} else {
				state = content
				title = parseTitle(line)
				lines = append(lines, line)
			}
		case leadingEmpty:
			if isEmpty(line) {
				// skip
			} else if isDelimiter(line) {
				state = delimiter
				// discard empty lines:
				lines = make([]string, 0)
			} else {
				state = content
				title = parseTitle(line)
				lines = append(lines, line)
			}
		case content:
			if isEmpty(line) {
				state = empty
				emptyLines = 1
				lines = append(lines, line)
			} else if isDelimiter(line) {
				state = delimiter
				writeFile(title, lines, emptyLines)
				lines = make([]string, 0)
			} else {
				lines = append(lines, line)
			}
		case empty:
			if isEmpty(line) {
				emptyLines++
				lines = append(lines, line)
			} else if isDelimiter(line) {
				state = delimiter
				writeFile(title, lines, emptyLines)
				lines = make([]string, 0)
				emptyLines = 0
			} else {
				state = content
				emptyLines = 0
				lines = append(lines, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// If input did not end with delimiter, pretend it did:

	switch state {
	case content:
		fallthrough
	case empty:
		writeFile(title, lines, emptyLines)
	}

	return nil
}

func isEmpty(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

func isDelimiter(line string) bool {
	trimmed := strings.TrimRightFunc(line, unicode.IsSpace)

	if len(trimmed) < delimiterLen {
		return false
	}

	for _, char := range trimmed {
		if char != delimiterChar {
			return false
		}
	}

	return true
}

func parseTitle(line string) string {
	if wikimode {
		rBoldItalic := regexp.MustCompile("''''(.+?)''''")
		rBold := regexp.MustCompile("'''(.+?)'''")
		rItalic := regexp.MustCompile("''(.+?)''")

		idxBoldItalic := rBoldItalic.FindStringIndex(line)
		idxBold := rBold.FindStringIndex(line)
		idxItalic := rItalic.FindStringIndex(line)

		idxMin := -1
		var m wikiMarkup

		if idxBoldItalic != nil {
			idxMin = idxBoldItalic[0]
			m = boldItalic
		}
		if idxBold != nil && (idxBold[0] < idxMin || idxMin == -1) {
			idxMin = idxBold[0]
			m = bold
		}
		if idxItalic != nil && (idxItalic[0] < idxMin || idxMin == -1) {
			m = italic
		}

		switch m {
		case boldItalic:
			return rBoldItalic.FindStringSubmatch(line)[1]
		case bold:
			return rBold.FindStringSubmatch(line)[1]
		case italic:
			return rItalic.FindStringSubmatch(line)[1]
		}

		log.Fatalf("Error: No title with MediaWiki markup found in line: %s\n", line)
		return ""
	}

	firstWord := strings.Fields(line)[0]
	return firstWord
}

func writeFile(title string, content []string, emptyLines int) {
	articlesCount++

	count, exists := titles[title]

	if exists {
		dupeFilesCount++
		if count == 1 {
			dupeTitlesCount++
		}

		count++
	} else {
		count = 1
	}

	titles[title] = count

	lines := len(content) - emptyLines

	linesCount += lines

	if doWrite {
		var filename string
		if count == 1 {
			filename = path.Join(outputDir, title+outputExt)
		} else {
			filename = path.Join(dupesDir, fmt.Sprintf("%s (%d)%s", title, count, outputExt))
		}

		file, err := os.Create(filename)

		if err != nil {
			log.Fatalf("Error opening file %s for writing: %s\n", filename, err)
		}
		defer file.Close()

		w := bufio.NewWriter(file)

		for idx, line := range content {
			if idx == lines {
				break
			}
			fmt.Fprintln(w, line)
		}

		err = w.Flush()
		if err != nil {
			log.Fatalf("Error writing to file %s: %s\n", filename, err)
		}
	}

	if doPrint {
		fmt.Printf("%s\n", title)
	}
}

func printStats() {
	var average int
	if articlesCount > 0 {
		average = linesCount / articlesCount
	}

	log.Printf("Number of files: %d\n", articlesCount)
	log.Printf("Number of lines: %d\n", linesCount)
	log.Printf("Average numer of lines: %d\n", average)
	log.Printf("Number of titles that appeared more than once: %d\n", dupeTitlesCount)
	log.Printf("Number of duplicate files: %d\n", dupeFilesCount)
}
