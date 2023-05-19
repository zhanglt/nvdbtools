/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/cnnvd"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "cnnvd数据下载",
	Long:  `从cnnvd官网下载xml数据文件`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取token参数
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Println("请输入正确token:", err)
			return

		}
		// 获取cve文件的ID列表
		list, err := cnnvd.GetIDlist()
		if err != nil {
			log.Println("获取cnnvd ID列表错误:", err)
			return

		}
		ch := make(chan string)
		for _, id := range list {

			go func(fid, token string) {
				// 下载xml文件
				ch <- cnnvd.GetXml(fid, token)
			}(id, token)

		}
		timeout := time.After(900 * time.Second)
		for idx := 0; idx < len(list); idx++ {
			select {
			case res := <-ch:
				nt := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[%s]完成下载： %s\n", nt, res)
			case <-timeout:
				fmt.Println("超时...")
				break
			}
		}

		fmt.Println("cnnvd xml 全部数据文件下载完成")
	},
}

func init() {
	cnnvdCmd.AddCommand(downloadCmd)
	cnnvdCmd.Flags().StringP("token", "t", "", "cnnvd官网登陆后获取的token字符串")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
