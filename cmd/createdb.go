/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/common"
)

// createdbCmd represents the createdb command
var createdbCmd = &cobra.Command{
	Use:   "createdb",
	Short: "重新打包cvedb数据库",
	Long:  `重新打包cvedb数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取重新打包后文件的存放路径
		dbPath, _ := cmd.Flags().GetString("dbPath")
		// 获取打包源文件路径
		srcPath, _ := cmd.Flags().GetString("srcPath")
		db, err := memdbOpen(dbPath)
		if err != nil {
			log.Fatalln("数据库初始化错误", err)
			return
		}
		defer db.Close()
		//  从keys文件中获取当前cvedb的版本号
		kver, err := getVersion(srcPath + "keys")
		if err != nil {
			log.Fatalln("获取数据版本错误:", err)
			return
		}
		// 打包数据库
		if db.RebuildDb(kver, srcPath) {
			log.Println("cvedb重新打包成功，文件保存在目录:", dbPath)
		} else {
			log.Fatal("cvedb重新打包失败")
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(createdbCmd)
	createdbCmd.Flags().StringP("dbPath", "d", "/tmp/nvdbtools/cvedbtemp/", "重新打包cvedb的存放目录")
	createdbCmd.Flags().StringP("srcPath", "s", "/tmp/nvdbtools/cvedbtarget/", "cvedb解压后的目录")
}
func getVersion(keyFile string) (string, error) {
	byteValue, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}
	var result common.KeyVer
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return "", err
	}
	return result.Version, nil
}
