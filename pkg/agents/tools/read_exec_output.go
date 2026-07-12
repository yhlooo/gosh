package tools

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/gosh/pkg/term"
)

const (
	// ReadExecOutputName 读取命令执行输出工具名
	ReadExecOutputName = "ReadExecOutput"

	// ReadExecOutputDesc 读取命令执行输出工具描述
	ReadExecOutputDesc = `读取指定命令执行时 stdout/stderr 输出内容。

通过 index 指定需要查询输出的命令。

每次读取内容不建议过多，通过 limit 和 offset 组合可以按需获取部分输出内容，建议每次读取不超过 8KiB (8192) 。
`
)

// ReadExecOutputInput 读命令执行输出工具入参
type ReadExecOutputInput struct {
	// 命令序号
	Index int `json:"index" jsonschema_description:"查询的命令序号"`
	// 读取偏移字节
	Offset int64 `json:"offset,omitempty" jsonschema_description:"从指定字节偏移开始读取"`
	// 限制最大读取字节数
	// 默认： 8Ki
	Limit int64 `json:"limit,omitempty" jsonschema:"description=限制最大读取字节数,default=8192,max=102400"`
}

// ReadExecOutputOutput 读命令执行输出工具出参
type ReadExecOutputOutput struct {
	Output string `json:"output" jsonschema_description:"命令执行输出内容"`
}

// ReadExecOutputFn 读命令执行输出工具方法
func ReadExecOutputFn(cc *term.CommandCollector) ai.ToolFunc[ReadExecOutputInput, ReadExecOutputOutput] {
	return func(ctx *ai.ToolContext, in ReadExecOutputInput) (ReadExecOutputOutput, error) {
		if in.Limit <= 0 {
			in.Limit = 8 << 10
		}
		if in.Limit > 100<<10 {
			in.Limit = 100 << 10
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
	return genkit.DefineTool(g, ReadExecOutputName, ReadExecOutputDesc, ReadExecOutputFn(cc))
}
