package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	//"github.com/vul-dbgen/updater"
	"github.com/zhanglt/nvdbtools/common"
	// utils "github.com/zhanglt/nvdbtools/share"
	"github.com/neuvector/neuvector/share/utils"
)

type cvdData struct {
	An string `json:"AN"`
	Av []struct {
		O string `json:"O"`
		V string `json:"V"`
	} `json:"AV"`
	D  string `json:"D"`
	Fv []struct {
		O string `json:"O"`
		V string `json:"V"`
	} `json:"FV"`
	Issue   time.Time   `json:"Issue"`
	L       string      `json:"L"`
	LastMod time.Time   `json:"LastMod"`
	Mn      string      `json:"MN"`
	Sc      float64     `json:"SC"`
	Sc3     float64     `json:"SC3"`
	Se      string      `json:"SE"`
	Uv      interface{} `json:"UV"`
	Vn      string      `json:"VN"`
	Vv2     string      `json:"VV2"`
	Vv3     string      `json:"VV3"`
}

var DB *sql.DB
var err error

func init() {
	//打开 cnvd数据库
	DB, err = sql.Open("sqlite3", "cnvd20230428.db")
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		return
	}
	DB.SetMaxOpenConns(200)                 //最大连接数
	DB.SetMaxIdleConns(10)                  //连接池里最大空闲连接数。必须要比maxOpenConns小
	DB.SetConnMaxLifetime(time.Second * 10) //最大存活保持时间
	DB.SetConnMaxIdleTime(time.Second * 10) //最大空闲保持时间

}

// 更具cve编号 搜索在cnvd中的中文说明
func getDescribe(db *sql.DB, cveid string, srcDescribe string) string {
	var number, title, serverity, products, isEvent, submitTime, openTime, discovererName, referenceLink, formalWay, description, patchDescription, patchName, cveNumber, bids, cveUrl, cves sql.NullString
	rows := db.QueryRow("SELECT * FROM cnvd20230428 where cveNumber=$1", cveid)
	err = rows.Scan(&number, &title, &serverity, &products, &isEvent, &submitTime, &openTime, &discovererName, &referenceLink, &formalWay, &description, &patchDescription, &patchName, &cveNumber, &bids, &cveUrl, &cves)
	if err != nil {
		//	log.Debug("cve ID:", cveid, "数据库中没有搜索到:", err)
		return srcDescribe
	}

	return description.String
}

type RawFile struct {
	Name string
	Raw  []byte
}

type memDB struct {
	keyVer   common.KeyVersion
	tbPath   string
	tmpPath  string
	vuls     map[string]common.VulFull
	appVuls  []common.AppModuleVul
	rawFiles []RawFile
}

func newMemDb() (*memDB, error) {
	var db memDB
	db.vuls = make(map[string]common.VulFull, 0)
	db.keyVer.Keys = make(map[string]string, 0)
	db.keyVer.Shas = make(map[string]string, 0)
	return &db, nil
}

var rawFilenames []string = []string{
	common.RHELCpeMapFile,
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

func (db *memDB) CreateDb(version string, srcPath string) bool {
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

	//dbs.rawSHA = make([][sha256.Size]byte, len(db.rawFiles))
	ok := loadDbs(db, &dbs, srcPath)
	if !ok {
		log.Error("load database error")
		return false
	}

	var compactDB common.DBFile
	var regularDB common.DBFile

	// Compact database is consumed by scanners running inside controller. This scanner
	// in old versions cannot parse the regular db because of the header size limit
	// No new entries should be added !!!
	{
		//从/tmp/neuvector/db/keys中载入数据
		keyVer := common.KeyVersion{
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

		compactDB.Filename = db.tbPath + common.CompactCVEDBName
		compactDB.Key = keyVer
		compactDB.Files = files
	}

	// regular files
	{
		keyVer := common.KeyVersion{
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

		regularDB.Filename = db.tbPath + common.RegularCVEDBName
		regularDB.Key = keyVer
		regularDB.Files = files
	}

	for _, dbf := range []*common.DBFile{&compactDB, &regularDB} {
		common.CreateDBFile(dbf)
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

func memdbOpen(path string) (*memDB, error) {
	dir, err := ioutil.TempDir("", "cve")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create tmp cve directory")
		return nil, err
	}
	var db memDB
	db.tbPath = path
	db.tmpPath = dir
	db.vuls = make(map[string]common.VulFull, 0)
	db.keyVer.Keys = make(map[string]string, 0)
	db.keyVer.Shas = make(map[string]string, 0)

	return &db, nil

}
func readCveData(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println("file open error:", err)
		return nil, err
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("readall  error:", err)
		return nil, err
	}
	//log.Println("是否为有效json格式:", fileName, "json格式:", validator.Valid(body))
	return body, nil
}
func (db *memDB) Close() {
	os.RemoveAll(db.tmpPath)
}

func cveDescrition(str string) string {
	buf := bytes.NewBufferString(str)
	scanner := bufio.NewScanner(buf)
	newBuf := bytes.Buffer{}
	for scanner.Scan() {
		newBuf.WriteString(scanner.Text())
	}
	return newBuf.String()
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

		var rtep *cvdData = new(cvdData)
		err := json.Unmarshal(json_message, rtep)
		if err != nil {
			i = i + 1
			log.Println("序列化失败======:", i, err)
		}
		//log.Println("cvd id =======:", rtep.Vn)
	}

	return body, nil

}
