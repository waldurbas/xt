package xt

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	CurrentDir   string
	LogFileName  string
	xargs        map[string]string
	xargsWithOut []string
)

// init: wird automatisch aufgerufen
func init() {
	CurrentDir = "."
	dir, err := os.Getwd()

	if err == nil {
		CurrentDir = dir
	}

	xargs = make(map[string]string)

	for _, v := range os.Args[1:] {
		ss := strings.Split(v, "=")
		if ss[0][0] == '-' {
			sx := strings.ToUpper(ss[0][1:])

			if len(ss) > 1 {
				xargs[sx] = ss[1]
			} else {
				xargs[sx] = "1"
			}
		} else {
			xargsWithOut = append(xargsWithOut, ss[0])
		}
	}
}

// Param
func Param(ix int, def string) string {
	if ix >= len(xargsWithOut) {
		return def
	}

	return xargsWithOut[ix]
}

// ParamExist
func ParamExist(sKey string) bool {
	uKey := strings.ToUpper(sKey)

	_, ok := xargs[uKey]
	return ok
}

// ParamValueExist
func ParamValueExist(sKey string) (string, bool) {
	uKey := strings.ToUpper(sKey)

	v, ok := xargs[uKey]
	return uKey, ok && len(v) > 0
}

// ParamValue
func ParamValue(sKey string, def string) string {
	uKey, ok := ParamValueExist(sKey)
	if !ok {
		xargs[uKey] = def
	}

	return xargs[uKey]
}

// ParamAsInt
func ParamAsInt(sKey string, def int) int {
	uKey, ok := ParamValueExist(sKey)
	if !ok {
		xargs[uKey] = strconv.Itoa(def)
	}

	return Esubstr2int(xargs[uKey], 0, 10)
}

// ParamAsBool
func ParamAsBool(sKey string, def bool) bool {
	uKey, ok := ParamValueExist(sKey)
	if !ok {
		ii := 0
		if def {
			ii = 1
		}

		xargs[uKey] = strconv.Itoa(ii)
	}

	return xargs[uKey] == "1"
}

// ParamSetDefault
func ParamSet(sKey string, def string) {
	uKey, _ := ParamValueExist(sKey)
	xargs[uKey] = def
}

// Printparam
func PrintParam() {
	fmt.Println("\n--> xParams:")
	for i, v := range xargsWithOut {
		fmt.Printf("%d. [%s]\n", i, v)
	}

	fmt.Println("----------------------------")

	var sk []string
	for k := range xargs {
		sk = append(sk, k)
	}
	sort.Strings(sk)

	for _, k := range sk {
		fmt.Printf("%-16.16s: [%s]\n", k, xargs[k])
	}
	fmt.Println("\n")
}

// PermitWeekDay
func PermitWeekDay(t time.Time, sDays []string) bool {
	ih := int(t.Weekday())
	ok := false
	for i := 0; i < len(sDays) && !ok; i++ {
		switch strings.ToLower(sDays[i]) {
		case "mo", "1":
			ok = ih == 1
		case "di", "2":
			ok = ih == 2
		case "mi", "3":
			ok = ih == 3
		case "do", "4":
			ok = ih == 4
		case "fr", "5":
			ok = ih == 5
		case "sa", "6":
			ok = ih == 6
		case "so", "7":
			ok = ih == 7
		}
	}

	return ok
}

// PermitHour: array: [ "12:00-18:00","1400-2200"]
func PermitHour(t time.Time, sh []string) bool {
	tt := t.Hour()*100 + t.Minute()
	ok := false

	var vt int
	var bt int
	for i := 0; i < len(sh) && !ok; i++ {
		s := strings.Split(sh[i], "-")

		if len(s) == 1 {
			vt = Esubstr2int(s[0], 0, 5)
			bt = 2400
		} else {
			vt = Esubstr2int(s[0], 0, 5)
			bt = Esubstr2int(s[1], 0, 5)
		}

		ok = tt >= vt && tt <= bt
	}

	return ok
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

	if len(s) > 0 {
		fmt.Print(stime)
		fmt.Print(s)
		fmt.Print("\n")
	}
	_log(stime, s)
}

func _log(stime string, s string) {
	sti := FTime()[0:8]

	LogFileName = CurrentDir + "/log/" + sti[0:4] + "/" + sti[4:6]
	if !DirExists(LogFileName) {
		CreateDir(LogFileName)
	}

	LogFileName = LogFileName + "/" + sti + ".log"

	txt := "\n"
	if len(s) > 0 {
		txt = txt + stime + " " + s
	}
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

// GzipFile
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

// DirExists
func DirExists(path string) bool {
	f, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return f.IsDir()
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
func StrnIcmp(a, b string, le int) bool {
	l1 := len(a)
	l2 := len(b)

	if l1 < le || l2 < le {
		return false
	}

	return strings.EqualFold(a[0:le], b[0:le])
}

// String Ignore Compare
func StrIcmp(a, b string) bool {
	return strings.EqualFold(a, b)
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

// Esubstr2int
func Esubstr2int(s string, ix int, le int) int {
	b := []byte(s[ix:])
	l := len(s)
	z := 0
	f := 1

	for i := 0; i < le && i < l; i++ {
		if b[i] >= '0' && b[i] <= '9' {
			z = z*10 + int(b[i]-'0')

		} else if b[i] == '-' {
			f = -1
		}
	}

	return z * f
}

// eSubStr
func Esubstr(s string, ix int, le int) string {
	l := len(s)

	if ix > l {
		return ""
	}

	if (ix + le) > l {
		le = l - ix
	}

	b := s[ix : ix+le]
	return b
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

// GetEnviron
func GetEnv(key, defval string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defval
}
