package xt

// ----------------------------------------------------------------------------------
// xBuf.go for Go's xt package
// Copyright 2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Buffer #
type Buffer struct {
	buf []byte
	off int
}

const (
	maxInt      = int(^uint(0) >> 1)
	errTooLarge = "ecv.Buffer: too large"
)

// Bytes #
func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// String #
func (b *Buffer) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Size #
func (b *Buffer) Size() int { return len(b.buf) }

// Cap #
func (b *Buffer) Cap() int { return cap(b.buf) }

// Pos #
func (b *Buffer) Pos() int { return b.off }

// Rewind #
func (b *Buffer) Rewind() { b.off = 0 }

// Clear #
func (b *Buffer) Clear() {
	b.buf = b.buf[:0]
	b.off = 0
}

// WriteString #
func (b *Buffer) WriteString(s string) (n int, err error) {
	le := len(s)
	m, ok := b.tryGrowByReslice(le)
	if !ok {
		m = b.grow(le)
	}
	return copy(b.buf[m:], s), nil
}

// WriteLine #
func (b *Buffer) WriteLine(s string) (n int, err error) {
	return b.WriteString(s + string(byte(10)))
}

// ReadLine #
func (b *Buffer) ReadLine(line *string) (err error) {
	slice, err := b.readSlice(byte(10))
	if err != nil {
		return err
	}

	le := len(slice)
	if le > 0 && slice[le-1] == byte(10) {
		le = le - 1
	}

	if le > 0 && slice[le-1] == byte(13) {
		le = le - 1
	}

	*line = string(slice[:le])
	return nil
}

// WriteToFile #
func (b *Buffer) WriteToFile(sFile string, toAdd bool) error {

	var flag int = os.O_WRONLY | os.O_CREATE
	if toAdd {
		flag = flag | os.O_APPEND
	} else {
		flag = flag | os.O_TRUNC
	}

	f, err := os.OpenFile(sFile, flag, 0666)
	if err != nil {
		return err
	}

	defer f.Close()

	return b.writeFile(f)
}

func (b *Buffer) writeFile(f *os.File) error {

	b.off = 0
	if _, err := f.WriteString(b.String()); err != nil {
		fmt.Println(err)
	}

	// Save file changes.
	if err := f.Sync(); err != nil {
		return err
	}

	return nil
}

func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

func (b *Buffer) grow(n int) int {
	m := len(b.buf) - b.off

	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Clear()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= 64 {
		b.buf = make([]byte, n, 64)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(errTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

func makeSlice(n int) []byte {
	defer func() {
		if recover() != nil {
			panic(errTooLarge)
		}
	}()

	return make([]byte, n)
}

func (b *Buffer) readSlice(delim byte) (line []byte, err error) {
	i := bytes.IndexByte(b.buf[b.off:], delim)
	end := b.off + i + 1
	if i < 0 {
		end = len(b.buf)
		err = io.EOF
	}
	line = b.buf[b.off:end]
	b.off = end
	return line, err
}
