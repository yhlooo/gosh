package controllers

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// InputState 输入模式
type InputState uint32

const (
	// InputToShell 输入到 shell
	InputToShell InputState = iota
	// InputToAgent 输入到 Agent
	InputToAgent
	// InputInterruptAgent 输入中断 Agent
	InputInterruptAgent
)

// InputHandler 返回作为输入处理器的控制器
func (ctl *Controller) InputHandler() *InputHandler {
	return (*InputHandler)(ctl)
}

// InputHandler 输入处理器
type InputHandler Controller

var _ io.Writer = (*InputHandler)(nil)

// Write 处理写到 pty 的输入内容
func (ctl *InputHandler) Write(p []byte) (n int, err error) {
	ctl.inputLock.Lock()
	defer ctl.inputLock.Unlock()

	curMode := ctl.inputState
	buff := make([]byte, 0, len(p))
	for i, c := range p {
		ctl.inputParser.Advance(c)
		if ctl.inputParser.State() != parser.GroundState {
			// 非 Ground 态，可能在输入切换模式序列，内容暂时缓冲
			buff = append(buff, c)
			continue
		}

		if curMode != ctl.inputState {
			// 切换了模式，丢弃缓冲区
			curMode = ctl.inputState
			buff = buff[:0]
			continue
		}

		// 回到 Ground 态，没有切换模式，把缓冲区刷掉
		if len(buff) > 0 {
			if _, err := ctl.writeUpstream(buff); err != nil {
				return i, err
			}
			buff = buff[:0]
		}

		if _, err := ctl.writeUpstream([]byte{c}); err != nil {
			return i, err
		}
	}

	return len(p), nil
}

// writeUpstream 写输入到上游
func (ctl *InputHandler) writeUpstream(p []byte) (n int, err error) {
	switch ctl.inputState {
	case InputToShell:
		return ctl.ptmx.Write(p)
	case InputToAgent:
		return ctl.agentInputBox.Write(p)
	case InputInterruptAgent:
		// 此时没有上游，忽略输入
		return len(p), nil
	default:
		return 0, fmt.Errorf("unknown input mode: %d", ctl.inputState)
	}
}

// ParseHandler 返回解析 ANSI 序列处理器
func (ctl *InputHandler) ParseHandler() ansi.Handler {
	return ansi.Handler{
		Execute:   ctl.handleExecute,
		HandleCsi: ctl.handleCSI,
	}
}

// handleExecute 处理控制字符
func (ctl *InputHandler) handleExecute(b byte) {
	switch {
	case b == '\r' && ctl.inputState == InputToAgent:
		// Enter 提交 Prompt 到 Agent

		content := ctl.agentInputBox.Content()
		ctl.agentInputBox.Reset()

		ctl.agentInputBox.Deactivate()
		ctl.inputState = InputInterruptAgent

		// 发送指令给 Agent
		go func() {
			ctl.logger.Info(fmt.Sprintf("send to agent: %q", content))
			if err := ctl.agent.Chat(ctl.ctx, content); err != nil {
				ctl.logger.Error(err, "chat with agent error")
				_, _ = ctl.output.Write([]byte(fmt.Sprintf(
					"\r\n\x1b[31m%s\x1b[0m",
					strings.ReplaceAll(err.Error(), "\n", "\r\n"),
				)))
			}

			ctl.inputLock.Lock()
			ctl.inputState = InputToAgent
			ctl.agentInputBox.Activate()
			ctl.inputLock.Unlock()
		}()

	case b == '\x03' && ctl.inputState == InputInterruptAgent:
		// Ctrl+C 中断 Agent
		if err := ctl.agent.Cancel(); err != nil {
			ctl.logger.Error(err, "cancel agent error")
		}

	case b == '\x03' && ctl.inputState == InputToAgent:
		ctl.inputState = InputToShell
		ctl.agentInputBox.Deactivate()
		ctl.agentInputBox.Reset()
	}
}

// handleCSI 处理 CSI 序列
func (ctl *InputHandler) handleCSI(cmd ansi.Cmd, _ ansi.Params) {
	// Shift + Tab 切换输入模式
	if cmd.Final() == 'Z' {
		switch ctl.inputState {
		case InputToShell:
			ctl.inputState = InputToAgent
			ctl.agentInputBox.Activate()
		case InputToAgent:
			ctl.inputState = InputToShell
			ctl.agentInputBox.Deactivate()
			ctl.agentInputBox.Reset()
		default:
		}
	}
}
