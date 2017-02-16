package main

import (
	"bufio"
	"fmt"
	"os"
)

type fileSystem interface {
	WriteOpen(filename string) error
	Fprintln(line string)
	FlushClose() error
}

// osFileSystem is a simple wrapper around the file system, so we can
// mock it out when testing.
type osFileSystem struct {
	file *os.File
	w    *bufio.Writer
}

func (fs *osFileSystem) WriteOpen(filename string) error {
	if fs.file != nil {
		panic("Can't open another file at the same time!")
	}

	var err error
	fs.file, err = os.Create(filename)

	if err != nil {
		return err
	}

	fs.w = bufio.NewWriter(fs.file)

	return nil
}

func (fs *osFileSystem) Fprintln(line string) {
	if fs.file == nil {
		panic("Can't write line before opening a file!")
	}

	fmt.Fprintln(fs.w, line)
}

func (fs *osFileSystem) FlushClose() error {
	if fs.file == nil {
		panic("Can't flush or close yet, no open file!")
	}

	defer func() {
		fs.file = nil
		fs.w = nil
	}()

	err := fs.w.Flush()
	if err != nil {
		return err
	}

	fs.file.Close()

	return nil
}
