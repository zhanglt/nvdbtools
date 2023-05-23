package common

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"github.com/neuvector/neuvector/share/utils"
)

type RawFile struct {
	Name string
	Raw  []byte
}

type memDB struct {
	keyVer   KeyVersion
	tbPath   string
	tmpPath  string
	vuls     map[string]VulFull
	appVuls  []AppModuleVul
	rawFiles []RawFile
}

var rawFilenames []string = []string{
	RHELCpeMapFile,
}

const (
	dbUbuntu = iota
	dbDebian
	dbCentos
	dbAlpine
	dbAmazon
	dbOracle
	dbMariner
	dbSuse
	dbMax
)

type dbBuffer struct {
	namespace string
	indexFile string
	fullFile  string
	indexBuf  bytes.Buffer
	fullBuf   bytes.Buffer
	indexSHA  [sha256.Size]byte
	fullSHA   [sha256.Size]byte
}

type dbSpace struct {
	buffers [dbMax]dbBuffer
	appBuf  bytes.Buffer
	appSHA  [sha256.Size]byte
	rawSHA  [][sha256.Size]byte
}

func (db *memDB) RebuildDb(version, srcPath string) bool {
	// if len(db.vuls) == 0 {
	// 		log.Errorf("CVE update FAIL")
	// 		return false
	// 	}
	var dbs dbSpace
	dbs.buffers[dbUbuntu] = dbBuffer{namespace: "ubuntu", indexFile: "ubuntu_index.tb", fullFile: "ubuntu_full.tb"}
	dbs.buffers[dbDebian] = dbBuffer{namespace: "debian", indexFile: "debian_index.tb", fullFile: "debian_full.tb"}
	dbs.buffers[dbCentos] = dbBuffer{namespace: "centos", indexFile: "centos_index.tb", fullFile: "centos_full.tb"}
	dbs.buffers[dbAlpine] = dbBuffer{namespace: "alpine", indexFile: "alpine_index.tb", fullFile: "alpine_full.tb"}
	dbs.buffers[dbAmazon] = dbBuffer{namespace: "amzn", indexFile: "amazon_index.tb", fullFile: "amazon_full.tb"}
	dbs.buffers[dbOracle] = dbBuffer{namespace: "oracle", indexFile: "oracle_index.tb", fullFile: "oracle_full.tb"}
	dbs.buffers[dbMariner] = dbBuffer{namespace: "mariner", indexFile: "mariner_index.tb", fullFile: "mariner_full.tb"}
	dbs.buffers[dbSuse] = dbBuffer{namespace: "sles", indexFile: "suse_index.tb", fullFile: "suse_full.tb"}

	ok := loadDbs(db, &dbs, srcPath)
	if !ok {
		log.Error("load database error")
		return false
	}

	var compactDB DBFile
	var regularDB DBFile

	// Compact database is consumed by scanners running inside controller. This scanner
	// in old versions cannot parse the regular db because of the header size limit
	// No new entries should be added !!!
	{
		//从/tmp/neuvector/db/keys中载入数据
		keyVer := KeyVersion{
			Version:    version,
			UpdateTime: time.Now().Format(time.RFC3339),
			Keys:       db.keyVer.Keys,
			Shas:       make(map[string]string, 0),
		}

		for _, i := range []int{dbUbuntu, dbDebian, dbCentos, dbAlpine} {
			buf := &dbs.buffers[i]
			keyVer.Shas[buf.indexFile] = fmt.Sprintf("%x", buf.indexSHA)
			keyVer.Shas[buf.fullFile] = fmt.Sprintf("%x", buf.fullSHA)
		}
		keyVer.Shas["apps.tb"] = fmt.Sprintf("%x", dbs.appSHA)

		var files []utils.TarFileInfo
		for _, i := range []int{dbUbuntu, dbDebian, dbCentos, dbAlpine} {
			buf := &dbs.buffers[i]
			files = append(files, utils.TarFileInfo{buf.indexFile, buf.indexBuf.Bytes()})
			files = append(files, utils.TarFileInfo{buf.fullFile, buf.fullBuf.Bytes()})
		}
		files = append(files, utils.TarFileInfo{"apps.tb", dbs.appBuf.Bytes()})

		compactDB.Filename = db.tbPath + CompactCVEDBName
		compactDB.Key = keyVer
		compactDB.Files = files
	}

	// regular files
	{
		keyVer := KeyVersion{
			Version:    version,
			UpdateTime: time.Now().Format(time.RFC3339),
			Keys:       db.keyVer.Keys,
			Shas:       make(map[string]string, 0),
		}

		for i := 0; i < dbMax; i++ {
			buf := &dbs.buffers[i]
			keyVer.Shas[buf.indexFile] = fmt.Sprintf("%x", buf.indexSHA)
			keyVer.Shas[buf.fullFile] = fmt.Sprintf("%x", buf.fullSHA)
		}
		keyVer.Shas["apps.tb"] = fmt.Sprintf("%x", dbs.appSHA)

		var files []utils.TarFileInfo
		for i := 0; i < dbMax; i++ {
			buf := &dbs.buffers[i]
			files = append(files, utils.TarFileInfo{buf.indexFile, buf.indexBuf.Bytes()})
			files = append(files, utils.TarFileInfo{buf.fullFile, buf.fullBuf.Bytes()})
			log.WithFields(log.Fields{"database": buf.namespace, "size": buf.fullBuf.Len()}).Info()
		}
		files = append(files, utils.TarFileInfo{"apps.tb", dbs.appBuf.Bytes()})
		log.WithFields(log.Fields{"database": "apps", "size": dbs.appBuf.Len()}).Info()
		for i, v := range db.rawFiles {
			files = append(files, utils.TarFileInfo{v.Name, v.Raw})
			keyVer.Shas[v.Name] = fmt.Sprintf("%x", dbs.rawSHA[i])
			log.WithFields(log.Fields{"database": v.Name, "size": len(v.Raw)}).Info()
		}

		regularDB.Filename = db.tbPath + RegularCVEDBName
		regularDB.Key = keyVer
		regularDB.Files = files
	}

	for _, dbf := range []*DBFile{&compactDB, &regularDB} {
		CreateDBFile(dbf)
	}

	return true
}
func loadDbs(db *memDB, dbs *dbSpace, srcPath string) bool {

	for i := 0; i < dbMax; i++ {
		buf := &dbs.buffers[i]
		if index, err := readCveData(srcPath + buf.indexFile); err == nil {
			buf.indexBuf.WriteString(fmt.Sprintf("%s", index)) //%s之后不要加\n
		} else {
			log.Println("读入:", buf.indexFile, "文件错误:", err)
		}

		if full, err := readCveData(srcPath + buf.fullFile); err == nil {
			buf.fullBuf.WriteString(fmt.Sprintf("%s", full)) //%s之后不要加\n
		} else {
			log.Println("读入:", buf.fullFile, "文件错误:", err)
		}
		buf.indexSHA = sha256.Sum256(buf.indexBuf.Bytes())
		buf.fullSHA = sha256.Sum256(buf.fullBuf.Bytes())
	}
	if app, err := readCveData(srcPath + "apps.tb"); err == nil {
		dbs.appBuf.WriteString(fmt.Sprintf("%s", app)) //%s之后不要加\n
	} else {
		log.Println("读入apps.tb文件错误:", err)
	}
	dbs.appSHA = sha256.Sum256(dbs.appBuf.Bytes())
	if raw, err := readCveData(srcPath + "rhel-cpe.map"); err == nil {
		//dbs.appBuf.WriteString(fmt.Sprintf("%s\n", raw))
		//db.rawFiles[0].Raw = raw
		db.rawFiles = append(db.rawFiles, RawFile{Name: "rhel-cpe.map", Raw: raw})
		//dbs.rawSHA[0] = sha256.Sum256(raw)
	}
	//rawFile := RawFile{Name: "rhel-cpe.map", Raw: db.rawFiles[0].Raw}
	dbs.rawSHA = make([][sha256.Size]byte, len(db.rawFiles))

	for i, v := range db.rawFiles {

		dbs.rawSHA[i] = sha256.Sum256(v.Raw)
	}

	return true
}

func MemdbOpen(path string) (*memDB, error) {
	// 创建临时目录
	dir, err := ioutil.TempDir("", "cve")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("创建临时CVE目录失败")
		return nil, err
	}
	var db memDB
	db.tbPath = path
	db.tmpPath = dir
	db.vuls = make(map[string]VulFull, 0)
	db.keyVer.Keys = make(map[string]string, 0)
	db.keyVer.Shas = make(map[string]string, 0)

	return &db, nil

}
func readCveData(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println("文件打开", fileName, "错误:", err)
		return nil, err
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("读取", fileName, "文件错误:", err)
		return nil, err
	}
	return body, nil
}
func (db *memDB) Close() {
	os.RemoveAll(db.tmpPath)
}
