/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cnnvd

import (
	"database/sql"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/cnnvd"
)

// importDBCmd represents the importDB command
var importDBCmd = &cobra.Command{
	Use:   "importDB",
	Short: "cnnvd xml数据导入到sqlite数据库",
	Long:  `cnnvd xml数据导入到sqlite数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, err := cmd.Flags().GetString("filePath")
		if err != nil {
			log.Println("请输入filePath:", filePath)
		}
		// 获取数据库对象
		db, err := getcveDB()
		if err != nil {
			log.Println("获取sqlite数据库错误", err)
			return

		}
		//获取XML文件名称列表
		xmlList, err := ioutil.ReadDir(filePath)
		if err != nil {
			log.Fatal(err)
		}
		for i := range xmlList {
			//sqlite并发写有问题
			//go func(index int, dbName *sql.DB) {
			//f, err := cnnvd.BuildCVE(filePath+xmlList[index].Name(), dbName)
			f, err := cnnvd.BuildCVE(filePath+xmlList[i].Name(), db)
			if err != nil {
				log.Printf("%s文件导入错误:%s/n", f, err)
			} else {
				log.Println("文件导入完成:", f)
			}

			//}(i, db)
		}

	},
}

func init() {
	CnnvdCmd.AddCommand(importDBCmd)
	importDBCmd.Flags().StringP("filePath", "f", "/tmp/nvdbtools/xml/", "cnnvd xml数据文件存放目录")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importDBCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importDBCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func getcveDB() (*sql.DB, error) {
	DB, err := sql.Open("sqlite3", "cnnvd.db")
	if err != nil {
		log.Println("数据(cnnvd.db)打开错误：", err)
		return nil, err
		err = DB.Ping()
		if err != nil {
			log.Println("数据库(cnnvd.db)测试错误：", err)
			return nil, err
		}
	}

	_, err = DB.Exec(tableInit)
	// 在不开启事务时提升数据插入性能
	DB.Exec("PRAGMA synchronous = 0;PRAGMA journal_mode = OFF")
	if err != nil {
		log.Println("初始化数据表错误", ":", err)
		return nil, err
	}
	//DB.exec(fmt.Sprintf("PRAGMA synchronous = OFF;"))

	return DB, nil
}

var tableInit string = `
DROP TABLE IF EXISTS cnnvd;
CREATE TABLE [cnnvd] (
[vuln_id] varchar(255),
[vuln_descript] text,
[other_id_cve_id] varchar(255)
);
CREATE INDEX cnnvd_other_id_cve_id_IDX ON cnnvd (other_id_cve_id);
`
