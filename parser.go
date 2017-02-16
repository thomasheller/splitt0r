package main

import (
	"bufio"
	"log"
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

type parser struct {
	delimiterChar rune
	delimiterLen  int
	wikiMode      bool

	state      parserState
	title      string   // current title
	lines      []string // current content
	emptyLines int

	writer *fileWriter
}

func newParser(char rune, len int, wiki bool) *parser {
	return &parser{delimiterChar: char, delimiterLen: len, wikiMode: wiki}
}

func (p *parser) parseFile(scanner *bufio.Scanner, w *fileWriter) error {
	p.writer = w

	p.state = leadingEmpty
	p.title = ""
	p.lines = make([]string, 0)
	p.emptyLines = 0

	for scanner.Scan() {
		line := scanner.Text()

		switch p.state {
		case delimiter:
			p.parseDelimiter(line)
		case leadingEmpty:
			p.parseLeadingEmpty(line)
		case content:
			p.parseContent(line)
		case empty:
			p.parseEmpty(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// If input did not end with delimiter, pretend it did:

	switch p.state {
	case content:
		fallthrough
	case empty:
		p.write()
	}

	return nil
}

func (p *parser) parseDelimiter(line string) {
	if p.isEmpty(line) {
		p.state = leadingEmpty
	} else if p.isDelimiter(line) {
		// skip
	} else {
		p.state = content
		p.title = p.parseTitle(line)
		p.lines = append(p.lines, line)
	}
}
func (p *parser) parseLeadingEmpty(line string) {
	if p.isEmpty(line) {
		// skip
	} else if p.isDelimiter(line) {
		p.state = delimiter
		// discard empty lines:
		p.lines = make([]string, 0)
	} else {
		p.state = content
		p.title = p.parseTitle(line)
		p.lines = append(p.lines, line)
	}
}
func (p *parser) parseContent(line string) {
	if p.isEmpty(line) {
		p.state = empty
		p.emptyLines = 1
		p.lines = append(p.lines, line)
	} else if p.isDelimiter(line) {
		p.state = delimiter
		p.write()
		p.lines = make([]string, 0)
	} else {
		p.lines = append(p.lines, line)
	}
}
func (p *parser) parseEmpty(line string) {
	if p.isEmpty(line) {
		p.emptyLines++
		p.lines = append(p.lines, line)
	} else if p.isDelimiter(line) {
		p.state = delimiter
		p.write()
		p.lines = make([]string, 0)
		p.emptyLines = 0
	} else {
		p.state = content
		p.emptyLines = 0
		p.lines = append(p.lines, line)
	}
}

func (p *parser) isEmpty(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

func (p *parser) isDelimiter(line string) bool {
	trimmed := strings.TrimRightFunc(line, unicode.IsSpace)

	if len(trimmed) < p.delimiterLen {
		return false
	}

	for _, char := range trimmed {
		if char != p.delimiterChar {
			return false
		}
	}

	return true
}

func (p *parser) parseTitle(line string) string {
	if p.wikiMode {
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

		log.Panicf("Error: No title with MediaWiki markup found in line: %s\n", line)
		return ""
	}

	firstWord := strings.Fields(line)[0]
	return firstWord
}

func (p *parser) write() {
	p.writer.WriteFile(p.title, p.lines, p.emptyLines)
}
