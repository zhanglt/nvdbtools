/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cve

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/common"
)

var unzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "解压scanner的cvedb数据库",
	Long:  `解压scanner的cvedb数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		// 从scanner中获取的cvedb数据库文件路径
		cvedbPath, err := cmd.Flags().GetString("cvedbPath")
		if err != nil {
			log.Println("请输入正确cvedbPath")
			return

		}
		// cvedb解压后的文件存放路径
		unzipPath, err := cmd.Flags().GetString("unzipPath")
		if err != nil {
			log.Println("请输入正确unzipPath")
			return
		}
		os.RemoveAll(unzipPath)
		os.MkdirAll(unzipPath, 0755)
		// 解压cvedb数据库
		if err := common.UNzipDb(cvedbPath, unzipPath); err == nil {
			log.Println("cvedb数据库解压完成")
		} else {
			log.Println("cvedb数据库解压失败")
			return
		}
	},
}

func init() {

	unzipCmd.Flags().StringP("cvedbPath", "c", "/tmp/nvdbtools/cvedb", "scanner中提取的cvedb文件路径")
	unzipCmd.Flags().StringP("unzipPath", "u", "/tmp/nvdbtools/cvedbsrc/", "cvedb中提取文件保存路径")
}
