/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

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
			log.Println("请输入filePath:", filePath)
		}
		// 获取数据库对象
		db, err := cnnvd.GetDB()
		if err != nil {
			log.Println("获取sqlite数据库错误", err)
			return

		}
		//获取XML文件名称列表
		xmlList, err := ioutil.ReadDir(filePath)
		if err != nil {
			log.Fatal(err)
		}
		times := len(xmlList)
		//创建一个互斥锁数组 多一个给主协程用
		var cc = make([]*sync.Mutex, times+1)
		//往数组中塞入互斥锁，默认直接加锁
		for i := 0; i < len(cc); i++ {
			m := &sync.Mutex{}
			m.Lock()
			cc[i] = m
		}
		for i := range xmlList {
			//创建子协程
			go func(index int, dbName *sql.DB) {
				//子协程尝试为数组中对应 index 位置的锁加锁，获取不到锁就等待
				//因为初始化的这些互斥锁默认就已经被锁住了，所以这里创建的子协程都会被阻塞
				//一旦获取到锁，就执行逻辑，最后将当前index的锁和index+1的锁释放，这样正在等待 index +1 位置的锁的子协程就可以继续执行了
				cc[index].Lock()
				fmt.Printf("this value is %d \n", index)
				cnnvd.BuildCVE(filePath+xmlList[index].Name(), dbName)
				cc[index].Unlock()
				cc[index+1].Unlock()
			}(i, db)
		}
		//将index 为 0 位置的锁解锁，让第一个子协程可以继续执行
		cc[0].Unlock()
		//为 index 为 times 的锁加锁，只有当最后一个子协程执行完毕后，这个锁才会解锁，主协程才能继续向下走
		cc[times].Lock()
		cc[times].Unlock()

	},
}

func init() {
	rootCmd.AddCommand(importDBCmd)
	importDBCmd.Flags().StringP("filePath", "f", "/tmp/nvdbtools/xml/", "cnnvd xml数据文件存放目录")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insterDBCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insterDBCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
