package main

import (
	"fmt"
	"log"
	"path"
)

type fileWriter struct {
	fileSystem fileSystem

	doWrite   bool
	doPrint   bool
	outputDir string
	outputExt string
	dupesDir  string

	articlesCount   int
	linesCount      int
	dupeTitlesCount int
	dupeFilesCount  int

	titles map[string]int
}

func newFileWriter(fs fileSystem, doWrite bool, doPrint bool, outputDir string, outputExt string, dupesDir string) *fileWriter {
	return &fileWriter{
		fileSystem: fs,
		doWrite:    doWrite,
		doPrint:    doPrint,
		outputDir:  outputDir,
		outputExt:  outputExt,
		dupesDir:   dupesDir,
		titles:     make(map[string]int),
	}
}

func (w *fileWriter) WriteFile(title string, content []string, emptyLines int) {
	w.articlesCount++

	count, exists := w.titles[title]

	if exists {
		w.dupeFilesCount++
		if count == 1 {
			w.dupeTitlesCount++
		}

		count++
	} else {
		count = 1
	}

	w.titles[title] = count

	lines := len(content) - emptyLines

	w.linesCount += lines

	if w.doWrite {
		var filename string
		if count == 1 {
			filename = path.Join(w.outputDir, title+w.outputExt)
		} else {
			filename = path.Join(w.dupesDir, fmt.Sprintf("%s (%d)%s", title, count, w.outputExt))
		}

		err := w.fileSystem.WriteOpen(filename)

		if err != nil {
			log.Fatalf("Error opening file %s for writing: %s\n", filename, err)
		}

		for idx, line := range content {
			if idx == lines {
				break
			}
			w.fileSystem.Fprintln(line)
		}

		err = w.fileSystem.FlushClose()
		if err != nil {
			log.Fatalf("Error writing to file %s: %s\n", filename, err)
		}
	}

	if w.doPrint {
		fmt.Printf("%s\n", title)
	}
}

func (w *fileWriter) ArticlesCount() int {
	return w.articlesCount
}

func (w *fileWriter) LinesCount() int {
	return w.linesCount
}

func (w *fileWriter) DupeTitlesCount() int {
	return w.dupeTitlesCount
}

func (w *fileWriter) DupeFilesCount() int {
	return w.dupeFilesCount
}
