package controllers

import (
	"io"

	"github.com/charmbracelet/x/ansi"

	"github.com/yhlooo/gosh/pkg/term"
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
func (ctl *OutputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{}
}
