package xt

// ----------------------------------------------------------------------------------
// webFile.go for Go's xt package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// WriteCounter #
type WriteCounter struct {
	Total uint64
}

// XFileInfo #
type XFileInfo struct {
	Size uint64
	Time time.Time
}

// DownloadFile #
type DownloadFile struct {
	FileName string
	Web      XFileInfo
	Loc      XFileInfo
	Changed  bool
	parent   *DownloadFiles
}

// DownloadFiles #
type DownloadFiles struct {
	url  string
	List []DownloadFile
}

// GetDownloadFiles #
func GetDownloadFiles(url string) (*DownloadFiles, error) {

	var downFiles DownloadFiles

	downFiles.url = url
	buf, err := urlDownloadListFile(downFiles.url + "/download.txt")
	if err != nil {
		return &downFiles, err
	}

	fList := strings.Split(buf, "\n")

	//gstock.exe.gz;4230524;2020-02-02 13:23:17
	//gstock.linux.gz;4363757;2020-02-02 13:23:15
	//gstock32.exe.gz;4105654;2020-02-02 13:23:18
	for _, line := range fList {
		items := strings.Split(line, ";")
		if len(items) > 2 {
			size, _ := strconv.Atoi(items[1])
			t, _ := time.Parse("2006-01-02 15:04:05", items[2])

			wInfo := XFileInfo{Size: uint64(size), Time: t}
			lInfo := XFileInfo{}

			file := DownloadFile{FileName: items[0], Web: wInfo, Loc: lInfo, parent: &downFiles}
			downFiles.List = append(downFiles.List, file)
		}
	}

	return &downFiles, nil
}

// Download #
func (f *DownloadFile) Download(toFile string) error {

	// Create the file with .tmp extension, so that we won't overwrite a
	// file until it's downloaded fully
	tmpFile := toFile + ".tmp"
	os.Remove(tmpFile)

	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(f.parent.url + "/" + f.FileName + ".gz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our bytes counter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Println()
	out.Close()

	// Rename the tmp file back to the original file
	time.Sleep(2 * time.Second)
	err = os.Rename(tmpFile, toFile)
	if err != nil {
		return err
	}

	return f.SetFileTime(toFile)
}

// SetFileTime #
func (f *DownloadFile) SetFileTime(toFile string) error {
	// setFileTime: change both atime and mtime to currenttime
	return os.Chtimes(toFile, f.Web.Time, f.Web.Time)
}

// GetFileInfo #
func (flist *DownloadFiles) GetFileInfo(FileName string) (*DownloadFile, error) {
	lowerFile := strings.ToLower(FileName)

	for _, f := range flist.List {
		wFile := strings.ToLower(f.FileName)

		if wFile == lowerFile {
			loc, _ := time.LoadLocation("UTC")

			st, err := os.Stat(f.FileName)
			if err != nil {
				f.Loc.Time = f.Web.Time
				f.Loc.Size = 0
			} else {
				f.Loc.Time = st.ModTime().In(loc)
				f.Loc.Size = uint64(st.Size())
			}

			//			dif := f.Loc.Time.Sub(f.Web.Time)
			f.Changed = (f.Web.Size != f.Loc.Size) || (f.Loc.Time != f.Web.Time)

			if Global.Debug > 0 {
				fmt.Printf("webFile: %d %v\n", f.Web.Size, f.Web.Time)
				fmt.Printf("locFile: %d %v\n", f.Loc.Size, f.Loc.Time)
			}

			return &f, nil
		}
	}

	return nil, nil
}

func urlDownloadListFile(url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// URLfileSize #
func URLfileSize(url string) (int, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	// Is our request ok?
	if resp.StatusCode != http.StatusOK {
		err := errors.New(resp.Status)
		return 0, err
	}

	// the Header "Content-Length" will let us know
	// the total file size to download
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	return size, nil
}

// Write #
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress #prints the progress of a file write
func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s\rDownloading... %s complete", strings.Repeat(" ", 50), ReadableBytes(wc.Total))
}
