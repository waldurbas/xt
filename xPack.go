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
	"io"
	"io/ioutil"
	"os"
)

// Gzip string
func Gzip(data *[]byte) (string, error) {
	var b bytes.Buffer

	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(*data); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// Gunzip #
func Gunzip(data *string) ([]byte, error) {

	dbuf, err := base64.StdEncoding.DecodeString(*data)
	if err != nil {
		return []byte{}, err
	}

	// wenn nicht gezippt ?
	if dbuf[0] != 0x1f || dbuf[1] != 0x8b || dbuf[2] != 0x08 {
		return dbuf, nil
	}

	gr, err := gzip.NewReader(bytes.NewBuffer(dbuf))
	if err != nil {
		return []byte{}, err
	}

	defer gr.Close()

	var bb []byte

	if bb, err = ioutil.ReadAll(gr); err != nil {
		return []byte{}, err
	}

	return bb, nil
}

// GzipBytes #
func GzipBytes(data *[]byte) []byte {
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

	return b.Bytes()
}

// GunzipBytes #
func GunzipBytes(src *[]byte, dst *[]byte) error {
	gr, err := gzip.NewReader(bytes.NewBuffer(*src))
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
