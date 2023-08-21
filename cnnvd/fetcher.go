/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cnnvd

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/beevik/etree"
)

type Rplist struct {
	Index int `json:"pageIndex"`
	Size  int `json:"pageSize"`
}
type Rpxml struct {
	ID       string `json:"id"`
	FileType int    `json:"downloadFileType"`
}
type cnnvdList struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Total   int `json:"total"`
		Records []struct {
			ID         string `json:"id"`
			TimeName   string `json:"timeName"`
			FileSize   string `json:"fileSize"`
			Version    int    `json:"version"`
			UpdateTime string `json:"updateTime"`
		} `json:"records"`
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
	} `json:"data"`
	Time string `json:"time"`
}

var index_id string = "CREATE INDEX cnnvd_other_id_cve_id_IDX ON cnnvd (other_id_cve_id)"
var index_desc string = "CREATE INDEX cnnvd_vuln_descript_IDX ON cnnvd (vuln_descript)"
var urlList string = "https://www.cnnvd.org.cn/web/vulDataDownload/getPageList"

// 获取cnnvd cveID 列表
func GetIDlist() ([]string, error) {

	var (
		err error
	)

	reqParam, err := json.Marshal(&Rplist{Index: 1, Size: 100})

	if err != nil {
		log.Fatalln("cnnvd数据文件id列表masrshal错误:%v", err)
		return nil, err
	}

	rb := strings.NewReader(string(reqParam))

	r, err := http.NewRequest("POST", urlList, rb)
	if err != nil {
		log.Fatalln("创建rquest错误 url: %s, reqBody: %s, err: %v", urlList, rb, err)
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// DO: HTTP请求
	rs, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("http.Do获取数据错误, url: %s, reqBody: %s, err:%v", urlList, rb, err)
		return nil, err

	}
	defer rs.Body.Close()
	// Read: HTTP结果
	rspBody, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		log.Println("ReadAll读取数据错误, url: %s, reqBody: %s, err: %v", urlList, rb, err)
		return nil, err

	}
	var result cnnvdList
	if err = json.Unmarshal(rspBody, &result); err != nil {
		log.Println("Unmarshal错误:%v", err)
		return nil, err

	}
	var idList []string
	for _, d := range result.Data.Records {
		idList = append(idList, d.ID)
	}

	return idList, nil
}

func GetXml(fid, token, savePath string) string {
	urlXml := "https://www.cnnvd.org.cn/web/vulDataDownload/download"
	//fid := "ac1d691cc42532556025bf9366ee6297"
	//token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VySW5mbyI6eyJ1c2VyTmFtZSI6ImtpdHNka0AxNjMuY29tIiwidXNlcklkIjoiNjFiYzE1OGIxYTMwYWVlMWZjZDQzMTg5YWVlY2MwODcifSwiYXVkIjoid2ViIiwiaXNzIjoiZ2Mtd2ViIiwiZXhwIjoxNjg0NDgzMTI3LCJpYXQiOjE2ODQ0Nzk1MjcsImp0aSI6IjNkYzJjYmJlMDFiNTQ4ZmU5MDEwNTI0MGNkYmNmNjhkIn0.-rjR9QFwBqgwtNiT7YfRweG2B2TSL6BZEeELO4_7uHE"
	rpxml, err := json.Marshal(&Rpxml{ID: fid, FileType: 1})

	if err != nil {
		log.Fatalln("Marshal 错误:%v", err)
		return ""
	}

	rb := strings.NewReader(string(rpxml))

	r, err := http.NewRequest("POST", urlXml, rb)
	if err != nil {
		log.Fatalln("创建rquest错误, url: %s, reqBody: %s, err: %v", urlXml, rb, err)
		return ""
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "application/json, text/plain, */*")
	r.Header.Add("Accept-Encoding", "gzip, deflate, br")
	r.Header.Add("token", token)

	// DO: HTTP请求
	rs, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("http.Do获取数据错误, url: %s, reqBody: %s, err:%v", urlXml, rb, err)
		return ""

	}
	defer rs.Body.Close()

	buf := make([]byte, 1024*1024*20)
	f, err := os.OpenFile(savePath+fid+".xml", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm) //可读写，追加的方式打开（或创建文件）
	if err != nil {
		log.Println("打开xml目标文件错误", err)
		return ""
	}
	defer f.Close()

	for {
		n, _ := rs.Body.Read(buf)
		if 0 == n {
			break
		}
		f.WriteString(string(buf[:n]))
	}

	return fid
}

type Cnnvd struct {
	XMLName         xml.Name `xml:"cnnvd"`
	Text            string   `xml:",chardata"`
	CnnvdXMLVersion string   `xml:"cnnvd_xml_version,attr"`
	PubDate         string   `xml:"pub_date,attr"`
	Xsi             string   `xml:"xsi,attr"`
	Entry           struct {
		Text         string `xml:",chardata"`
		Name         string `xml:"name"`
		VulnID       string `xml:"vuln-id"`
		Published    string `xml:"published"`
		Modified     string `xml:"modified"`
		Source       string `xml:"source"`
		Severity     string `xml:"severity"`
		VulnType     string `xml:"vuln-type"`
		VulnDescript string `xml:"vuln-descript"`
		OtherID      struct {
			Text      string `xml:",chardata"`
			CveID     string `xml:"cve-id"`
			BugtraqID string `xml:"bugtraq-id"`
		} `xml:"other-id"`
		VulnSolution string `xml:"vuln-solution"`
	} `xml:"entry"`
}

// 解析XML文件提取cveid及description,并写入sqlite数据库
func BuildCVE(fileXml string, db *sql.DB) (string, error) {
	var i int
	var id, description, vulnid string
	//var description ,description string
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fileXml); err != nil {

		log.Println("解析XML文件", fileXml, "错误:", err)
		return fileXml, err
	}
	root := doc.SelectElement("cnnvd")
	// 开始数据库事务
	tx, _ := db.Begin()
	for _, entry := range root.SelectElements("entry") {
		i = i + 1
		//fmt.Println("CHILD element:", book.Tag)
		if desc := entry.SelectElement("vuln-descript"); desc != nil {
			//lang := title.SelectAttrValue("lang", "unknown")
			//fmt.Printf("  Description: %s\n", desc.Text())
			description = desc.Text()
		}
		if vid := entry.SelectElement("vuln-id"); vid != nil {
			//lang := title.SelectAttrValue("lang", "unknown")
			//fmt.Printf("  vuln_id: %s\n", vid.Text())
			vulnid = vid.Text()
		}

		for _, cveid := range entry.SelectElements("other-id") {
			if cveid := cveid.SelectElement("cve-id"); cveid != nil {
				//fmt.Printf("CVEID:%s\n", cveid.Text())
				id = cveid.Text()
			}
		}

		tx.Exec("insert into cnnvd (other_id_cve_id,vuln_id, vuln_descript ) values(?,?,?)", id, vulnid, description)
	}
	// 事务提交
	tx.Commit()
	log.Println("导入文档数量：", i)
	return fileXml, nil
}
