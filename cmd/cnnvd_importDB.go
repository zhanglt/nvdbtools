/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/cnnvd"
)

// insterDBCmd represents the insterDB command
var importDBCmd = &cobra.Command{
	Use:   "importDB",
	Short: "cnnvd xml数据导入到sqlite数据库",
	Long:  `cnnvd xml数据导入到sqlite数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath, err := cmd.Flags().GetString("filePath")
		if err != nil {
			log.Println("请输入filePath：", filePath)
		}
		// 获取数据库对象
		db, err := cnnvd.GetDB()
		if err != nil {
			log.Println("获取sqlite数据库错误", err)
			return

		}
		// cve写入数据库及
		cnnvd.BuildCVE(filePath, db)
		fmt.Println("insterDB called")
	},
}

func init() {
	cnnvdCmd.AddCommand(importDBCmd)
	cnnvdCmd.Flags().StringP("filePath", "f", "", "cnnvd xml数据文件存放目录")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insterDBCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insterDBCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
