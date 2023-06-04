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
		tmp := getStr(s, "Description:", "Remediation:", ts)
		if tmp != "" {
			s = tmp
		}
		write.WriteString(fmt.Sprintf("%s\n", s))
	}

	err = write.Flush()
	return err

}
func getStr(str, substr1, substr2 string, ts *translator.Translator) string {

	if strings.Contains(str, substr1) {
		return substring(str, substr1, ts)
	}
	if strings.Contains(str, substr2) {
		return substring(str, substr2, ts)
	}

	return ""
}

func substring(str, substr string, ts *translator.Translator) string {
	var s string
	l := len(str)
	//获取字符串的位置
	i := strings.Index(str, substr)
	// 截取需要翻译的部分
	s = str[i+len(substr)+2 : l-1]
	//判断翻译部分是否为空或者None.
	if s != "" && s != "None." {
		s = t(s, ts)
	}
	//拼接字符串
	s = str[0:i+len(substr)+2] + s + str[l-1:l]
	return s
}
