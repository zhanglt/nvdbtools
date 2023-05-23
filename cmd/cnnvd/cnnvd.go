/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cnnvd

import (
	"github.com/spf13/cobra"
)

// cnnvdCmd represents the cnnvd command
var CnnvdCmd = &cobra.Command{
	Use:   "cnnvd",
	Short: "实现cnnvd数据相关操作",
	Long:  `实现cnnvd相关数据操作，包括两个自命令，1、从cnnvd官网下载xml历史数据，2、将xml数据导入到sqlite数据库`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cnnvdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cnnvdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
