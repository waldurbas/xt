package xt

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

// XFile #
type XFile struct {
	FileName string
	Size     uint64
	Time     time.Time
}

// DownloadFile #
type DownloadFile struct {
	WebFile XFile
	LocFile XFile
	Changed bool
	parent  *DownloadFiles
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

			wfile := XFile{FileName: items[0], Size: uint64(size), Time: t}
			lfile := XFile{}

			file := DownloadFile{WebFile: wfile, LocFile: lfile, parent: &downFiles}
			downFiles.List = append(downFiles.List, file)
		}
	}

	return &downFiles, nil
}

// Download #
func (f *DownloadFile) Download(toFile string) error {

	// Create the file with .tmp extension, so that we won't overwrite a
	// file until it's downloaded fully
	out, err := os.Create(toFile + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(f.parent.url + "/" + f.WebFile.FileName)
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

	// Rename the tmp file back to the original file
	err = os.Rename(toFile+".tmp", toFile)
	if err != nil {
		return err
	}

	// setFileTime: change both atime and mtime to currenttime
	return os.Chtimes(toFile, f.WebFile.Time, f.WebFile.Time)
}

// GetLocalFileInfo #
func (flist *DownloadFiles) GetLocalFileInfo(fname string) (*DownloadFile, error) {
	lowerFile := strings.ToLower(fname)

	for _, f := range flist.List {
		wFile := strings.ToLower(f.WebFile.FileName)

		if wFile == lowerFile {
			f.LocFile.FileName = f.WebFile.FileName

			st, err := os.Stat(f.LocFile.FileName)
			if err != nil {
				f.LocFile.Time = f.WebFile.Time
				f.LocFile.Size = 0
			} else {
				loc, _ := time.LoadLocation("UTC")
				f.LocFile.Time = st.ModTime().In(loc)
				f.LocFile.Size = uint64(st.Size())
			}

			f.Changed = f.WebFile.Size != f.LocFile.Size || (f.WebFile.Time != f.LocFile.Time)
			//			fmt.Printf("webFile: %d %v\n", f.WebFile.Size, f.WebFile.Time)
			//			fmt.Printf("locFile: %d %v\n", f.LocFile.Size, f.LocFile.Time)
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

// PrintProgress prints the progress of a file write
func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 50))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	//fmt.Printf("\rDownloading... %d complete", wc.Total)
	fmt.Printf("\rDownloading... %s complete", ReadableBytes(wc.Total))
}
