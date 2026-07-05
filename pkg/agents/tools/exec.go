package tools

import (
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/gosh/pkg/agents/generic"
)

// ExecDesc 命令执行工具描述
const ExecDesc = `在当前 shell 中执行命令。

通过 cmdline 指定需要执行的命令行，注意需要符合当前 shell 语法。

命令执行结束后返回执行命令的记录。在命令执行未完成和命令输出长度超过 8KiB 时命令输出 (output) 会被省略，通过 outputOmitted 可以确定输出是否被省略。
命令执行输出被省略时，如有必要可以通过 ReadExecOutput 使用命令 index 查询获取输出内容。
无论工具返回的结果中是否省略了输出内容，用户都能看到输出内容，如果只是为了让用户知晓输出内容并不需要使用 ReadExecOutput ，只需要让用户“查看上面的输出内容”。

在命令执行未完成时，其 duration / exitCode / outputLength 为空。

执行的命令和命令执行时的输出内容会同步展示给用户。没有必要单纯复述这些内容给用户，如需提醒用户查看可以说“看上面的输出内容”。
`

// ExecInput 命令执行工具输入
type ExecInput struct {
	CommandLine string `json:"cmdline" jsonschema_description:"执行的命令行"`
}

// ExecOutput 命令执行工具输出
type ExecOutput struct {
	CommandRecord
}

// ExecFn 命令执行工具方法
func ExecFn(ctl generic.ShellController) ai.ToolFunc[ExecInput, ExecOutput] {
	cc := ctl.CommandCollector()
	return func(ctx *ai.ToolContext, in ExecInput) (ExecOutput, error) {
		logger := logr.FromContextOrDiscard(ctx)

		if in.CommandLine == "" {
			return ExecOutput{}, fmt.Errorf("empty cmdline")
		}
		rec, err := ctl.ExecuteCommand(ctx, in.CommandLine)
		if err != nil {
			return ExecOutput{}, err
		}

		var outputLen *int64
		if rec.ExecOutputEnd != nil {
			outputLen = new(*rec.ExecOutputEnd - rec.ExecOutputStart)
		}
		outputOmitted := outputLen == nil || *outputLen > 8<<10
		output := ""
		if !outputOmitted {
			outputRaw, err := cc.ReadExecOutput(rec.Index, 0, 8<<10)
			if err != nil {
				logger.Error(err, "read exec output error")
			}
			output = string(outputRaw)
		}
		var duration *time.Duration
		if rec.EndTime != nil {
			duration = new(rec.EndTime.Sub(rec.StartTime))
		}

		return ExecOutput{
			CommandRecord: CommandRecord{
				Index:         rec.Index,
				CommandLine:   rec.CommandLine,
				StartTime:     rec.StartTime,
				Duration:      duration,
				ExitCode:      rec.ExitCode,
				Output:        output,
				OutputOmitted: outputOmitted,
				OutputLength:  outputLen,
			},
		}, nil
	}
}

// ToolExec 命令执行工具
type ToolExec = *ai.ToolDef[ExecInput, ExecOutput]

// DefineToolExec 注册 Exec 工具
func DefineToolExec(g *genkit.Genkit, ctl generic.ShellController) ToolExec {
	return genkit.DefineTool(g, "Exec", ExecDesc, ExecFn(ctl))
}
