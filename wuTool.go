package xt

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	CurrentDir  string
	LogFileName string
)

// Initialize modul
func Initialize() error {
	dir, err := os.Getwd()

	if err == nil {
		CurrentDir = dir
	}

	return err
}

// Fatal-Error
func Fatal(v ...interface{}) {
	stime := STime(time.Now())
	fmt.Printf("\n%s", stime)
	s := fmt.Sprint(v...)
	fmt.Print(s)
	fmt.Print("\n")

	_log(stime, s)

	os.Exit(1)
}

// LogFunction
func Log(v ...interface{}) {
	s := fmt.Sprint(v...)

	buf := []byte(s)
	if buf[0] == '\n' {
		fmt.Print("\n")
		s = s[1:len(s)]
	}

	stime := STime(time.Now())

	fmt.Print(stime)
	fmt.Print(s)
	fmt.Print("\n")
	_log(stime, s)
}

func _log(stime string, s string) {
	LogFileName = CurrentDir + "/" + FTime()[0:8] + ".log"
	txt := "\n" + stime + " " + s
	AppendFile(LogFileName, txt)
}

// Time asString for Log
func STime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// Time asString for FileName
func FTime() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// loadfiles string
func LoadFiles(path, match string) (files []string, err error) {

	d, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	dfiles, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, fInfo := range dfiles {
		if fInfo.Mode().IsRegular() {
			//			fmt.Println("file:", fInfo.Name())

			matched, err := filepath.Match(match, fInfo.Name())
			if err != nil {
				fmt.Println(err)
			}

			if matched {
				files = append(files, fInfo.Name())
			}
		}
	}

	sort.Strings(files)

	return files, nil
}

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

// Create Directory
func CreateDir(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		fmt.Println("CreateDir:", dirName)
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		fmt.Println(dirName, "already exist as a file!")
		return false
	}

	return false
}

// FileExists
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// Create File
func CreateFile(path string) (err error) {
	// check if file exists
	_, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err == nil {
			defer file.Close()
		}
	}

	return
}

// DeleteFile
func DeleteFile(path string) (err error) {
	// delete file
	err = os.Remove(path)
	return
}

// AppendFile
func AppendFile(path string, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		fmt.Println(err)
	}
}

// WriteFile
func WriteFile(path string, data string) (int, error) {
	DeleteFile(path)
	CreateFile(path)

	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var n int
	// Write some text line-by-line to file.
	n, err = file.WriteString(data)
	if err != nil {
		return 0, err
	}

	// Save file changes.
	err = file.Sync()
	if err != nil {
		return 0, err
	}

	return n, nil
}

// string-compare
func StrComp(a, b string) int {
	if a == b {
		return 0
	}

	if a < b {
		return -1
	}

	return 1
}

// wenn TeilString gefunden, den Rest liefern
func StrStr(fStr string, needle string) string {
	if needle == "" {
		return ""
	}
	idx := strings.Index(fStr, needle)
	if idx == -1 {
		return ""
	}
	return fStr[idx:]
}

// StrnIcmp
func StrnIcmp(s1, s2 string, le int) int {
	b1 := []byte(s1)
	b2 := []byte(s2)

	l1 := len(b1)
	l2 := len(b2)

	for i := 0; i < le && i < l1 && i < l2; i++ {
		if b1[i] != b2[i] {
			return -1
		}

		if (i + 1) == le {
			return 0
		}
	}

	return -1
}

// ISO8859_1 to UTF8
func ToUTF8(s string) string {

	iso8859_1_buf := []byte(s)

	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		if b == 0x80 {
			buf[i] = '€'
		} else {
			buf[i] = rune(b)
		}
	}
	return string(buf)
}

// UTF8 to ANSI
func ToAnsi(buf *[]byte) []byte {
	ansiBuf := make([]byte, len(*buf))

	a := 0
	for i := 0; i < len(*buf); i++ {
		switch (*buf)[i] {
		case 0xe2: // € = e2 82 ac
			i++
			if (*buf)[i] == 0x82 {
				i++
				if (*buf)[i] == 0xac {
					ansiBuf[a] = 0x80
					a++
				}
			}
		case 0xc2:
			i++
			ansiBuf[a] = (*buf)[i]
			a++
		case 0xc3:
			i++
			ansiBuf[a] = (*buf)[i] + 0x40
			a++
		default:
			ansiBuf[a] = (*buf)[i]
			a++
		}
	}

	return ansiBuf[:a]
}

// SHex
func SHex(buf *[]byte) string {

	out := ""

	for i := 0; i < len(*buf); i++ {
		c := fmt.Sprintf("%.2x ", (*buf)[i])
		out = out + c
	}

	return out
}
