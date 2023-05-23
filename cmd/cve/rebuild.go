/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cve

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/zhanglt/nvdbtools/common"
)

// createdbCmd represents the createdb command
var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "重新打包cvedb数据库",
	Long:  `重新打包cvedb数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取重新打包后文件的存放路径
		dbPath, _ := cmd.Flags().GetString("dbPath")
		os.RemoveAll(dbPath)
		os.MkdirAll(dbPath, 0755)

		// 获取打包源文件路径
		srcPath, err := cmd.Flags().GetString("srcPath")
		if err != nil {
			log.Fatalln("获取源文件路径错误", err)
			return
		}
		db, err := common.MemdbOpen(dbPath)
		if err != nil {
			log.Fatalln("数据库初始化错误", err)
			return
		}
		defer db.Close()
		//  从key文件中获取当前cvedb的版本号
		kver, err := getVersion(srcPath + "keys")
		if err != nil {
			log.Fatalln("获取数据版本错误:", err)
			return
		}
		// 打包数据库
		if db.RebuildDb(kver, srcPath) {
			log.Println("cvedb重新打包成功,文件保存在目录:", dbPath)
		} else {
			log.Fatal("cvedb重新打包失败")
			return
		}

	},
}

func init() {

	rebuildCmd.Flags().StringP("dbPath", "d", "/tmp/nvdbtools/cvedbtemp/", "重新打包cvedb的存放目录")
	rebuildCmd.Flags().StringP("srcPath", "s", "/tmp/nvdbtools/cvedbtarget/", "cvedb解压后的目录")
}

// 从解压出来的key文件中读取cvedb版本号
func getVersion(keyFile string) (string, error) {
	byteValue, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}
	var result common.KeyVer
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Fatalln("获取数据版本号错误", err)
		return "", err
	}
	return result.Version, nil
}
