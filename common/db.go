package common

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"unicode"

	log "github.com/sirupsen/logrus"

	//"github.com/neuvector/neuvector/share"

	"github.com/neuvector/neuvector/share/utils"
)

func CreateDBFile(dbFile *DBFile) error {
	log.WithFields(log.Fields{"file": dbFile.Filename}).Info("Create database file")

	header, _ := json.Marshal(dbFile.Key)

	buf, err := utils.MakeTar(dbFile.Files)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Make tar file error")
		return err
	}
	zb := utils.GzipBytes(buf.Bytes())

	// Use local encrypt function
	cipherData, err := encrypt(zb, getCVEDBEncryptKey())
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Encrypt tar file fail")
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
		log.WithFields(log.Fields{"error": err}).Error("Create db file fail")
		return err
	}
	defer fdb.Close()

	n, err := fdb.Write(allb.Bytes())
	if err != nil || n != allb.Len() {
		log.WithFields(log.Fields{"error": err}).Error("Write file error")
		return err
	}

	log.WithFields(log.Fields{"file": dbFile.Filename, "size": allb.Len()}).Info("Create database done")
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
