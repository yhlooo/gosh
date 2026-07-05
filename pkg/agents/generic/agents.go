package generic

import (
	"context"

	"github.com/firebase/genkit/go/ai"

	"github.com/yhlooo/gosh/pkg/term"
)

// Agent 接口
type Agent interface {
	// Initialize 初始化
	Initialize(ctx context.Context, opts Options) error
	// Chat 发送指令开始一轮对话并等待指令处理完成
	Chat(ctx context.Context, prompt string) error
	// Cancel 取消当前正在处理的指令
	Cancel() error
}

// ShellController shell 控制器
type ShellController interface {
	// CommandCollector 返回命令收集器
	CommandCollector() *term.CommandCollector
	// ExecuteCommand 执行命令
	ExecuteCommand(ctx context.Context, cmdline string) (*term.CommandRecord, error)
}

// Options Agent 运行选项
type Options struct {
	// 终端类型， TERM 变量的值
	TerminalType string
	// 与 Agent 对话输出流处理器
	ChatOutputStreamHandler ai.ModelStreamCallback
	// Shell 控制器
	ShellController ShellController
}
