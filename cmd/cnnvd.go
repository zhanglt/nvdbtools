/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cnnvdCmd represents the cnnvd command
var cnnvdCmd = &cobra.Command{
	Use:   "cnnvd",
	Short: "cnnvd数据的下载及入库",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cnnvd")

	},
}

func init() {
	rootCmd.AddCommand(cnnvdCmd)
	// 获取token参数

}
