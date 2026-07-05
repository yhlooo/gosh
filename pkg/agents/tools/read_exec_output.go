package tools

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/gosh/pkg/term"
)

// ReadExecOutputInput 读命令执行输出工具入参
type ReadExecOutputInput struct {
	// 命令序号
	Index int `json:"index"`
	// 读取偏移字节
	Offset int64 `json:"offset,omitempty"`
	// 限制最大读取字节数
	// 默认： 8Ki
	Limit int64 `json:"limit,omitempty"`
}

// ReadExecOutputOutput 读命令执行输出工具出参
type ReadExecOutputOutput struct {
	Output string `json:"output"`
}

// ReadExecOutputFn 读命令执行输出工具方法
func ReadExecOutputFn(cc *term.CommandCollector) ai.ToolFunc[ReadExecOutputInput, ReadExecOutputOutput] {
	return func(ctx *ai.ToolContext, in ReadExecOutputInput) (ReadExecOutputOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 8 << 10
		}

		content, err := cc.ReadExecOutput(in.Index, in.Offset, in.Limit)
		if err != nil {
			return ReadExecOutputOutput{}, err
		}

		return ReadExecOutputOutput{Output: string(content)}, nil
	}
}

// ToolReadExecOutput 读命令执行输出工具
type ToolReadExecOutput = *ai.ToolDef[ReadExecOutputInput, ReadExecOutputOutput]

// DefineToolReadExecOutput 注册 ReadExecOutput 工具
func DefineToolReadExecOutput(g *genkit.Genkit, cc *term.CommandCollector) ToolReadExecOutput {
	return genkit.DefineTool(
		g, "ReadExecOutput",
		"读取指定命令执行输出内容", // TODO: ...
		ReadExecOutputFn(cc),
	)
}
