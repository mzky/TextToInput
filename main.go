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
	VK_ESCAPE         = 0x1B
	VK_RETURN         = 0x0D // 添加回车键的虚拟键码
	VK_CONTROL        = 0x11 // 添加Ctrl键的虚拟键码
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
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// 添加通用的按键发送函数
func sendKey(vk uint16, unicode uint16) {
	input := INPUT{
		Type: INPUT_KEYBOARD,
		Ki: KEYBDINPUT{
			WVk:   vk,
			WScan: unicode,
			DwFlags: func() uint32 {
				if vk == 0 {
					return KEYEVENTF_UNICODE
				}
				return 0
			}(),
		},
	}

	// 按下和释放按键
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
	input.Ki.DwFlags |= KEYEVENTF_KEYUP
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), unsafe.Sizeof(input))
}

// Unicode字符发送函数
func sendUnicodeChar(ch rune) {
	sendKey(0, uint16(ch))
}

// 换行函数
func sendNewLine() {
	sendKey(VK_RETURN, 0)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func isEscapePressed() bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(VK_ESCAPE))
	return ret&0x8000 != 0
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

	// 读取文件并处理跨平台换行
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("扫描文件失败: %v\n", err)
		return
	}

	totalLines := len(lines)

	fmt.Println("按 ESC 键可以随时终止输入")
	fmt.Println("请将光标放到要输入的位置...")
	for i := 5; i > 0; i-- {
		fmt.Printf("\r%d秒后开始输入", i)
		time.Sleep(time.Second)
	}
	fmt.Println("\n开始输入...")

	for lineNum, line := range lines {
		if isEscapePressed() {
			fmt.Println("\n用户终止输入")
			return
		}

		fmt.Printf("\r正在输入第 %d/%d 行... ", lineNum+1, totalLines)

		for _, r := range line {
			if isEscapePressed() {
				fmt.Println("\n用户终止输入")
				return
			}
			sendUnicodeChar(r)
			delay := time.Duration(rand.Float64()*30+20) * time.Millisecond
			time.Sleep(delay)
		}
		sendNewLine()

		// 每10行暂停一次，让作家助手计算字数
		if (lineNum+1)%10 == 0 && lineNum < totalLines-1 {
			time.Sleep(time.Second)
		}
	}

	fmt.Println("\n输入完成!")
}
