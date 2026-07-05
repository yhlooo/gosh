package tools

import (
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/gosh/pkg/agents/generic"
)

// ExecInput 命令执行工具输入
type ExecInput struct {
	CommandLine string `json:"cmdline"`
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
	return genkit.DefineTool(
		g, "Exec",
		"在当前 shell 中执行命令", // TODO: ...
		ExecFn(ctl),
	)
}
