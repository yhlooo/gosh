package controllers

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
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

	if !ctl.ready {
		// 未就绪期间忽略所有输入
		return len(p), nil
	}

	if ctl.inputInterceptor != nil {
		// 全部转发到拦截器处理
		return ctl.inputInterceptor.Write(p)
	}

	curState := (*Controller)(ctl).State()
	for i, c := range p {
		ctl.inputParser.Advance(c)

		if ctl.inputParser.State() != parser.GroundState {
			ctl.inputBuff.WriteByte(c)
			continue
		}

		if curState != (*Controller)(ctl).State() {
			// 输入序列切换了状态，丢弃缓冲区
			curState = (*Controller)(ctl).State()
			ctl.inputBuff.Reset()
			continue
		}

		// 回到 Ground 态，没有切换模式，把缓冲区刷掉
		_, err = ctl.writeUpstream(append(ctl.inputBuff.Bytes(), c))
		ctl.inputBuff.Reset()
		if err != nil {
			return i, err
		}
	}

	return len(p), nil
}

// writeUpstream 写输入到上游
func (ctl *InputHandler) writeUpstream(p []byte) (n int, err error) {
	state := (*Controller)(ctl).State()
	switch state {
	case Shell, Exec:
		return ctl.shellPtmx.Write(p)
	case AgentInput:
		return ctl.agentInputBox.Write(p)
	case AgentOutput:
		// 此时没有上游，忽略输入
		return ctl.agentPtmx.Write(p)
	default:
		return 0, fmt.Errorf("unknown input mode: %d", state)
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
	state := (*Controller)(ctl).State()

	switch {
	case b == '\r' && state == AgentInput:
		// Enter 提交 Prompt 到 Agent

		content := ctl.agentInputBox.Content()
		ctl.agentInputBox.Reset()

		ctl.agentInputBox.Deactivate()
		ctl.inAgentOutput = true

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
			ctl.inAgentOutput = false
			ctl.agentInputBox.Activate()
			ctl.inputLock.Unlock()
		}()

	case b == '\x03' && state == AgentOutput:
		// Ctrl+C 中断 Agent
		if err := ctl.agent.Cancel(); err != nil {
			ctl.logger.Error(err, "cancel agent error")
		}

	case b == '\x03' && state == AgentInput:
		ctl.inAgent = false
		ctl.agentInputBox.Deactivate()
		ctl.agentInputBox.Reset()
		_, _ = ctl.output.Write(ctl.commandCollector.CurrentPromptAndCommand())
	}
}

// handleCSI 处理 CSI 序列
func (ctl *InputHandler) handleCSI(cmd ansi.Cmd, _ ansi.Params) {
	// Shift + Tab 切换输入模式
	if cmd.Final() == 'Z' {
		switch (*Controller)(ctl).State() {
		case Shell:
			ctl.inAgent = true
			ctl.agentInputBox.Activate()
		case AgentInput:
			ctl.inAgent = false
			ctl.agentInputBox.Deactivate()
			ctl.agentInputBox.Reset()
			_, _ = ctl.output.Write(ctl.commandCollector.CurrentPromptAndCommand())
		default:
		}
	}
}
