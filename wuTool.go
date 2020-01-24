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

// GlobalData #
type GlobalData struct {
	CurrentDir   string
	Xargs        map[string]string
	xargsWithOut []string

	logDir      string
	logPfx      string
	logFileName string
}

// Global #
var Global GlobalData

// init: wird automatisch aufgerufen
func init() {
	Global.CurrentDir = "."
	dir, err := os.Getwd()

	if err == nil {
		Global.CurrentDir = dir
	}

	Global.logDir = Global.CurrentDir + "/log"
	Global.Xargs = make(map[string]string)

	var prev string
	for _, v := range os.Args[1:] {
		if v[0] == '-' {
			prev = strings.ToLower(v[1:2])
			if prev == "q" || prev == "x" {
				Global.Xargs[prev] = v[2:]
			} else {
				ix := strings.Index(v, "=")
				prev = ""
				if ix > 0 {
					prev = strings.ToLower(v[1:ix])
					Global.Xargs[prev] = v[ix+1:]
				} else {
					prev = strings.ToLower(v[1:])
					Global.Xargs[prev] = ""
				}
			}
		} else {
			Global.xargsWithOut = append(Global.xargsWithOut, v)
			if len(prev) > 0 {
				if len(Global.Xargs[prev]) == 0 {
					Global.Xargs[prev] = v
				}
			}
		}
	}
}

// SetLog #
func SetLog(logPfx string, logDir string) {
	if len(logDir) > 0 {
		Global.logDir = logDir
	}

	Global.logPfx = logPfx
}

// Param #
func Param(ix int, def string) string {
	if ix >= len(Global.xargsWithOut) {
		return def
	}

	return Global.xargsWithOut[ix]
}

// ParamValue #
func ParamValue(sKey string, def string) string {
	lKey, ok := ParamValueExist(sKey)
	if !ok {
		return def
	}

	return Global.Xargs[lKey]
}

// ParamKeyExist #
func ParamKeyExist(sKey string) bool {
	_, ok := ParamExist(sKey)
	return ok
}

// ParamExist #
func ParamExist(sKey string) (string, bool) {
	lKey := strings.ToLower(sKey)

	v, ok := Global.Xargs[lKey]
	//	fmt.Println("ParamExist.Key: ", uKey, ", ok: ", ok, ", v: ", v)

	return v, ok
}

// ParamValueExist #
func ParamValueExist(sKey string) (string, bool) {
	lKey := strings.ToLower(sKey)
	v, ok := Global.Xargs[lKey]
	return lKey, ok && len(v) > 0
}

// ParamAsInt #
func ParamAsInt(sKey string, def int) int {
	v, ok := ParamExist(sKey)
	if !ok || len(v) == 0 {
		return def
	}

	return Esubstr2int(v, 0, 10)
}

// ParamSet #
func ParamSet(sKey string, def string) {
	lKey := strings.ToLower(sKey)
	Global.Xargs[lKey] = def
}

// ParamValueCheck #
func ParamValueCheck(sKey string, def string) {
	v, ok := ParamExist(sKey)

	if ok && len(v) == 0 {
		ParamSet(sKey, def)
	}
}

// PrintParam #
func PrintParam() {
	fmt.Println("\n--> xParams:")
	for i, v := range Global.xargsWithOut {
		fmt.Printf("%d. [%s]\n", i, v)
	}

	fmt.Println("----------------------------")

	var sk []string
	for k := range Global.Xargs {
		sk = append(sk, k)
	}
	sort.Strings(sk)

	for _, k := range sk {
		fmt.Printf("%-16.16s: [%s]\n", k, Global.Xargs[k])
	}
	fmt.Print("\n\n")
}

// PermitWeekDay for
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
		case "so", "0":
			ok = ih == 0
		}
	}

	return ok
}

// PermitHour # array: [ "12:00-18:00","1400-2200"]
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

// Fatal #Error
func Fatal(v ...interface{}) {
	stime := STime(time.Now())
	fmt.Printf("\n%s", stime)
	s := fmt.Sprint(v...)
	fmt.Print(s)
	fmt.Print("\n")

	_log(stime, s)

	os.Exit(1)
}

// FatalF #Formatiert
func FatalF(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)

	_logx(s)

	os.Exit(1)
}

// LogF #Format-Function
func LogF(format string, v ...interface{}) (ss string) {
	s := fmt.Sprintf(format, v...)

	ss = _logx(s)
	return
}

// PrintStdErr #
func PrintStdErr(format string, v ...interface{}) (ss string) {
	ss = fmt.Sprintf(format, v...)
	fmt.Fprint(os.Stderr, ss)

	return
}

// Log #Function
func Log(v ...interface{}) {
	s := fmt.Sprint(v...)

	_logx(s)
}

func _logx(s string) (ss string) {
	buf := []byte(s)
	if buf[0] == '\n' {
		fmt.Fprint(os.Stderr, "\n")
		s = s[1:len(s)]
	}

	stime := STime(time.Now())
	if len(s) > 0 {
		fmt.Fprint(os.Stderr, stime)
		e := s[len(s)-1:]
		if e == "#" {
			s = s[:len(s)-1]
		}
		fmt.Fprint(os.Stderr, s)
		ss = stime + s

		if e != "#" {
			fmt.Fprint(os.Stderr, "\n")
		}
	}
	_log(stime, s)

	return
}

func _log(stime string, s string) {
	sti := FTime()[0:8]

	Global.logFileName = Global.logDir + "/" + sti[0:4] + "/" + sti[4:6]
	if !DirExists(Global.logFileName) {
		CreateDir(Global.logFileName)
	}

	Global.logFileName = Global.logFileName + "/" + Global.logPfx + sti + ".log"

	txt := "\n"
	if len(s) > 0 {
		txt = txt + stime + " " + s
	}
	AppendFile(Global.logFileName, txt)
}

// STime  #asString for Log
func STime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// FTime #asString for FileName
func FTime() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// TimeDif #
func TimeDif(tA time.Time, tL time.Time) (xs int, hh int, mm int, ss int) {
	dif := tL.Sub(tA)
	hh = int(dif.Hours())
	mm = int(dif.Minutes())
	ss = int(dif.Seconds())
	xs = ss

	if hh > 0 {
		mm -= hh * 60
		ss -= hh * 3600
	}

	if mm > 0 {
		ss -= mm * 60
	}

	return
}

// STimeDif #Differenz as String
func STimeDif(tA time.Time, tL time.Time) string {

	_, hh, mm, ss := TimeDif(tA, tL)
	s := fmt.Sprintf("%.2d:%.2d:%.2d", hh, mm, ss)
	return s
}

// LoadFiles #string
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

// CreateDir #
func CreateDir(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		fmt.Println("CreateDir:", dirName)

		err = nil
		sDirs := strings.Split(dirName, "/")
		cDir := ""
		for i := 1; err == nil && i < len(sDirs); i++ {
			cDir = cDir + "/" + sDirs[i]
			_, e := os.Stat(cDir)
			if e != nil {
				err = os.Mkdir(cDir, 0777)
				if err == nil {
					os.Chmod(cDir, 0777)
				}
			}
		}

		//		err = os.MkdirAll(dirName, 0777)
		if err != nil {
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

// DirExists #
func DirExists(path string) bool {
	f, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return f.IsDir()
}

// FileExists #
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

// CreateFile #
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

// DeleteFile #
func DeleteFile(path string) (err error) {
	// delete file
	err = os.Remove(path)
	return
}

// AppendFile #
func AppendFile(path string, data string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	os.Chmod(path, 0666)
	defer f.Close()

	if _, err := f.WriteString(data); err != nil {
		fmt.Println(err)
	}
}

// WriteFile #
func WriteFile(path string, data string) (int, error) {
	DeleteFile(path)
	CreateFile(path)

	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(path, os.O_RDWR, 0666)
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

// StrStr #wenn TeilString gefunden, den Rest liefern
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

// StrnIcmp #
func StrnIcmp(a, b string, le int) bool {
	l1 := len(a)
	l2 := len(b)

	if l1 < le || l2 < le {
		return false
	}

	return strings.EqualFold(a[0:le], b[0:le])
}

// StrIcmp #String Ignore Compare
func StrIcmp(a, b string) bool {
	return strings.EqualFold(a, b)
}

// StrComp #string-compare
func StrComp(a, b string) int {
	if a == b {
		return 0
	}

	if a < b {
		return -1
	}

	return 1
}

// Esubstr2int #
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
		} else if b[i] == ';' {
			break
		}
	}

	return z * f
}

// Esubstr #
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

// FormatInt64 #Format Integer mit Tausend Points
func FormatInt64(n int64) string {
	in := strconv.FormatInt(n, 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = '.'
		}
	}
}

// FormatInt #Format Integer mit Tausend Points
func FormatInt(n int) string {
	return FormatInt64(int64(n))
}

// ToUTF8 #ISO8859_1 to UTF8
func ToUTF8(s string) string {

	iso8859Buf := []byte(s)

	buf := make([]rune, len(iso8859Buf))
	for i, b := range iso8859Buf {
		if b == 0x80 {
			buf[i] = '€'
		} else {
			buf[i] = rune(b)
		}
	}
	return string(buf)
}

// ToAnsi #UTF8 to ANSI
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

// SHex #
func SHex(buf *[]byte) string {

	out := ""

	for i := 0; i < len(*buf); i++ {
		c := fmt.Sprintf("%.2x ", (*buf)[i])
		out = out + c
	}

	return out
}

// GetEnv #
func GetEnv(key, defval string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defval
}

// GetVersion #
func GetVersion(ss string) string {

	s := strings.Split(ss, ".")

	if len(s) != 4 {
		return "0.0.0.0"
	}

	var v [4]int

	for i := 0; i < 4; i++ {
		v[i] = Esubstr2int(s[i], 0, 4)
	}

	return strconv.Itoa(v[0]) + "." +
		strconv.Itoa(v[1]) + "." +
		strconv.Itoa(v[2]) + "." +
		strconv.Itoa(v[3])
}
