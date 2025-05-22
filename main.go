package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
)

const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_UNICODE = 0x0004
	KEYEVENTF_KEYUP   = 0x0002
)

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	Type    uint32
	Ki      KEYBDINPUT
	Padding uint64 // 保证64位系统下结构体对齐
}

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

func sendUnicodeChar(ch rune) {
	input := INPUT{
		Type: INPUT_KEYBOARD,
		Ki: KEYBDINPUT{
			WVk:     0,
			WScan:   uint16(ch),
			DwFlags: KEYEVENTF_UNICODE,
		},
	}

	// 按下按键
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))

	// 释放按键
	input.Ki.DwFlags |= KEYEVENTF_KEYUP
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
}

func main() {
	// 创建文件选择对话框
	dlg := new(walk.FileDialog)

	// 设置对话框属性
	dlg.Title = "选择要读取的文本文件"
	dlg.Filter = "所有文件 (*.*)|*.*"

	// 显示文件选择对话框
	if ok, err := dlg.ShowOpen(nil); err != nil {
		fmt.Printf("打开文件对话框失败: %v\n", err)
		return
	} else if !ok {
		fmt.Println("未选择文件")
		return
	}

	// 获取选择的文件路径
	filePath := dlg.FilePath

	// 读取文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	// 给用户5秒时间将光标放到目标位置
	fmt.Println("请将光标放到要输入的位置...")
	for i := 5; i > 0; i-- {
		fmt.Printf("\r%d秒后开始输入", i)
		time.Sleep(time.Second)
	}
	fmt.Println("\n开始输入...")

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 将字符串转换为rune切片以正确处理中文字符
		runes := []rune(line)
		// 逐字符输入
		for _, r := range runes {
			sendUnicodeChar(r)
			// 生成0.05~0.5秒的随机延迟时间
			delay := rand.Float64()*490 + 50 // 50~500毫秒
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		// 在每行结束时输入回车
		sendUnicodeChar('\r')
		sendUnicodeChar('\n')
		// 每行之间稍微多等待一下
		time.Sleep(time.Duration(rand.Float64()*500+500) * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件时发生错误: %v\n", err)
	}
}
