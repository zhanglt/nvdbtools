/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zhanglt/nvdbtools/common"
)

func main() {
	//ReadlineValid("/opt/nv/nvdbtools/dbsrc/apps.tb")
	var nvCveDB = "/tmp/cvedb"                      //从scanner中提取的cvedb文件路径
	var srcPath string = "/opt/nv/nvdbtools/dbsrc/" //cvedb 解压后的存放目录
	var targetPath string = "/tmp"                  //存放重新压缩打包文件的目录
	if err := common.UNzipDb(nvCveDB, srcPath); err == nil {
		log.Println("数据库解压完成")
	} else {
		log.Println("数据库解压失败")
		os.Exit(2)
	}
	db, err := memdbOpen(targetPath)
	if err != nil {
		log.Fatalln("数据库初始化错误", err)
		os.Exit(2)
	}

	if db.CreateDb("3.086", srcPath) {
		log.Println("cvedb重新打包成功")
	} else {
		log.Fatal("cvedb重新打包失败")
	}
	/*
			var cmdPull = &cobra.Command{
				Use:   "pull [OPTIONS] NAME[:TAG|@DIGEST]",
				Short: "Pull an image or a repository from a registry",
				Run: func(cmd *cobra.Command, args []string) {
					fmt.Println("Pull: " + strings.Join(args, " "))
					db, err := memdbOpen("/tmp/")
					if err != nil {
						os.Exit(2)
					}
					db.CreateDb("3.8.6")
				},
			}

		var rootCmd = &cobra.Command{}
		rootCmd.AddCommand(cmdPull)
		rootCmd.Execute()
		//cmd.Execute()
	*/
}
