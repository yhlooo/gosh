package ui

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
)

// NewInputBox 创建输入框
func NewInputBox(echoWriter io.Writer) *InputBox {
	ib := &InputBox{
		echoWriter: echoWriter,
		content:    [][]rune{nil},
	}
	ib.parser = ansi.NewParser()
	ib.parser.SetHandler(ib.ansiHandler())
	return ib
}

// InputBox 输入框
type InputBox struct {
	lock       sync.Mutex
	echoWriter io.Writer
	parser     *ansi.Parser

	content   [][]rune
	rowCursor int
	colCursor int
}

var _ io.Writer = (*InputBox)(nil)

// Write 往输入框写内容
func (ib *InputBox) Write(p []byte) (n int, err error) {
	ib.lock.Lock()
	defer ib.lock.Unlock()

	for _, c := range p {
		ib.parser.Advance(c)
	}

	return len(p), nil
}

// Activate 激活
func (ib *InputBox) Activate() {
	if ib.echoWriter == nil {
		return
	}

	// 发送 rmkx 和输入提示
	_, _ = ib.echoWriter.Write([]byte("\r\n\x1b[?1l\x1b>\x1b[0;34m> "))
}

// Deactivate 停用
func (ib *InputBox) Deactivate() {
	if ib.echoWriter == nil {
		return
	}

	// 关闭输入提示
	_, _ = ib.echoWriter.Write([]byte("\x1b[0m\r\n"))
}

// Content 返回输入内容
func (ib *InputBox) Content() string {
	ib.lock.Lock()
	defer ib.lock.Unlock()

	ret := &strings.Builder{}
	for _, line := range ib.content {
		ret.WriteString(string(line) + "\n")
	}

	return strings.TrimSuffix(ret.String(), "\n")
}

// Reset 清空缓冲内容，重置状态
func (ib *InputBox) Reset() {
	ib.lock.Lock()
	defer ib.lock.Unlock()

	ib.content = [][]rune{nil}
	ib.rowCursor = 0
	ib.colCursor = 0
}

// ansiHandler 返回 ANSI 序列处理器
func (ib *InputBox) ansiHandler() ansi.Handler {
	return ansi.Handler{
		Print:     ib.handlePrint,
		Execute:   ib.handleExecute,
		HandleCsi: ib.handleCSI,
	}
}

// handlePrint 控制打印字符
func (ib *InputBox) handlePrint(r rune) {
	ib.insert(r)
}

// handleExecute 处理控制字符
func (ib *InputBox) handleExecute(b byte) {
	switch b {
	case SOH: // 行首 Ctrl+A
		ib.toLineStart()
	case STX: // 左移 Ctrl+B
		ib.moveLeft(1)
	case ENQ: // 行尾 Ctrl+E
		ib.toLineEnd()
	case ACK: // 右移 Ctrl+F
		ib.moveRight(1)
	case BS: // 退格 Ctrl+H
		ib.backspace()
	case LF, CR: // 回车 Enter / Ctrl+M / Ctrl+J
		ib.insertNewLine()
	case HT: // 制表符 Tab / Ctrl+I
		ib.insert('\t')
	case DEL:
		ib.backspace()
	}
}

// handleCSI 处理 CSI 序列
func (ib *InputBox) handleCSI(cmd ansi.Cmd, params ansi.Params) {
	switch cmd.Final() {
	case 'A': // 上移 Up
		n := 1
		if len(params) == 1 {
			n = int(params[0])
		}
		ib.moveUp(n)
	case 'B': // 下移 Down
		n := 1
		if len(params) == 1 {
			n = int(params[0])
		}
		ib.moveDown(n)
	case 'C': // 右移 Right
		n := 1
		if len(params) == 1 {
			n = int(params[0])
		}
		ib.moveRight(n)
	case 'D': // 左移 Left
		n := 1
		if len(params) == 1 {
			n = int(params[0])
		}
		ib.moveLeft(n)
	case 'H': // 行首 Home
		ib.toLineStart()
	case 'F': // 行尾 End
		ib.toLineEnd()
	case '~':
		if len(params) == 1 && params[0] == 3 {
			// Delete
			ib.delete()
		}
	}
}

// moveUp 光标上移
func (ib *InputBox) moveUp(n int) {
	if n > ib.rowCursor {
		n = ib.rowCursor
	}
	if n <= 0 {
		return
	}

	ib.rowCursor -= n

	oldCol := ib.colCursor
	if ib.colCursor > len(ib.content[ib.rowCursor]) {
		ib.colCursor = len(ib.content[ib.rowCursor])
	}

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dA", n)))
		if oldCol > ib.colCursor {
			_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dD", oldCol-ib.colCursor)))
		}
	}
}

// moveDown 光标下移
func (ib *InputBox) moveDown(n int) {
	if n > len(ib.content)-ib.rowCursor-1 {
		n = len(ib.content) - ib.rowCursor - 1
	}
	if n <= 0 {
		return
	}

	ib.rowCursor += n

	oldCol := ib.colCursor
	if ib.colCursor > len(ib.content[ib.rowCursor]) {
		ib.colCursor = len(ib.content[ib.rowCursor])
	}

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dB", n)))
		if oldCol > ib.colCursor {
			_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dD", oldCol-ib.colCursor)))
		}
	}
}

// moveRight 光标右移
func (ib *InputBox) moveRight(n int) {
	if n > len(ib.content[ib.rowCursor])-ib.colCursor {
		n = len(ib.content[ib.rowCursor]) - ib.colCursor
	}
	if n <= 0 {
		return
	}

	ib.colCursor += n

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dC", n)))
	}
}

// moveLeft 光标左移
func (ib *InputBox) moveLeft(n int) {
	if n > ib.colCursor {
		n = ib.colCursor
	}
	if n <= 0 {
		return
	}

	ib.colCursor -= n

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dD", n)))
	}
}

// toLineStart 光标移动到行首
func (ib *InputBox) toLineStart() {
	n := ib.colCursor
	if n <= 0 {
		return
	}

	ib.colCursor = 0
	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dD", n)))
	}
}

// toLineEnd 光标移动到行尾
func (ib *InputBox) toLineEnd() {
	n := len(ib.content[ib.rowCursor]) - ib.colCursor
	if n <= 0 {
		return
	}

	ib.colCursor = len(ib.content[ib.rowCursor])
	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[%dC", n)))
	}
}

// insert 插入字符
func (ib *InputBox) insert(r rune) {
	before := ib.content[ib.rowCursor][:ib.colCursor]
	after := ib.content[ib.rowCursor][ib.colCursor:]
	ib.content[ib.rowCursor] = append(before, append([]rune{r}, after...)...)
	ib.colCursor++

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte("\x1b[1@" + string(r)))
	}
}

// insertNewLine 插入新行
func (ib *InputBox) insertNewLine() {
	newLine := ib.content[ib.rowCursor][ib.colCursor:]
	ib.content[ib.rowCursor] = ib.content[ib.rowCursor][:ib.colCursor]
	beforeLines := ib.content[:ib.rowCursor+1]
	afterLines := ib.content[ib.rowCursor+1:]
	ib.content = append(beforeLines, append([][]rune{newLine}, afterLines...)...)

	ib.rowCursor++
	ib.colCursor = 0

	if ib.echoWriter != nil {
		if ib.rowCursor == len(ib.content)-1 {
			_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[K\r\n> %s\r> ", string(newLine))))
		} else {
			followingLines := len(ib.content) - ib.rowCursor - 1
			_, _ = ib.echoWriter.Write([]byte(fmt.Sprintf("\x1b[K\x1b[%dB\r\n\x1b[%dA\x1b[L> %s\r> ", followingLines, followingLines, string(newLine))))
		}
	}
}

// backspace 删除光标前一个字符
func (ib *InputBox) backspace() {
	if ib.colCursor == 0 {
		return
	}
	line := ib.content[ib.rowCursor]
	line = append(line[:ib.colCursor-1], slices.Clone(line[ib.colCursor:])...)
	ib.content[ib.rowCursor] = line
	ib.colCursor--

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte("\b\x1b[1P"))
	}
}

// delete 删除光标后一个字符
func (ib *InputBox) delete() {
	line := ib.content[ib.rowCursor]
	if ib.colCursor == len(line) {
		return
	}

	line = append(line[:ib.colCursor], slices.Clone(line[ib.colCursor+1:])...)
	ib.content[ib.rowCursor] = line

	if ib.echoWriter != nil {
		_, _ = ib.echoWriter.Write([]byte("\x1b[1P"))
	}
}
