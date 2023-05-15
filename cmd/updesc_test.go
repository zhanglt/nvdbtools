package cmd

import (
	"os"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// 准备测试数据

	srcFilePath := "/tmp/source/file.txt"
	dstFilePath := "/tmp/destination/file.txt"
	// 创建源文件
	srcFile, err := os.Create(srcFilePath)
	if err != nil {
		t.Fatalf("无法创建源文件: %s", err)
	}
	defer srcFile.Close()

	// 写入源文件的内容
	data := []byte("这是源文件的内容")
	_, err = srcFile.Write(data)
	if err != nil {
		t.Fatalf("无法写入源文件: %s", err)
	}

	// 调用被测试的函数
	_, err = CopyFile(dstFilePath, srcFilePath)
	if err != nil {
		t.Fatalf("复制文件时发生错误: %s", err)
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(dstFilePath); os.IsNotExist(err) {
		t.Fatalf("复制后的目标文件不存在")
	}

	// 打开目标文件并读取内容
	dstFile, err := os.Open(dstFilePath)
	if err != nil {
		t.Fatalf("无法打开目标文件: %s", err)
	}
	defer dstFile.Close()

	// 读取目标文件的内容
	dstData := make([]byte, len(data))
	_, err = dstFile.Read(dstData)
	if err != nil {
		t.Fatalf("无法读取目标文件: %s", err)
	}

	// 检查源文件和目标文件的内容是否一致
	if string(dstData) != string(data) {
		t.Fatalf("复制后的文件内容与源文件不一致")
	}

	// 清理测试数据
	err = os.Remove(srcFilePath)
	if err != nil {
		t.Fatalf("无法删除源文件: %s", err)
	}

	err = os.Remove(dstFilePath)
	if err != nil {
		t.Fatalf("无法删除目标文件: %s", err)
	}
}
