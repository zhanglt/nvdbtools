/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package main

import (
	"database/sql"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zhanglt/nvdbtools/common"
)

func main() {

	//ReadlineValid("/opt/nv/nvdbtools/dbsrc/apps.tb")
	var nvCveDB = "/tmp/nvdbtools/cvedb"     //从scanner中提取的cvedb文件路径
	var srcPath = "/tmp/nvdbtools/cvedbsrc/" //cvedb 解压后的存放目录
	var dbPath = "/tmp/nvdbtools/cvedbtemp/"
	var targetPath = "/tmp/nvdbtools/cvedbtarget/" //存放重新压缩打包文件的目录
	//os.MkdirAll("/tmp/nvdbtools/cvedbsrc/", 0766)
	//os.MkdirAll("/tmp/nvdbtools/cvedbtemp/", 0766)
	//os.MkdirAll("/tmp/nvdbtools/cvedbtarget/", 0766)

	DB, err := sql.Open("sqlite3", "cnvd20230428.db")
	if err != nil {
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		return
	}
	//--------------------解压获取原数据--------------------------------

	if err := common.UNzipDb(nvCveDB, srcPath); err == nil {
		log.Println("数据库解压完成")
	} else {
		log.Println("数据库解压失败")
		os.Exit(2)
	}
	//os.RemoveAll("/tmp/nvdbtools")

	//------------------------数据处理-----------------------------
	src := []string{"alpine_full.tb", "amazon_full.tb", "centos_full.tb", "debian_full.tb", "mariner_full.tb", "oracle_full.tb", "suse_full.tb", "ubuntu_full.tb"}

	for _, file := range src {
		common.UpdateDescription(srcPath+file, targetPath+file, "", DB)
	}
	common.UpdateDescription(srcPath+"apps.tb", targetPath+"apps.tb", "apps", DB)

	//------------------重新打包成sanner所需的cvedb数据文件-----------

	kver, _, _ := common.GetDbVersion(nvCveDB)

	db, err := memdbOpen(dbPath)
	if err != nil {
		log.Fatalln("数据库初始化错误", err)
		os.Exit(2)
	}
	defer db.Close()
	if db.RewriteDb(fmt.Sprintf("%f", kver), srcPath) {
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
