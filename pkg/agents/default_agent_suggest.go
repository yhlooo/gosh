package agents

import (
	"context"
	"time"
)

// GenerateCommandLineSuggestion 生成建议的命令行后续内容
func (a *GoshAgent) GenerateCommandLineSuggestion(_ context.Context) (string, error) {
	// TODO: 建议内容生成有待实现，写个固定内容
	time.Sleep(time.Second)
	return "TODO not implemented 你好", nil
}
