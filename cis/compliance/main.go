/*
**
翻译neuvector/share/scan/compliance.go 中的合规条目
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	translator "github.com/Conight/go-googletrans"
)

func t(text string, t *translator.Translator) string {

	result, err := t.Translate(text, "en", "zh")
	if err != nil {
		panic(err)
	}
	//fmt.Println(result.Text)
	return result.Text
}

func main() {
	c := translator.Config{
		Proxy: "http://127.0.0.1:10809",
	}
	ts := translator.New(c)
	ReadLines("./compliance.go.txt", "complianc_zh.go.txt", ts)
}
func ReadLines(inFile, outFile string, ts *translator.Translator) error {
	in, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer out.Close()

	write := bufio.NewWriter(out)
	read := bufio.NewScanner(in)
	for {
		if !read.Scan() {
			break //文件读完了,退出for
		}
		s := read.Text() //获取每一行
		l := len(s)
		if l > 17 {
			if strings.Contains(s[0:14], "Description:") {
				tmp := (t(s[16:l-2], ts))
				s = s[0:16] + tmp + s[l-2:l]
			}

			if strings.Contains(s[0:14], "Remediation:") {
				tmp := (t(s[16:l-2], ts))
				s = s[0:16] + tmp + s[l-2:l]

			}
		} else {
			s = s + "\n"
		}
		write.WriteString(fmt.Sprintf("%s\n", s))
	}

	err = write.Flush()
	return err

}
