/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

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

// updescCmd represents the updesc command
var updescCmd = &cobra.Command{
	Use:   "updesc",
	Short: "用cnvd数据库更新cve的说明信息",
	Long:  `用cnvd数据库更新cve的说明信息`,
	Run: func(cmd *cobra.Command, args []string) {

		unzipPath, _ := cmd.Flags().GetString("unzipPath")
		targetPath, _ := cmd.Flags().GetString("targetPath")
		full := []string{"alpine_full.tb", "amazon_full.tb", "centos_full.tb", "debian_full.tb", "mariner_full.tb", "oracle_full.tb", "suse_full.tb", "ubuntu_full.tb"}
		index := []string{"alpine_index.tb", "amazon_index.tb", "centos_index.tb", "debian_index.tb", "mariner_index.tb", "oracle_index.tb", "suse_index.tb", "ubuntu_index.tb"}

		DB, err := sql.Open("sqlite3", "cnvd20230428.db")
		if err != nil {
			log.Println("数据打开错误：", err)
			return
			err = DB.Ping()
			if err != nil {
				log.Println("数据库测试错误：", err)
				return
			}
		}
		for _, file := range full {
			common.UpdateDescription(unzipPath+file, targetPath+file, "full", DB)
		}
		common.UpdateDescription(unzipPath+"apps.tb", targetPath+"apps.tb", "apps", DB)
		for _, file := range index {
			CopyFile(targetPath+file, unzipPath+file)
		}
		CopyFile(targetPath+"rhel-cpe.map", unzipPath+"rhel-cpe.map")

		CopyFile(targetPath+"keys", unzipPath+"keys")
		fmt.Println("cve说明更新完毕")
	},
}

func init() {
	rootCmd.AddCommand(updescCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updescCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updescCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	updescCmd.Flags().StringP("unzipPath", "u", "/tmp/nvdbtools/cvedbsrc/", "cvedb解压后的目录")
	updescCmd.Flags().StringP("targetPath", "t", "/tmp/nvdbtools/cvedbtarget/", "cvedb解压后的目录")

}
func CopyFile(dstFilePath string, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Printf("打开源文件错误，错误信息=%v\n", err)
	}
	defer srcFile.Close()
	reader := bufio.NewReader(srcFile)

	dstFile, err := os.OpenFile(dstFilePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("打开目标文件错误，错误信息=%v\n", err)
		return
	}
	writer := bufio.NewWriter(dstFile)
	defer dstFile.Close()
	return io.Copy(writer, reader)
}
