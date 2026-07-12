package tools

import (
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/gosh/pkg/term"
)

const (
	// GetHistoryName 获取命令执行历史工具名
	GetHistoryName = "GetHistory"

	// GetHistoryDesc 获取命令执行历史工具描述
	GetHistoryDesc = `获取最近执行的命令历史。

通过 lastN 指定获取最近多少个命令，如无必要建议初始获取不超过 10 个历史命令。

在返回结果超过 10 条记录时，结果将省略命令执行输出 (output) ，除此之外未执行完成命令的输出和超过 8KiB 的命令输出也会被省略，通过 outputOmitted 可以确定输出是否被省略。
命令执行输出被省略时，如有必要可以通过 ReadExecOutput 使用命令 index 查询获取输出内容。

在命令执行未完成时，其 duration / exitCode / outputLength 为空。

因 shell 输出流解析问题， cmdline 中可能混杂部分输入提示符
`
)

// GetHistoryInput 获取命令执行历史工具入参
type GetHistoryInput struct {
	// 获取最近多少个命令
	// 默认： 10
	LastN int `json:"lastN,omitempty" jsonschema:"description=获取最近多少个命令,default=10,max=100"`
}

// GetHistoryOutput 获取命令执行历史工具出参
type GetHistoryOutput struct {
	Items []CommandRecord `json:"items" jsonschema_description:"历史执行的命令列表，按执行时间升序排列"`
}

// CommandRecord 命令记录
type CommandRecord struct {
	// 命令序号
	Index int `json:"index" jsonschema_description:"命令序号"`
	// 命令行
	CommandLine string `json:"cmdline" jsonschema_description:"命令行"`
	// 开始时间
	StartTime time.Time `json:"startTime" jsonschema_description:"命令执行开始时间"`
	// 执行耗时
	Duration *time.Duration `json:"duration,omitempty" jsonschema_description:"命令执行耗时"`
	// 命令执行退出码
	ExitCode *int `json:"exitCode,omitempty" jsonschema_description:"退出码"`
	// 命令执行输出内容
	Output string `json:"output,omitempty" jsonschema_description:"命令执行期间 stdout/stderr 输出内容，在 outputOmitted=true 时该值为空"`
	// 是否省略了命令执行输出
	OutputOmitted bool `json:"outputOmitted,omitempty" jsonschema_description:"当前命令记录是否省略了命令执行输出"`
	// 输出长度
	OutputLength *int64 `json:"outputLength,omitempty" jsonschema_description:"命令执行输出内容字节长度"`
}

// GetHistoryFn 返回获取命令执行历史工具方法
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
			var duration *time.Duration
			if record.EndTime != nil {
				duration = new(record.EndTime.Sub(record.StartTime))
			}

			ret = append(ret, CommandRecord{
				Index:         record.Index,
				CommandLine:   record.CommandLine,
				StartTime:     record.StartTime,
				Duration:      duration,
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
	return genkit.DefineTool(g, GetHistoryName, GetHistoryDesc, GetHistoryFn(cc))
}
