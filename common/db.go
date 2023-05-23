package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"unicode"

	log "github.com/sirupsen/logrus"

	//"github.com/neuvector/neuvector/share"

	"github.com/neuvector/neuvector/share/utils"
)

func CreateDBFile(dbFile *DBFile) error {
	log.WithFields(log.Fields{"文件": dbFile.Filename}).Info("创建cvedb数据库文件")

	header, _ := json.Marshal(dbFile.Key)

	buf, err := utils.MakeTar(dbFile.Files)
	if err != nil {
		log.WithFields(log.Fields{"错误": err}).Error("tar打包错误")
		return err
	}
	zb := utils.GzipBytes(buf.Bytes())

	// Use local encrypt function
	cipherData, err := encrypt(zb, getCVEDBEncryptKey())
	if err != nil {
		log.WithFields(log.Fields{"错误": err}).Error("加密tar文件错误")
		return err
	}

	b0 := make([]byte, 0)
	allb := bytes.NewBuffer(b0)

	keyLen := int32(len(header))
	binary.Write(allb, binary.BigEndian, &keyLen)
	allb.Write(header)
	allb.Write(cipherData)

	// write to db file
	fdb, err := os.Create(dbFile.Filename)
	if err != nil {
		log.WithFields(log.Fields{"错误": err}).Error("创建db文件错误")
		return err
	}
	defer fdb.Close()

	n, err := fdb.Write(allb.Bytes())
	if err != nil || n != allb.Len() {
		log.WithFields(log.Fields{"错误": err}).Error("写文件错误")
		return err
	}

	log.WithFields(log.Fields{"文件": dbFile.Filename, "size": allb.Len()}).Info("打包数据库完毕")
	return nil
}

func ParseYear(name string) (int, error) {
	for i, r := range name {
		if !unicode.IsDigit(r) {
			return strconv.Atoi(name[:i])
		}
	}
	return strconv.Atoi(name)
}

const maxExtractSize = 0 // No extract limit
const maxVersionHeader = 100 * 1024
const maxBufferSize = 1024 * 1024

// 解压源cvedb库
func UNzipDb(nvCvedbPath, nvUnzipPath string) error {
	encryptKey := getCVEDBEncryptKey()
	f, err := os.Open(nvCvedbPath)
	if err != nil {
		log.Info("Open zip db file fail")
		return err
	}
	defer f.Close()

	f.Seek(0, 0)

	// read keys len
	bhead := make([]byte, 4)
	nlen, err := f.Read(bhead)
	if err != nil || nlen != 4 {
		log.WithFields(log.Fields{"error": err}).Error("Read db file error")
		return err
	}
	var headLen int32
	err = binary.Read(bytes.NewReader(bhead), binary.BigEndian, &headLen)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Write header len error")
		return err
	}
	if headLen > maxVersionHeader {
		log.Info("Version Header too big:", headLen)
		return err
	}

	// Read head and write keys file
	bhead = make([]byte, headLen)
	nlen, err = f.Read(bhead)
	if err != nil || nlen != int(headLen) {
		log.WithFields(log.Fields{"error": err}).Error("Read db file error")
		return err
	}
	err = ioutil.WriteFile(nvUnzipPath+"keys", bhead, 0400)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Write keys file error")
		return err
	} else {
		log.Println("keys文件解压完成")
	}

	// Read the rest of DB
	cipherData, err := ioutil.ReadAll(f)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Read db file tar part error")
		return err
	}

	// Use local decrypt function
	plainData, err := decrypt(cipherData, encryptKey)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Decrypt tar file error")
		return err
	}

	tarFile := bytes.NewReader(plainData)
	err = utils.ExtractAllArchiveToFiles(nvUnzipPath, tarFile, maxExtractSize, nil)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Extract db file error")
		return err
	}

	return nil
}
func GetDbVersion(nvCvedbPath string) (float64, string, error) {
	f, err := os.Open(nvCvedbPath)
	if err != nil {
		return 0, "", fmt.Errorf("Read db file fail: %v", err)
	}
	defer f.Close()

	bhead := make([]byte, 4)
	nlen, err := f.Read(bhead)
	if err != nil || nlen != 4 {
		return 0, "", fmt.Errorf("Read db file error: %v", err)
	}
	var headLen int32
	err = binary.Read(bytes.NewReader(bhead), binary.BigEndian, &headLen)
	if err != nil {
		return 0, "", fmt.Errorf("Read header len error: %v", err)
	}

	if headLen > maxVersionHeader {
		return 0, "", fmt.Errorf("Version Header too big: %v", headLen)
	}

	bhead = make([]byte, headLen)
	nlen, err = f.Read(bhead)
	if err != nil || nlen != int(headLen) {
		return 0, "", fmt.Errorf("Read db file version error:%v", err)
	}

	var keyVer KeyVersion

	err = json.Unmarshal(bhead, &keyVer)
	if err != nil {
		return 0, "", fmt.Errorf("Unmarshal keys error:%v", err)
	}
	verFl, err := strconv.ParseFloat(keyVer.Version, 64)
	if err != nil {
		return 0, "", fmt.Errorf("Invalid version value:%v", err)
	}

	return verFl, keyVer.UpdateTime, nil
}
func ReadlineValid(fileName string) ([]byte, error) {
	var body []byte
	var i int = 0
	fp, err := os.Open(fileName)
	if err != nil {
		log.Println("file open error:", err)

	}
	defer fp.Close()
	br := bufio.NewReader(fp)
	for {
		json_message, _, c := br.ReadLine() //按行读取文件
		if c == io.EOF {
			break
		}
		//log.Println("json======:", string(json_message))

		//if !validator.Valid(json_message) {
		//	log.Println("格式验证错误==========:")
		//}
		body = json_message

		var rtep *Apps = new(Apps)
		err := json.Unmarshal(json_message, rtep)
		if err != nil {
			i = i + 1
			log.Println("序列化失败======:", i, err)
		}
		//log.Println("cvd id =======:", rtep.Vn)
	}

	return body, nil

}
