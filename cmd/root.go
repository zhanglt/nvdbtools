/*
Copyright © 2023 NAME HERE <kitsdk@163.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanglt/nvdbtools/cmd/cnnvd"
	"github.com/zhanglt/nvdbtools/cmd/cve"
)

var RootCmd = &cobra.Command{
	Use:   "nvdbtools",
	Short: "nv数据库转换工具",
	Long:  "nv数据库转换工具",
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(cve.CveCmd)
	RootCmd.AddCommand(cnnvd.CnnvdCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nvdbtools.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action(cn(cn(cnvd202(ccreatedb28) is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
