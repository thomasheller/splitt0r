package main

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestSplitSingleFile(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nbar\n",
	}, 1, 2, 0, 0, fs, w)
}

func TestSplitTwoFiles(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"bar",
		"=====",
		"baz baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nbar\n",
		"output/baz.txt": "baz baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitCustomDelimiter(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('-', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"-----",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo\nbar\n",
		"output/baz.txt": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitIgnoresShortDelimiters(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"====",
		"bar",
		"=====",
		"baz",
		"====",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo\n====\nbar\n",
		"output/baz.txt": "baz\n====\nbaz\n",
	}, 2, 6, 0, 0, fs, w)
}

func TestSplitCustomDelimiterLength(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 10, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"=====",
		"bar",
		"==========",
		"baz",
		"=====",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo\n=====\nbar\n",
		"output/baz.txt": "baz\n=====\nbaz\n",
	}, 2, 6, 0, 0, fs, w)
}

func TestSplitCustomOutputDirectory(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "myout", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"=====",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"myout/foo.txt": "foo\nbar\n",
		"myout/baz.txt": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitCustomOutputDirectoryWorkingDir(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, ".", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"=====",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"foo.txt": "foo\nbar\n",
		"baz.txt": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitCustomOutputDirectoryWorkingEmpty(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"=====",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"foo.txt": "foo\nbar\n",
		"baz.txt": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitCustomOutputExtension(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".md", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"=====",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo.md": "foo\nbar\n",
		"output/baz.md": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitCustomOutputExtensionEmpty(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", "", "output/dupes")

	p.parseFile(sl([]string{
		"foo",
		"bar",
		"=====",
		"baz",
		"baz",
	}), w)

	expect(t, map[string]string{
		"output/foo": "foo\nbar\n",
		"output/baz": "baz\nbaz\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitWikiMode(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, true)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"''foo'' foo",
		"123",
		"=====",
		"'''bar''' bar",
		"456",
		"=====",
		"''''baz'''' baz",
		"789",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "''foo'' foo\n123\n",
		"output/bar.txt": "'''bar''' bar\n456\n",
		"output/baz.txt": "''''baz'''' baz\n789\n",
	}, 3, 6, 0, 0, fs, w)
}

func TestSplitDuplicates(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"123",
		"=====",
		"foo foo",
		"456",
		"=====",
		"bar bar",
		"789",
		"=====",
		"bar bar",
		"012",
		"=====",
		"bar bar",
		"345",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt":           "foo foo\n123\n",
		"output/dupes/foo (2).txt": "foo foo\n456\n",
		"output/bar.txt":           "bar bar\n789\n",
		"output/dupes/bar (2).txt": "bar bar\n012\n",
		"output/dupes/bar (3).txt": "bar bar\n345\n",
	}, 5, 10, 2, 3, fs, w)
}

func TestSplitIgnoresPreceedingDelimiters(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"======",
		"======",
		"foo foo",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nbar\n",
	}, 1, 2, 0, 0, fs, w)
}

func TestSplitIgnoresPreceedingEmptyLines(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"",
		"",
		"foo foo",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nbar\n",
	}, 1, 2, 0, 0, fs, w)
}

func TestSplitIgnoresPreceedingEmptyLinesInContent(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"foo",
		"======",
		"",
		"",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitIgnoresDoubleDelimiterLines(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"foo",
		"======",
		"======",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitIgnoresEmptyContent(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"foo",
		"======",
		"",
		"======",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitIgnoresEmptyContentWhitespace(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"foo",
		"======",
		"   ",
		"======",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 4, 0, 0, fs, w)
}

func TestSplitIgnoresTrailingEmptyLines(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"",
		"foo",
		"",
		"",
		"======",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\n\nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 5, 0, 0, fs, w)
}

func TestSplitIgnoresTrailingEmptyLinesWithWhitespace(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"   ",
		"foo",
		"   ",
		"======",
		"bar bar",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\n   \nfoo\n",
		"output/bar.txt": "bar bar\nbar\n",
	}, 2, 5, 0, 0, fs, w)
}

func TestSplitWhitespaceAroundDelimiters(t *testing.T) {
	fs := newMemoryFileSystem()
	p := newParser('=', 5, false)
	w := newFileWriter(fs, true, true, "output", ".txt", "output/dupes")

	p.parseFile(sl([]string{
		"foo foo",
		"foo",
		"======   ",
		"bar bar",
		"   =====",
		"bar",
	}), w)

	expect(t, map[string]string{
		"output/foo.txt": "foo foo\nfoo\n",
		"output/bar.txt": "bar bar\n   =====\nbar\n",
	}, 2, 5, 0, 0, fs, w)
}

// sl turns string slice into Scanner for testing convenience
func sl(lines []string) *bufio.Scanner {
	b := &bytes.Buffer{}
	for _, line := range lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	s := bufio.NewScanner(b)
	return s
}

// expect compares expected files with memory file system state
func expect(t *testing.T, expected map[string]string, expectedCount, expectedLines, expectedDupeTitles, expectedDupeFiles int, fs *memoryFileSystem, w *fileWriter) {
	actual := fs.Files()

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Test failed.\nExpected:\n%v\nGot:\n%v\n", expected, actual)
	}

	if expectedCount != w.ArticlesCount() {
		t.Fatalf("Expected articles count %d, got: %d\n", expectedCount, w.ArticlesCount())
	}

	if expectedLines != w.LinesCount() {
		t.Fatalf("Expected lines count %d, got: %d\n", expectedLines, w.LinesCount())
	}

	if expectedDupeTitles != w.DupeTitlesCount() {
		t.Fatalf("Expected dupes count %d, got: %d\n", expectedDupeTitles, w.DupeTitlesCount())
	}

	if expectedDupeFiles != w.DupeFilesCount() {
		t.Fatalf("Expected dupe files count %d, got: %d\n", expectedDupeFiles, w.DupeFilesCount())
	}
}

// memoryFileSystem provides a simulated file system for testing,
// without actualyl writing files to disk
type memoryFileSystem struct {
	files       map[string]string
	currentFile string
}

func newMemoryFileSystem() *memoryFileSystem {
	return &memoryFileSystem{files: make(map[string]string)}
}

func (fs *memoryFileSystem) Files() map[string]string {
	return fs.files
}

func (fs *memoryFileSystem) WriteOpen(filename string) error {
	fs.currentFile = filename
	fs.files[fs.currentFile] = ""
	return nil
}

func (fs *memoryFileSystem) Fprintln(line string) {
	fs.files[fs.currentFile] = fs.files[fs.currentFile] + line + "\n"
}

func (fs *memoryFileSystem) FlushClose() error {
	fs.currentFile = ""
	return nil
}
