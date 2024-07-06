package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// type declaration
/*
const
*/
const (
	singleReadNo  = 1
	readWriteFlag = 0
	readFlag      = 1
	writeFlag     = 2
	unnsfw        = false
	nsfw          = true
	layout        = "2006/01/02 15:04:05 MST"
	startTime     = " 00:00:00 JST"
	endTime      = " 23:59:59 JST"
)

//

/*
Log data structures
*/
type CONTENTS struct {
	Date    string `json:"date"`
	PubKey  string `json:"pubkey"`
	Content string `json:"content"`
}
type NOSTRLOG struct {
	Id       string
	Contents CONTENTS
}

//

/*
main {{{
*/
func main() {
	var (
		f = flag.String("f", "", "source file path name")
		d = flag.String("d", "2019/01/01", "date string : 2019/01/01")
	)
	flag.Parse()

	var cc confClass
	if err := cc.existConfiguration(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := cc.loadConfiguration(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var str string
	var err error
	var wb []NOSTRLOG
	if len(*f) < 1 {
		str, err = readStdIn()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if err := unmarchalStr(str, &wb); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	} else {
		if err := loadSourceFile(*f, &wb); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	start, end, err := parseDate(*d)
	dest := filterLogsByDateRange(wb, start, end)
	sort.Slice(dest, func(i, j int) bool {
		return dest[i].Contents.Date > dest[j].Contents.Date
	})
	fmt.Println("{")
	last := len(dest) - 1
	cnt := 0
	for i := range dest {
		if cnt < last {
			fmt.Printf("\t\"%v\" :{\n\t\t\"date\" : \"%v\",\n\t\t\"pubkey\" : \"%v\",\n\t\t\"content\" : \"%v\"\n\t},\n",
				dest[i].Id, dest[i].Contents.Date, dest[i].Contents.PubKey, dest[i].Contents.Content)
		} else {
			fmt.Printf("\t\"%v\" :{\n\t\t\"date\" : \"%v\",\n\t\t\"pubkey\" : \"%v\",\n\t\t\"content\" : \"%v\"\n\t}\n",
				dest[i].Id, dest[i].Contents.Date, dest[i].Contents.PubKey, dest[i].Contents.Content)
		}
		cnt++
	}
	fmt.Println("}")
}

// }}}

/*
parseDate {{{
*/
func parseDate(d string) (int64, int64, error) {
	var strStart string
	var strEnd string
	switch len(d) {
	case 4:
		strStart = fmt.Sprintf("%s/01/01%s", d, startTime)
		strEnd = fmt.Sprintf("%s/12/31%s", d, endTime)
	case 7:
		year := FirstFourChars(d)
		mon := getLastTwoChars(d)
		day, err := GetLastDayOfMonth(year, mon)
		if err != nil {
			return 0, 0, err
		}
		strStart = fmt.Sprintf("%s/%s/01%s", year, mon, startTime)
		strEnd = fmt.Sprintf("%s/%s/%s%s", year, mon, day, endTime)
	case 10:
		strStart = fmt.Sprintf("%s%s", d, startTime)
		strEnd = fmt.Sprintf("%s%s", d, endTime)
	default:
		return 0, 0, errors.New("Invalid date specification!")
	}
	layout := "2006/01/02 15:04:05 MST"
	sdate, err := time.Parse(layout, strStart)
	if err != nil {
		return 0, 0, err
	}
	edate, err := time.Parse(layout, strEnd)
	if err != nil {
		return 0, 0, err
	}
	return sdate.Unix(), edate.Unix(), nil
}

// }}}

/*
GetLastDayOfMonth {{{
*/
func GetLastDayOfMonth(year, month string) (string, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}

	dateStr := fmt.Sprintf("%s-%s-01", year, month)
	t, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return "", err
	}
	lastDay := t.AddDate(0, 1, -1).Day()

	return fmt.Sprintf("%02d", lastDay), nil
}

// }}}

/*
FirstFourChars {{{
*/
func FirstFourChars(s string) string {
	if len(s) < 4 {
		return s // 文字列の長さが4未満の場合、全ての文字列を返す
	}
	return s[:4]
}

// }}}

/*
getLastTwoChars {{{
*/
func getLastTwoChars(s string) string {
	if len(s) < 2 {
		return s // 文字列が2文字未満の場合はそのまま返す
	}
	runes := []rune(s) // 文字列をルーンスライスに変換
	return string(runes[len(runes)-2:]) // 最後の2文字を取得
}

// }}}

/*
filterLogsByDateRange {{{
*/
func filterLogsByDateRange(logs []NOSTRLOG, startUnix, endUnix int64) []NOSTRLOG {
	var filteredLogs []NOSTRLOG

	for _, log := range logs {
		// DateをUnix時間に変換
		unixTime, err := strconv.ParseInt(log.Contents.Date, 10, 64)
		if err != nil {
			continue
		}

		// 指定された範囲内にあるか確認
		if unixTime >= startUnix && unixTime <= endUnix {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs
}

// }}}

/*
readStdIn {{{
*/
func readStdIn() (string, error) {
	cn := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		var buff bytes.Buffer
		for sc.Scan() {
			fmt.Fprintln(&buff, sc.Text())
		}
		cn <- buff.String()
	}()
	timer := time.NewTimer(time.Second*5)
	defer timer.Stop()
	select {
	case text := <-cn:
		return text, nil
	case <-timer.C:
		return "", errors.New("Time out input from standard input")
	}
}

// }}}

/*
loadSourceFile {{{
*/
func loadSourceFile(path string, wb *[]NOSTRLOG) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var data interface{}
	dec := json5.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}
	b, err := json5.Marshal(data)
	if err != nil {
		return err
	}

	if err := unmarchalStr(string(b), wb); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return nil
}

// }}}

/*
unmarchalStr {{{
*/
func unmarchalStr(str string, wb *[]NOSTRLOG) error {
	p := make(map[string]CONTENTS)
	if err := json5.Unmarshal([]byte(str), &p); err != nil {
		return err
	}
	for i := range p {
		var content CONTENTS
		content.Date = p[i].Date
		content.PubKey = p[i].PubKey
		buf := p[i].Content
		buf = strings.Replace(buf, "\"", "\\\"", -1)
		buf = strings.Replace(buf, "\n", "\\n", -1)
		buf = strings.Replace(buf, "\b", "\\b", -1)
		buf = strings.Replace(buf, "\f", "\\f", -1)
		buf = strings.Replace(buf, "\r", "\\r", -1)
		buf = strings.Replace(buf, "\t", "\\t", -1)
		buf = strings.Replace(buf, "\\", "\\\\", -1)
		buf = strings.Replace(buf, "/", "\\/", -1)
		buf = strings.Replace(buf, "\"", "\\\"", -1)
		content.Content = buf
		tmp := NOSTRLOG{i, content}
		*wb = append(*wb, tmp)
	}
	return nil
}

// }}}

/*
debugPrint {{{
*/
func startDebug(s string) {
	f, err := os.OpenFile(s, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	log.Println("start debug")
}

// }}}
