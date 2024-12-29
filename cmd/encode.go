/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"z2zloader/tools"
)

// encodeCmd represents the encode command
var encodeCmd = &cobra.Command{
	Use:   "encode ",
	Short: "aes加密模块",
	Long: `使用base64+aes加密shellcode For example:
.\z2zloader encode -f file_path.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		tools.Encode(file)
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().StringP("file", "f", "", "加密本地shellcode文件")
	encodeCmd.MarkFlagRequired("file")
}
