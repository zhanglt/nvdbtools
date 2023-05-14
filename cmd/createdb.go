/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/common"
)

// createdbCmd represents the createdb command
var createdbCmd = &cobra.Command{
	Use:   "createdb",
	Short: "重新打包cvedb数据库",
	Long:  `重新打包cvedb数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		dbPath, _ := cmd.Flags().GetString("dbPath")
		srcPath, _ := cmd.Flags().GetString("srcPath")
		db, err := memdbOpen(dbPath)
		if err != nil {
			log.Fatalln("数据库初始化错误", err)
			os.Exit(2)
		}
		defer db.Close()
		kver, _ := getVersion(srcPath + "keys")
		log.Println("数据库版本号：", kver)
		if db.RewriteDb(fmt.Sprintf("%f", kver), srcPath) {
			log.Println("cvedb重新打包成功")
		} else {
			log.Fatal("cvedb重新打包失败")
		}
		fmt.Println("createdb called")
	},
}

func init() {
	rootCmd.AddCommand(createdbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createdbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createdbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	createdbCmd.Flags().StringP("dbPath", "d", "/tmp/nvdbtools/cvedbtemp/", "重新打包cvedb的存放目录")
	createdbCmd.Flags().StringP("srcPath", "s", "/tmp/nvdbtools/cvedbtarget/", "cvedb解压后的目录")
}
func getVersion(keyFile string) (string, error) {
	// Read json buffer from jsonFile
	byteValue, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}

	// We have known the outer json object is a map, so we define  result as map.
	// otherwise, result could be defined as slice if outer is an array
	var result common.KeyVer
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return "", err
	}
	return result.Version, nil
}
