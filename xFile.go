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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"os"
	"time"
)

// XFile #
type XFile struct {
	FileName string
	FileSize int64
	FileTime time.Time
	CheckSum string
	FileType string
	FileData []byte
}

// SFile #
type SFile struct {
	FileName string `json:"filename"`
	FileSize int64  `json:"filesize"`
	FileTime string `json:"filetime"`
	CheckSum string `json:"checksum"`
	FileType string `json:"filetype"`
	FileData string `json:"filedata"`
}

// XFile2SFile #
func XFile2SFile(xf *XFile) (sf SFile) {
	sf.FileName = xf.FileName
	sf.FileSize = xf.FileSize
	sf.FileTime = xf.FileTime.Format(time.RFC3339)
	sf.CheckSum = xf.CheckSum
	sf.FileType = xf.FileType
	return
}

// SFile2XFile #
func SFile2XFile(sf *SFile) (xf XFile) {
	xf.FileName = sf.FileName
	xf.FileSize = sf.FileSize
	xf.FileTime, _ = time.Parse(time.RFC3339, sf.FileTime)
	xf.CheckSum = sf.CheckSum
	xf.FileType = sf.FileType
	return
}

// LoadFile #
func LoadFile(sfile string) (*XFile, error) {
	stat, err := os.Stat(sfile)
	if os.IsNotExist(err) {
		return nil, err
	}

	if stat.IsDir() {
		return nil, errors.New("is not a file")
	}

	file, err := os.Open(sfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cf XFile

	loc, _ := time.LoadLocation("UTC")

	cf.FileName = sfile
	cf.FileSize = stat.Size()
	cf.FileTime = stat.ModTime().In(loc)
	cf.FileData = make([]byte, cf.FileSize)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(cf.FileData)

	if err != nil {
		return nil, err
	}

	chk := md5.Sum(cf.FileData)
	cf.CheckSum = hex.EncodeToString(chk[:16])
	return &cf, nil
}
