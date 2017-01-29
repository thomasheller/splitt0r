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

type State int

const (
	DELIMITER = iota
	LEADINGEMPTY
	CONTENT
	EMPTY
)

type Markup int

const (
	NONE = iota
	ITALIC
	BOLD
	BOLDITALIC
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

	titles          map[string]int = map[string]int{}
	articlesCount   int
	linesCount      int
	dupeTitlesCount int
	dupeFilesCount  int
)

func main() {
	filename := flag.String("file", "", "input filename")
	delimiter := flag.String("char", "=", "delimiter char")
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

	if len(*delimiter) != 1 {
		log.Fatal("Error: delimiter must be a single character")
	}

	delimiterChar = []rune(*delimiter)[0]

	if !doWrite && !doPrint && !doStats {
		doStats = true
	}

	dupesDir = path.Join(outputDir, "dupes")

	if doWrite {
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

	var scanner *bufio.Scanner

	if *filename == "" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(*filename)
		if err != nil {
			log.Fatalf("Error opening file %s:\n%s", *filename, err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	state := LEADINGEMPTY
	emptyLines := 0
	var title string
	content := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case DELIMITER:
			if isEmpty(line) {
				state = LEADINGEMPTY
			} else if isDelimiter(line) {
				// skip
			} else {
				state = CONTENT
				title = parseTitle(line)
				content = append(content, line)
			}
		case LEADINGEMPTY:
			if isEmpty(line) {
				// skip
			} else if isDelimiter(line) {
				state = DELIMITER
				// discard empty lines:
				content = make([]string, 0)
			} else {
				state = CONTENT
				title = parseTitle(line)
				content = append(content, line)
			}
		case CONTENT:
			if isEmpty(line) {
				state = EMPTY
				emptyLines = 1
				content = append(content, line)
			} else if isDelimiter(line) {
				state = DELIMITER
				writeFile(title, content, emptyLines)
				content = make([]string, 0)
			} else {
				content = append(content, line)
			}
		case EMPTY:
			if isEmpty(line) {
				emptyLines++
				content = append(content, line)
			} else if isDelimiter(line) {
				state = DELIMITER
				writeFile(title, content, emptyLines)
				content = make([]string, 0)
				emptyLines = 0
			} else {
				state = CONTENT
				emptyLines = 0
				content = append(content, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if *filename == "" {
			log.Fatalf("Error reading from stdin: %s\n", err)
		} else {
			log.Fatalf("Error reading from file %s: %s\n", *filename, err)
		}
	}

	switch state {
	case CONTENT:
		// pretend we hit a delimiter
		writeFile(title, content, emptyLines)
	case EMPTY:
		// pretend we hit a delimiter
		writeFile(title, content, emptyLines)
	}

	if doStats {
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
		var markup Markup

		if idxBoldItalic != nil {
			idxMin = idxBoldItalic[0]
			markup = BOLDITALIC
		}
		if idxBold != nil && (idxBold[0] < idxMin || idxMin == -1) {
			idxMin = idxBold[0]
			markup = BOLD
		}
		if idxItalic != nil && (idxItalic[0] < idxMin || idxMin == -1) {
			markup = ITALIC
		}

		switch markup {
		case BOLDITALIC:
			return rBoldItalic.FindStringSubmatch(line)[1]
		case BOLD:
			return rBold.FindStringSubmatch(line)[1]
		case ITALIC:
			return rItalic.FindStringSubmatch(line)[1]
		}

		log.Fatalf("Error: No title with MediaWiki markup found in line: %s\n", line)
		return ""
	} else {
		firstWord := strings.Fields(line)[0]
		return firstWord
	}
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
