package tools

import (
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/gosh/pkg/term"
)

// GetHistoryInput 获取命令执行历史工具入参
type GetHistoryInput struct {
	// 获取最近多少个命令
	// 默认： 10
	LastN int `json:"lastN,omitempty"`
}

// GetHistoryOutput 获取命令执行历史工具出参
type GetHistoryOutput struct {
	Items []CommandRecord `json:"items"`
}

// CommandRecord 命令记录
type CommandRecord struct {
	// 命令序号
	Index int `json:"index"`
	// 命令行
	CommandLine string `json:"commandLine"`
	// 开始时间
	StartTime time.Time `json:"startTime"`
	// 结束时间
	EndTime *time.Time `json:"endTime,omitempty"`
	// 命令执行退出码
	ExitCode *int `json:"exitCode,omitempty"`
	// 命令执行输出内容
	Output string `json:"output,omitempty"`
	// 是否省略了命令执行输出
	OutputOmitted bool `json:"outputOmitted,omitempty"`
	// 输出长度
	OutputLength *int64 `json:"outputLength,omitempty"`
}

// GetHistoryFn 返回获取命令执行历史方法
func GetHistoryFn(cc *term.CommandCollector) ai.ToolFunc[GetHistoryInput, GetHistoryOutput] {
	return func(ctx *ai.ToolContext, in GetHistoryInput) (GetHistoryOutput, error) {
		if in.LastN <= 0 {
			in.LastN = 10
		}

		records := cc.ListLastNCommands(in.LastN)

		if len(records) == 0 {
			return GetHistoryOutput{}, nil
		}

		// 多于 10 个命令时不在结果中展示命令输出
		showOutput := true
		if len(records) > 10 {
			showOutput = false
		}

		ret := make([]CommandRecord, 0, len(records))
		for _, record := range records {
			var outputLen *int64
			if record.ExecOutputEnd != nil {
				outputLen = new(*record.ExecOutputEnd - record.ExecOutputStart)
			}
			outputOmitted := !showOutput || outputLen == nil || *outputLen > 8<<10
			output := ""
			if !outputOmitted {
				outputRaw, err := cc.ReadExecOutput(record.Index, 0, 8<<10)
				if err != nil {
					return GetHistoryOutput{Items: ret},
						fmt.Errorf("read command %d output error: %w", record.Index, err)
				}
				output = string(outputRaw)
			}

			ret = append(ret, CommandRecord{
				Index:         record.Index,
				CommandLine:   record.CommandLine,
				StartTime:     record.StartTime,
				EndTime:       record.EndTime,
				ExitCode:      record.ExitCode,
				Output:        output,
				OutputOmitted: outputOmitted,
				OutputLength:  outputLen,
			})
		}

		return GetHistoryOutput{Items: ret}, nil
	}
}

// ToolGetHistory 获取命令执行历史工具
type ToolGetHistory = *ai.ToolDef[GetHistoryInput, GetHistoryOutput]

// DefineToolGetHistory 注册 GetHistory 工具
func DefineToolGetHistory(g *genkit.Genkit, cc *term.CommandCollector) ToolGetHistory {
	return genkit.DefineTool(
		g, "GetHistory",
		"获取历史执行命令", // TODO: ...
		GetHistoryFn(cc),
	)
}
