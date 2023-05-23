/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cnnvd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/cnnvd"
)

var rex string = "s/<vuln-descript>/<vuln-descript><![CDATA[/g;s%</vuln-descript>%]]></vuln-descript>%g;s/<name>/<name><![CDATA[/g;s%</name>%]]></name>%g"

// getxmlCmd represents the getxml command
var getxmlCmd = &cobra.Command{
	Use:   "getxml",
	Short: "cnnvd数据下载",
	Long:  `从cnnvd官网下载xml数据文件`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取token参数
		token, err := cmd.Flags().GetString("token")
		if err != nil || token == "" {
			log.Println("请输入正确token:", err)
			return

		}
		// xml文件保存路径
		savePath, err := cmd.Flags().GetString("savePath")
		if err != nil {
			log.Println("请输入正确filePath:", err)
			return
		}

		if err != nil {
			log.Println("输入正确下载超时参数（秒）", err)
			return
		}
		// 初始化路径
		os.RemoveAll(savePath)
		os.MkdirAll(savePath, 0755)

		// 获取cve文件的ID列表
		list, err := cnnvd.GetIDlist()
		if err != nil {
			log.Println("获取cnnvd ID列表错误:", err)
			return

		}
		ch := make(chan string)
		for _, id := range list {

			go func(fid, token, spath string) {
				// 下载xml文件
				ch <- cnnvd.GetXml(fid, token, spath)
			}(id, token, savePath)

		}
		timeout := time.After(900 * time.Second)
		for idx := 0; idx < len(list); idx++ {
			select {
			case res := <-ch:
				nt := time.Now().Format("2006-01-02 15:04:05")
				// 处理文件中的特殊字符
				command := exec.Command("sed", "-i", rex, savePath+res+".xml")
				_, err := command.CombinedOutput()
				if err != nil {
					fmt.Printf("[%s]文件%s%s.xml预处理错误:%s\n", nt, savePath, res, err)
				}
				fmt.Printf("[%s]完成下载：%s%s.xml\n", nt, list[idx], res)
			case <-timeout:
				fmt.Println("超时...:", list[idx]+".xml")
				break
			}
		}

		fmt.Println("cnnvd xml 全部数据文件下载完成")
	},
}

func init() {
	CnnvdCmd.AddCommand(getxmlCmd)
	getxmlCmd.Flags().StringP("token", "t", "", "cnnvd官网登陆后获取的token字符串")
	getxmlCmd.Flags().StringP("savePath", "s", "/tmp/nvdbtools/xml/", "cnnvd xml数据库的保存目录")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getxmlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getxmlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
