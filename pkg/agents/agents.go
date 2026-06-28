package agents

import (
	"context"

	"github.com/firebase/genkit/go/ai"
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

// Options Agent 运行选项
type Options struct {
	ChatOutputStreamHandler ai.ModelStreamCallback
}
