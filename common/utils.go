package common

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

func UpdateDescription(srcFile, targetFile, structType string, db *sql.DB) error {
	var i int
	fp, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err) //打开文件错误
		return err
	}
	defer fp.Close()
	file, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
		return err
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)

	scanner := bufio.NewScanner(fp)
	buf := make([]byte, 0, bufio.MaxScanTokenSize*10) //根据自己的需要调整这个倍数
	scanner.Buffer(buf, cap(buf))
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
				apps.D = getDescribe(db, apps.Vn, apps.D)
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
				structData.D = getDescribe(db, structData.N, structData.D)
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

// 更具cve编号 搜索在cnvd中的中文说明
func getDescribe(db *sql.DB, cveid string, srcDescribe string) string {
	var vuln_descript, other_id_cve_id sql.NullString
	rows := db.QueryRow("SELECT vuln_descript, other_id_cve_id FROM cnnvd where other_id_cve_id=$1", cveid)
	err := rows.Scan(&vuln_descript, &other_id_cve_id)
	if err != nil {
		//log.Println("没有查到中文说明：", cveid)
		//log.Println("cve ID:", cveid, "数据库中没有搜索到:", err)
		return srcDescribe
	}
	//log.Println("查到中文说明：", cveid)
	return vuln_descript.String

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
