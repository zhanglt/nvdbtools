package common

import (
	"time"

	"github.com/neuvector/neuvector/share/utils"
	//"github.com/neuvector/neuvector/share/utils"
)

const CompactCVEDBName = "cvedb.compact"
const RegularCVEDBName = "cvedb.regular"

const RHELCpeMapFile = "rhel-cpe.map"

type NVDMetadata struct {
	Description      string `json:"description,omitempty"`
	CVSSv2           CVSS
	CVSSv3           CVSS
	VulnVersions     []NVDvulnerableVersion
	PublishedDate    time.Time
	LastModifiedDate time.Time
}

type NVDvulnerableVersion struct {
	StartIncluding string
	StartExcluding string
	EndIncluding   string
	EndExcluding   string
}

type CVSS struct {
	Vectors string
	Score   float64
}

// database format

type DBFile struct {
	Filename string
	Key      KeyVersion
	Files    []utils.TarFileInfo
}

type KeyVersion struct {
	Version    string
	UpdateTime string
	Keys       map[string]string
	Shas       map[string]string
}

type FeaShort struct {
	Name    string `json:"N"`
	Version string `json:"V"`
	MinVer  string `json:"MV"`
}

type VulShort struct {
	Name      string `json:"N"`
	Namespace string `json:"NS"`
	Fixin     []FeaShort
	CPEs      []string `json:"CPE"`
}

type FeaFull struct {
	Name      string `json:"N"`
	Namespace string `json:"NS"`
	Version   string `json:"V"`
	MinVer    string `json:"MV"`
	AddedBy   string `json:"A"`
}

type VulFull struct {
	Name        string    `json:"N"`
	Namespace   string    `json:"NS"`
	Description string    `json:"D"`
	Link        string    `json:"L"`
	Severity    string    `json:"S"`
	CVSSv2      CVSS      `json:"C2"`
	CVSSv3      CVSS      `json:"C3"`
	FixedBy     string    `json:"FB"`
	FixedIn     []FeaFull `json:"FI"`
	CPEs        []string  `json:"CPE,omitempty"`
	CVEs        []string  `json:"CVE,omitempty"`
	FeedRating  string    `json:"RATE,omitempty"`
	IssuedDate  time.Time `json:"Issue"`
	LastModDate time.Time `json:"LastMod"`
}

type AppModuleVersion struct {
	OpCode  string `json:"O"`
	Version string `json:"V"`
}

type AppModuleVul struct {
	VulName       string             `json:"VN"`
	AppName       string             `json:"AN"`
	ModuleName    string             `json:"MN"`
	Description   string             `json:"D"`
	Link          string             `json:"L"`
	Score         float64            `json:"SC"`
	Vectors       string             `json:"VV2"`
	ScoreV3       float64            `json:"SC3"`
	VectorsV3     string             `json:"VV3"`
	Severity      string             `json:"SE"`
	AffectedVer   []AppModuleVersion `json:"AV"`
	FixedVer      []AppModuleVersion `json:"FV"`
	UnaffectedVer []AppModuleVersion `json:"UV",omitempty`
	IssuedDate    time.Time          `json:"Issue"`
	LastModDate   time.Time          `json:"LastMod"`
	CVEs          []string           `json:"-"`
}

// ---

// UbuntuReleasesMapping translates Ubuntu code names to version numbers
var UbuntuReleasesMapping = map[string]string{
	"upstream":         "upstream",
	"precise":          "12.04",
	"precise/esm":      "12.04",
	"quantal":          "12.10",
	"raring":           "13.04",
	"trusty":           "14.04",
	"trusty/esm":       "14.04",
	"utopic":           "14.10",
	"vivid":            "15.04",
	"wily":             "15.10",
	"xenial":           "16.04",
	"esm-infra/xenial": "16.04",
	"yakkety":          "16.10",
	"zesty":            "17.04",
	"artful":           "17.10",
	"bionic":           "18.04",
	"cosmic":           "18.10",
	"disco":            "19.04",
	"eoan":             "19.10",
	"focal":            "20.04",
	"groovy":           "20.10",
	"hirsute":          "21.04",
	"impish":           "21.10",
	"jammy":            "22.04",
	"kinetic":          "22.10",
}

var DebianReleasesMapping = map[string]string{
	// Code names
	"squeeze":  "6",
	"wheezy":   "7",
	"jessie":   "8",
	"stretch":  "9",
	"buster":   "10",
	"bullseye": "11",
	"sid":      "unstable",

	// Class names
	"oldoldstable": "7",
	"oldstable":    "8",
	"stable":       "9",
	"testing":      "10",
	"unstable":     "unstable",
}

type Centos struct {
	N  string `json:"N"`
	NS string `json:"NS"`
	D  string `json:"D"`
	L  string `json:"L"`
	S  string `json:"S"`
	C2 struct {
		Vectors string  `json:"Vectors"`
		Score   float64 `json:"Score"`
	} `json:"C2"`
	C3 struct {
		Vectors string  `json:"Vectors"`
		Score   float64 `json:"Score"`
	} `json:"C3"`
	FB string `json:"FB"`
	FI []struct {
		N  string `json:"N"`
		NS string `json:"NS"`
		V  string `json:"V"`
		MV string `json:"MV"`
		A  string `json:"A"`
	} `json:"FI"`
	CPE     []string  `json:"CPE"`
	CVE     []string  `json:"CVE"`
	RATE    string    `json:"RATE"`
	Issue   time.Time `json:"Issue"`
	LastMod time.Time `json:"LastMod"`
}
type Apps struct {
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
type KeyVer struct {
	Version    string    `json:"Version"`
	UpdateTime time.Time `json:"UpdateTime"`
	Keys       struct {
	} `json:"Keys"`
	Shas struct {
		AlpineFullTb   string `json:"alpine_full.tb"`
		AlpineIndexTb  string `json:"alpine_index.tb"`
		AmazonFullTb   string `json:"amazon_full.tb"`
		AmazonIndexTb  string `json:"amazon_index.tb"`
		AppsTb         string `json:"apps.tb"`
		CentosFullTb   string `json:"centos_full.tb"`
		CentosIndexTb  string `json:"centos_index.tb"`
		DebianFullTb   string `json:"debian_full.tb"`
		DebianIndexTb  string `json:"debian_index.tb"`
		MarinerFullTb  string `json:"mariner_full.tb"`
		MarinerIndexTb string `json:"mariner_index.tb"`
		OracleFullTb   string `json:"oracle_full.tb"`
		OracleIndexTb  string `json:"oracle_index.tb"`
		RhelCpeMap     string `json:"rhel-cpe.map"`
		SuseFullTb     string `json:"suse_full.tb"`
		SuseIndexTb    string `json:"suse_index.tb"`
		UbuntuFullTb   string `json:"ubuntu_full.tb"`
		UbuntuIndexTb  string `json:"ubuntu_index.tb"`
	} `json:"Shas"`
}
