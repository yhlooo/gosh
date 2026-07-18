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

		writeExtraPrompt := ctl.commandCollector.State() == term.OutputPrompt &&
			!ctl.wroteExtraPrompt &&
			c != '\r' && c != '\n'

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

		if writeExtraPrompt {
			ctl.writeExtraPrompt()
		}
		if _, err := ctl.output.Write(writeData); err != nil {
			return i, err
		}
		if ctl.deferOutput.Len() > 0 {
			_, _ = ctl.output.Write(ctl.deferOutput.Bytes())
			ctl.deferOutput.Reset()
		}
	}

	return len(p), nil
}

// ParseHandler 返回解析 ANSI 序列处理器
func (ctl *ShellOutputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{
		HandleOsc: ctl.handleOSC,
	}
}

// handleOSC 处理 OSC 序列
func (ctl *ShellOutputHandler) handleOSC(cmd int, data []byte) {
	switch cmd {
	case 133:
		if len(data) < 5 {
			return
		}
		switch string(data[:5]) {
		case "133;A":
			// 提示符开始
			ctl.wroteExtraPrompt = false
		case "133;B":
			// 命令开始
			if !ctl.wroteExtraPrompt {
				ctl.writeExtraPrompt()
			}
		case "133;C":
			// 命令开始执行
		case "133;D":
			// 命令执行结束
		}
	}
}

// writeExtraPrompt 写额外的输入提示符
func (ctl *ShellOutputHandler) writeExtraPrompt() {
	_, _ = ctl.output.WriteString(ctl.opts.Prompt)
	ctl.wroteExtraPrompt = true
}
