/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cve

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/zhanglt/nvdbtools/common"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "用cnnvd数据库更新cve的说明信息",
	Long:  `用cnnvd数据库更新cve的说明信息,如果不在cnnvd中的数据调用google翻译引擎做翻译`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取cvedb解压后的文件路径
		unzipPath, _ := cmd.Flags().GetString("unzipPath")
		// 获取更新description之后的文件存放路径
		targetPath, _ := cmd.Flags().GetString("targetPath")
		os.RemoveAll(targetPath)
		os.MkdirAll(targetPath, 0755)
		full := []string{"alpine_full.tb", "amazon_full.tb", "centos_full.tb", "debian_full.tb", "mariner_full.tb", "oracle_full.tb", "suse_full.tb", "ubuntu_full.tb"}
		index := []string{"alpine_index.tb", "amazon_index.tb", "centos_index.tb", "debian_index.tb", "mariner_index.tb", "oracle_index.tb", "suse_index.tb", "ubuntu_index.tb"}
		// 打开cnvd数据
		DB, err := getDB()
		if err != nil {
			log.Println("获取sqlit db数据库错误:", err)
		}

		for _, file := range full {
			// 更新full数据
			err := common.UpdateDescription(unzipPath+file, targetPath+file, "full", DB)
			if err != nil {
				log.Println(unzipPath+file, "数据文件更新(cve说明)错误：", err)
			} else {
				log.Println("文件更新完毕:", targetPath+file)
			}

		}
		// 更新apps数据
		err = common.UpdateDescription(unzipPath+"apps.tb", targetPath+"apps.tb", "apps", DB)
		if err != nil {
			log.Println(unzipPath+"apps.tb", "数据文件更新(cve说明)错误：", err)
		} else {
			log.Println("文件更新完毕:", targetPath+"apps.tb")
		}
		for _, file := range index {
			// 复制 index数据文件到目标路径
			CopyFile(targetPath+file, unzipPath+file)
			log.Println("文件更新完毕:", targetPath+file)
		}
		// 复制 cpe数据到目标路径
		CopyFile(targetPath+"rhel-cpe.map", unzipPath+"rhel-cpe.map")
		// 复制keys数据到目标路径
		CopyFile(targetPath+"keys", unzipPath+"keys")
		fmt.Println("cve说明更新完毕")
	},
}

func init() {

	updateCmd.Flags().StringP("unzipPath", "u", "/tmp/nvdbtools/cvedbsrc/", "cvedb解压后的目录")
	updateCmd.Flags().StringP("targetPath", "t", "/tmp/nvdbtools/cvedbtarget/", "cvedb解压后的目录")

}
func CopyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		log.Println("打开源文件错误，错误信息:", err)
	}
	defer srcFile.Close()
	reader := bufio.NewReader(srcFile)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("打开目标文件错误，错误信息:", err)
		return
	}
	writer := bufio.NewWriter(dstFile)
	defer dstFile.Close()
	return io.Copy(writer, reader)
}
func getDB() (*sql.DB, error) {
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
	// 创建新表，保存cnnvd之外的条目的翻译信息
	_, err = DB.Exec(translateTable)
	return DB, nil
}

// 表结构用于保存cnnvd之外的条目的翻译信息
var translateTable string = `
CREATE TABLE IF NOT EXISTS [translate] (
[cve_id] varchar(255),
[descript] text
);
CREATE INDEX translate_cve_id_IDX ON "translate" (cve_id);
`
