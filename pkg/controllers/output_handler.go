package controllers

import (
	"io"

	"github.com/charmbracelet/x/ansi"

	"github.com/yhlooo/gosh/pkg/term"
)

// ShellOutputHandler 返回作为输出处理器的控制器
func (ctl *Controller) ShellOutputHandler() *ShellOutputHandler {
	return (*ShellOutputHandler)(ctl)
}

// ShellOutputHandler 输出处理器
type ShellOutputHandler Controller

var _ io.Writer = (*ShellOutputHandler)(nil)

// Write 处理 pty 输出的内容
func (ctl *ShellOutputHandler) Write(p []byte) (n int, err error) {
	ctl.outputLock.Lock()
	defer ctl.outputLock.Unlock()

	for i, c := range p {
		ctl.outputParser.Advance(c)

		writeData := []byte{c}
		if _, err := ctl.commandCollector.Write(writeData); err != nil {
			return i, err
		}

		ccState := ctl.commandCollector.State()
		switch ccState {
		case term.OutputOthers, term.OutputPrompt, term.OutputCommand:
			ctl.inExec = false
		case term.OutputCommandExec:
			ctl.inExec = true
		}

		if !ctl.ready {
			if ccState != term.OutputOthers {
				// 首次检测到 OSC133 切换状态后说明 shell 已就绪
				ctl.ready = true
			}
			// 未就绪期间隐藏输出，转为就绪的第一个字符也忽略
			continue
		}
		if ccState == term.OutputPrompt && (*Controller)(ctl).State() != Shell {
			// Agent 输出期间隐藏 prompt
			continue
		}
		if _, err := ctl.output.Write(writeData); err != nil {
			return i, err
		}
	}

	return len(p), nil
}

// ParseHandler 返回解析 ANSI 序列处理器
func (ctl *ShellOutputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{}
}
