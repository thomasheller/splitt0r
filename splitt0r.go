package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"path"
)

func main() {
	filename := *flag.String("file", "", "input filename")
	char := *flag.String("char", "=", "delimiter char")
	delimiterLen := *flag.Int("len", 5, "minimum number of delimiter chars")
	wikiMode := *flag.Bool("wiki", false, "detect titles with MediaWiki markup")
	doWrite := *flag.Bool("write", false, "actually write output files")
	doPrint := *flag.Bool("print", false, "just print titles")
	doStats := *flag.Bool("stats", false, "just print statistics")
	outputDir := *flag.String("outdir", "output", "output directory name")
	outputExt := *flag.String("outext", ".txt", "output files extension")

	flag.Parse()

	if delimiterLen <= 0 {
		log.Fatal("Error: delimiter length must be 1 or greater")
	}

	if len(char) != 1 {
		log.Fatal("Error: delimiter must be a single character")
	}

	useStdin := filename == ""

	delimiterChar := []rune(char)[0]

	if !doWrite && !doPrint && !doStats {
		doStats = true
	}

	dupesDir := path.Join(outputDir, "dupes")

	if doWrite {
		prepareOutputDirs(outputDir, dupesDir)
	}

	var scanner *bufio.Scanner

	if useStdin {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Error opening file %s:\n%s", filename, err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	writer := newFileWriter(&osFileSystem{}, doWrite, doPrint, outputDir, outputExt, dupesDir)
	p := newParser(delimiterChar, delimiterLen, wikiMode)

	err := p.parseFile(scanner, writer)

	if err != nil {
		if useStdin {
			log.Fatalf("Error reading from stdin: %s\n", err)
		} else {
			log.Fatalf("Error reading from file %s: %s\n", filename, err)
		}
	}

	if doStats {
		printStats(writer)
	}
}

func prepareOutputDirs(outputDir string, dupesDir string) {
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

func printStats(w *fileWriter) {
	var average int
	if w.ArticlesCount() > 0 {
		average = w.LinesCount() / w.ArticlesCount()
	}

	log.Printf("Number of files: %d\n", w.ArticlesCount())
	log.Printf("Number of lines: %d\n", w.LinesCount())
	log.Printf("Average numer of lines: %d\n", average)
	log.Printf("Number of titles that appeared more than once: %d\n", w.DupeTitlesCount())
	log.Printf("Number of duplicate files: %d\n", w.DupeFilesCount())
}
