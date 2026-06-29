package controllers

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// OutputState 输出状态
type OutputState uint32

const (
	// OutputOthers 输出其它（无需关心的）内容
	OutputOthers OutputState = iota
	// OutputPrompt 输出提示符
	OutputPrompt
	// OutputCommand 输出命令内容
	OutputCommand
	// OutputCommandExec 输出命令执行输出
	OutputCommandExec
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

	lastGround := 0
	for i, c := range p {
		ctl.outputParser.Advance(c)

		if _, err := ctl.output.Write([]byte{c}); err != nil {
			return i, err
		}
		if ctl.outputParser.State() == parser.GroundState {
			switch ctl.outputState {
			case OutputOthers:
			case OutputPrompt:
				ctl.promptBuff.Write(p[lastGround : i+1])
			case OutputCommand:
				_, _ = ctl.commandBuff.Write(p[lastGround : i+1])
			case OutputCommandExec:
				_, _ = ctl.outputLog.Write(p[lastGround : i+1])
			}
			lastGround = i + 1
		}
	}

	return len(p), nil
}

// ParseHandler 返回解析 ANSI 序列处理器
func (ctl *OutputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{
		HandleOsc: ctl.handleOSC,
	}
}

// handleOSC 处理 OSC 序列
func (ctl *OutputHandler) handleOSC(cmd int, data []byte) {
	switch cmd {
	case 1:
	case 2:
	case 7:
	case 133:
		if len(data) < 5 {
			return
		}
		switch string(data[:5]) {
		case "133;A":
			// 提示符开始
			ctl.outputState = OutputPrompt
			ctl.promptBuff.Reset()
		case "133;B":
			// 命令开始
			ctl.outputState = OutputCommand
			ctl.commandBuff.ResetState()
		case "133;C":
			// 命令开始执行
			ctl.outputState = OutputCommandExec
			_, _ = ctl.outputLog.WriteString("\n---------------- Command Start ----------------\n")
			_, _ = ctl.outputLog.WriteString(ctl.commandBuff.String())
			_, _ = ctl.outputLog.WriteString("\n---------------- Command Executed ----------------\n")
		case "133;D":
			// 命令执行结束
			ctl.outputState = OutputOthers
			exitCode := 0
			dataDivided := strings.Split(string(data), ";")
			if len(dataDivided) >= 3 {
				exitCode, _ = strconv.Atoi(dataDivided[2])
			}
			_, _ = ctl.outputLog.WriteString(fmt.Sprintf(
				"\n---------------- Command Finished (%d) ----------------\n",
				exitCode,
			))
		}
	case 1337:
	}
}
