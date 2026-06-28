package controllers

import (
	"io"

	"github.com/charmbracelet/x/ansi"
)

// OutputHandler 返回作为输出处理器的控制器
func (ctl *Controller) OutputHandler() *OutputHandler {
	return (*OutputHandler)(ctl)
}

// OutputHandler 输出处理器
type OutputHandler Controller

var _ io.Writer = (*OutputHandler)(nil)

// Write 处理 pty 输出的内容
func (ctl *OutputHandler) Write(p []byte) (n int, err error) {
	ctl.outputLock.Lock()
	defer ctl.outputLock.Unlock()

	for i, c := range p {
		ctl.outputParser.Advance(c)
		if _, err := ctl.output.Write([]byte{c}); err != nil {
			return i, err
		}
	}

	return len(p), nil
}

// ParseHandler 返回解析 ANSI 序列处理器
func (ctl *OutputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{}
}
