/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/common"
)

// uncveCmd represents the uncve command
var uncveCmd = &cobra.Command{
	Use:   "uncve",
	Short: "解压scanner的cvedb数据库",
	Long:  `解压scanner的cvedb数据库`,
	Run: func(cmd *cobra.Command, args []string) {

		cvedbPath, err := cmd.Flags().GetString("cvedbPath")
		if err != nil {
			log.Println("请输入正确cvedbPath")
			return

		}
		unzipPath, err := cmd.Flags().GetString("unzipPath")
		if err != nil {
			log.Println("请输入正确cvePath")
			return
		}

		log.Println(cvedbPath, unzipPath)

		if err := common.UNzipDb(cvedbPath, unzipPath); err == nil {
			log.Println("数据库解压完成")
		} else {
			log.Println("数据库解压失败")
			return
		}

		fmt.Println("uncve called")
	},
}

func init() {
	rootCmd.AddCommand(uncveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uncveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uncveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	uncveCmd.Flags().StringP("cvedbPath", "c", "/tmp/nvdbtools/cvedb", "scanner中提取的cvedb文件路径")
	uncveCmd.Flags().StringP("unzipPath", "u", "/tmp/nvdbtools/cvedbsrc/", "cvedb中提取文件保存路径")
}
