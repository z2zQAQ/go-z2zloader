package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func CreateGoFile(goFile string, exeFile string, funcContent string) error {
	// 构建新Go文件的路径，这里假设是与main函数所在目录同级的名为new_file.go的文件，可根据实际调整
	funcContent = strings.Replace(funcContent, "\\", "\\\\", -1)
	goFile = "output/" + goFile
	exeFile = "output/" + exeFile
	// 要写入新Go文件的内容，这里是示例的函数a以及引入库的内容，可按需修改
	fileContent := fmt.Sprintf(`package main

import (
    "z2zloader/tools"
)
func main(){
	%s
}

`, funcContent)
	// 创建或覆盖写入文件

	err := ioutil.WriteFile(goFile, []byte(fileContent), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 构建编译命令，这里假设使用go build命令进行编译，具体命令可根据实际情况调整，比如添加更多编译参数等
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", exeFile, goFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行编译命令
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("编译文件失败: %v", err)
	}

	return nil
}
