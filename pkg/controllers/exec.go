package controllers

import (
	"context"
	"fmt"

	"github.com/yhlooo/gosh/pkg/term"
)

// ExecuteCommand 在 shell 中执行命令
func (ctl *Controller) ExecuteCommand(ctx context.Context, cmdline string) (*term.CommandRecord, error) {

	if state := ctl.commandCollector.State(); state != term.OutputCommand {
		return nil, fmt.Errorf("the shell is busy, cannot execute new commands")
	}

	ch := make(chan term.CommandRecord, 1)
	ctl.commandCollector.WaitNextCommand(ch)

	// 恢复 prompt 显示
	_, _ = ctl.output.Write(append([]byte{'\r', '\n'}, ctl.commandCollector.CurrentPromptAndCommand()...))

	// 发送命令
	// Ctrl+U Ctrl+K cmdline Enter
	ctl.inputLock.Lock()
	if _, err := ctl.ptmx.WriteString("\x15\v" + cmdline + "\r"); err != nil {
		ctl.inputLock.Unlock()
		return nil, fmt.Errorf("send cmdline to shell error: %w", err)
	}
	ctl.inputLock.Unlock()

	// 等待命令结束
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("wait for command finish error: channel closed unexpectedly")
		}

		return &r, nil
	}
}
