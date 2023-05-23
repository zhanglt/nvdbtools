/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cve

import (
	"github.com/spf13/cobra"
)

// cveCmd represents the cvd command
var CveCmd = &cobra.Command{
	Use:   "cve",
	Short: "实现scanner的cvedb数据库相关操作",
	Long: `实现scanner的cvedb数据库相关操作包三个命令
	1、解压cvedb数据库;
	2、更新解压后的数据;
	3、重新做数据打包成cvedb`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	CveCmd.AddCommand(rebuildCmd)
	CveCmd.AddCommand(unzipCmd)
	CveCmd.AddCommand(updateCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cvdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cvdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
