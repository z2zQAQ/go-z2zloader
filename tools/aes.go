package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
)

func bytestohex(originalBytes []byte) []string {
	newBytes := make([]byte, 0, len(originalBytes))
	for _, v := range originalBytes {
		newBytes = append(newBytes, v)
	}
	hexFormatBytes := make([]string, len(newBytes))
	for i, v := range newBytes {
		hexFormatBytes[i] = fmt.Sprintf("0x%02x", v)
	}
	return hexFormatBytes
}

func hexToByte(hexStr string) (byte, error) {
	// 去除十六进制字符串开头的 "0x"
	hexStr = hexStr[2:]

	// 将十六进制字符串转换为 uint64 类型
	num, err := strconv.ParseUint(hexStr, 16, 8)
	if err != nil {
		return 0, err
	}

	// 将 uint64 类型转换为 byte 类型
	return byte(num), nil
}

func hexStrToBytes(hexStrs []string) ([]byte, error) {
	result := make([]byte, len(hexStrs))
	for i, hexStr := range hexStrs {
		byteValue, err := hexToByte(hexStr)
		if err != nil {
			return nil, err
		}
		result[i] = byteValue
	}
	return result, nil
}

func aesEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	// 创建AES加密算法的block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 选择填充模式，这里使用CBC模式，需要一个初始向量IV
	iv := make([]byte, aes.BlockSize)
	// 这里可以使用更安全的随机生成IV的方式，为了简单示例就全初始化为0了
	for i := range iv {
		iv[i] = 0
	}
	// 按照CBC模式创建加密模式实例
	mode := cipher.NewCBCEncrypter(block, iv)
	// 对明文进行填充处理，以满足分组加密要求，Go标准库内部会按PKCS7填充
	paddedText := pkcs7Padding(plaintext, aes.BlockSize)
	// 创建一个用于存放密文的切片，长度和填充后的明文长度一致
	ciphertext := make([]byte, len(paddedText))
	// 执行加密操作
	mode.CryptBlocks(ciphertext, paddedText)
	return ciphertext, nil
}

// PKCS7填充函数，按照AES分组大小进行填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("输入数据为空")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, fmt.Errorf("无效的填充")
	}
	return data[:length-padding], nil
}

func aesDecode(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	for i := range iv {
		iv[i] = 0
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	return pkcs7Unpadding(plaintext)
}

func Encode(fp string) {
	//读文件 转换为16进制bytes
	file := fp
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
		return
	}
	// 定义AES加密密钥，长度必须是16、24或32字节，这里以16字节为例
	key := []byte("z2zQAQ1111111111")
	ciphertext, err := aesEncrypt(content, key)
	if err != nil {
		fmt.Println("加密出现错误:", err)
		return
	}
	base64Data := base64.StdEncoding.EncodeToString(ciphertext)
	encodeData := []byte(base64Data)
	//保存加密后的文件
	file_output := "output/encodeshellcode.bin"
	err = os.WriteFile(file_output, encodeData, 0644)
	if err != nil {
		log.Fatalf("保存文件时出错: %v", err)
	}
	log.Println("文件已成功保存，文件保存为output目录下的encodeshellcode.bin")
}

func Decode(encodeDataByte []byte) []byte {
	encodeData := string(encodeDataByte)
	content, err := base64.StdEncoding.DecodeString(encodeData)
	if err != nil {
		fmt.Println("base64解密出错")
	}
	// 定义AES加密密钥，长度必须是16、24或32字节，这里以16字节为例
	key := []byte("z2zQAQ1111111111")
	shellcode, err := aesDecode(content, key)
	if err != nil {
		fmt.Println("解密出现错误:", err)
		return nil
	}
	return shellcode
}
