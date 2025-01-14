/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"z2zloader/tools"
)

func main() {
	//tools.OriginalLoader("./output.bin")
	//cmd.Execute()
	//tools.ThreadShellcodeInject("./output.bin")
	//tools.ApcShellcodeInject("./output.bin")

	tools.EarlybirlInject("./output.bin")
	//tools.MappingInject("E:\\mygo\\z2zloader\\output.bin")
}
