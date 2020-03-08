package xt

// ----------------------------------------------------------------------------------
// xFile.go for Go's xt package
// Copyright 2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"bufio"
	"os"
	"time"
)

// XFile #
type XFile struct {
	FileName string
	FileSize int64
	FileTime time.Time
	FileData []byte
}

// LoadFile #
func LoadFile(sfile string) (*XFile, error) {
	file, err := os.Open(sfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var cf XFile

	cf.FileName = sfile
	cf.FileSize = stat.Size()
	cf.FileTime = stat.ModTime()
	cf.FileData = make([]byte, cf.FileSize)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(cf.FileData)

	if err != nil {
		return nil, err
	}

	return &cf, nil
}
