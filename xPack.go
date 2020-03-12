package xt

// ----------------------------------------------------------------------------------
// xPack.go for Go's xt package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Gzip string
func Gzip(data *[]byte) string {
	var b bytes.Buffer

	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(*data); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	str := base64.StdEncoding.EncodeToString(b.Bytes())

	return str
}

// Gunzip #
func Gunzip(data *string, dst *[]byte) error {
	src, err := base64.StdEncoding.DecodeString(*data)
	if err != nil {
		return err
	}

	gr, err := gzip.NewReader(bytes.NewBuffer(src))
	if err != nil {
		return err
	}

	defer gr.Close()

	*dst, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}

	return nil
}

// GzipFile #
func GzipFile(fileName string) (bool, error) {
	rawfile, err := os.Open(fileName)

	if err != nil {
		return false, err
	}
	defer rawfile.Close()

	// calculate the buffer size for rawfile
	info, _ := rawfile.Stat()

	var size int64 = info.Size()
	rawbytes := make([]byte, size)

	// read rawfile content into buffer
	buffer := bufio.NewReader(rawfile)
	_, err = buffer.Read(rawbytes)

	if err != nil {
		return false, err
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	writer.Write(rawbytes)
	writer.Close()

	err = ioutil.WriteFile(fileName+".gz", buf.Bytes(), info.Mode())

	if err != nil {
		return false, err
	}

	return true, nil

}

// GunzipFile #
func GunzipFile(fromFile string, toFile string) error {
	gzipfile, err := os.Open(fromFile)

	if err != nil {
		return err
	}

	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(toFile)

	if err != nil {
		return err
	}
	defer writer.Close()

	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	return nil
}
