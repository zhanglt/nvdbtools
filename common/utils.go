package common

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	translator "github.com/Conight/go-googletrans"
	log "github.com/sirupsen/logrus"
)

func UpdateDescription(srcFile, targetFile, structType, proxy string, db *sql.DB) error {
	var i int
	fp, err := os.Open(srcFile)
	if err != nil {
		fmt.Println("打开文件错误：", srcFile, ":", err) //打开文件错误
		return err
	}
	defer fp.Close()
	file, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", targetFile, ":", err)
		return err
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	scanner := bufio.NewScanner(fp)
	buf := make([]byte, 0, bufio.MaxScanTokenSize*10) //根据自己的需要调整这个倍数
	scanner.Buffer(buf, cap(buf))
	// 配置google翻译引擎的proxy
	c := translator.Config{
		Proxy: proxy,
	}
	//用proxy配置来初始化翻译引擎实例
	t := translator.New(c)
	if structType == "apps" {

		for {
			if !scanner.Scan() {
				break //文件读完了,退出for
			}
			line := scanner.Text() //获取每一行
			var apps *Apps = new(Apps)
			err := json.Unmarshal(s2b(line), apps)
			if err != nil {
				i = i + 1
				log.Println("序列化失败======:", i, err)
			} else {
				apps.D = getDescribe(db, apps.Vn, apps.D, t)
			}
			data, err := json.Marshal(&apps)
			write.WriteString(fmt.Sprintf("%s\n", data))

			//log.Println("cvd id =======:", apps.Vn)
		}
		//Flush将缓存的文件真正写入到文件中
		err = write.Flush()

	} else {

		for {
			if !scanner.Scan() {
				break //文件读完了,退出for
			}
			line := scanner.Text() //获取每一行
			var structData *Centos = new(Centos)
			err := json.Unmarshal(s2b(line), structData)

			if err != nil {
				i = i + 1
				log.Println("序列化失败======:", i, err)
			} else {
				structData.D = getDescribe(db, structData.N, structData.D, t)
			}

			data, err := json.Marshal(&structData)

			write.WriteString(fmt.Sprintf("%s\n", data))
		}
	}

	//log.Println("cvd id =======:", structData.Vn)

	//Flush将缓存的文件真正写入到文件中
	err = write.Flush()
	return err

}

// 临时计数器
var tindex int

// 根据cve编号 搜索在cnnvd、translate表中cveid对应的中文说明。
// 如果都查不到调用google翻译翻译并写入translate数据表
func getDescribe(db *sql.DB, cveid string, srcDescribe string, t *translator.Translator) string {
	var vuln_descript, other_id_cve_id sql.NullString
	rows := db.QueryRow("SELECT vuln_descript, other_id_cve_id FROM cnnvd where other_id_cve_id=$1", cveid)
	err := rows.Scan(&vuln_descript, &other_id_cve_id)
	if err != nil {
		// 如果 cnnvd数据表中没有对应数据，则去translate表中查找
		rows = db.QueryRow("SELECT descript ,cve_id FROM translate  where cve_id=$1", cveid)
		err = rows.Scan(&vuln_descript, &other_id_cve_id)
		// translate数据表中不存在的记录，强行用google进行翻译，并将翻译结果写入translate表中
		if err != nil {
			ts := translate(srcDescribe, t)
			if ts == "" { //翻译出错的，直接返回原值。
				return srcDescribe
			}
			//翻译成功的写入translate数据库
			// 问题：如果每次都用相同大的sqltie数据库文件，则可以达到迭代增加翻译条目。
			// 但是如果sqlite数据库重置或者新建，则会出现大量的翻译条目，造成更新时间几个小时。
			// 后续增加一个单独的功能，通过并发方式调用翻译引擎，解决首次翻译造成的时间问题。
			tx, _ := db.Begin()
			tx.Exec("insert into translate (cve_id, descript ) values(?,?)", cveid, ts)
			tindex = tindex + 1
			log.Println("---:", tindex, "---", cveid)
			tx.Commit()
			//返回翻译后的description
			return ts
		}
		// 如果在translate表中查到对应数据，返回description
		return vuln_descript.String
	}
	//log.Println("查到中文说明：", cveid)

	return vuln_descript.String

}

// 英文翻译函数
func translate(text string, t *translator.Translator) string {
	result, err := t.Translate(text, "en", "zh")
	if err != nil {
		log.Println("翻译错误：", err,"------------:",text)
		//log.Println("文本：",text)
		return ""
	}
	//fmt.Println(result.Text)
	return result.Text
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func InitPath(path string) error {
	err := os.RemoveAll(path)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatalf("%s路径初始化错误:%s\n", path, err)
		return err
	}

	return nil
}
