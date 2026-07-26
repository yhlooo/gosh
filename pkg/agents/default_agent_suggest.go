package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/gosh/pkg/agents/tools"
	"github.com/yhlooo/gosh/pkg/genkitplugins/oai"
	"github.com/yhlooo/gosh/pkg/models"
	"github.com/yhlooo/gosh/pkg/tokentracker"
)

// GenCmdlineSuggestion 生成建议的命令行后续内容
func (a *GoshAgent) GenCmdlineSuggestion(ctx context.Context) (string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	curIndex := a.commandCollector.LastCommandIndex()
	prefix := string(a.commandCollector.CurrentCommand())

	// 首先从缓存获取
	suggestion, ok := a.cmdlineSuggestionCache.Get(curIndex, prefix)
	if ok {
		return suggestion, nil
	}

	// 获取命令执行历史
	cmdRecords := a.commandCollector.ListLastNCommands(10)
	history := make([]tools.CommandRecord, 0, len(cmdRecords))
	for i, record := range cmdRecords {
		var outputLen *int64
		if record.ExecOutputEnd != nil {
			outputLen = new(*record.ExecOutputEnd - record.ExecOutputStart)
		}
		outputOmitted := len(cmdRecords)-i > 3 || outputLen == nil || *outputLen > 8<<10
		output := ""
		if !outputOmitted {
			outputRaw, err := a.commandCollector.ReadExecOutput(record.Index, 0, 8<<10)
			if err != nil {
				logger.Error(err, fmt.Sprintf("read command %d output error", record.Index))
			}
			output = string(outputRaw)
		}
		var duration *time.Duration
		if record.EndTime != nil {
			duration = new(record.EndTime.Sub(record.StartTime))
		}

		history = append(history, tools.CommandRecord{
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

	useModels := a.currentModels
	if useModels.Primary == "" {
		return "", fmt.Errorf("no available model")
	}
	useModels.ReasoningLevel = new(0)

	ctx = models.ContextWithModels(ctx, useModels)
	ctx = tokentracker.ContextWithTokenTracker(ctx, a.tokenTracker)

	// 生成建议
	suggestOut, err := a.genCmdlineSuggestionFlow.Run(ctx, GenCmdlineSuggestionInput{
		History: history,
		Prefix:  prefix,
	})
	if err != nil {
		return "", err
	}

	// 记录缓存
	a.cmdlineSuggestionCache.Add(curIndex, prefix, suggestOut.Suggestion)

	return suggestOut.Suggestion, nil
}

// GenCmdlineSuggestionInput 生成命令行建议输入
type GenCmdlineSuggestionInput struct {
	History []tools.CommandRecord `json:"history,omitempty"`
	Prefix  string                `json:"prefix,omitempty"`
}

// GenCmdlineSuggestionOutput 生成命令行建议输出
type GenCmdlineSuggestionOutput struct {
	Suggestion string `json:"suggestion,omitempty"`
}

// GenCmdlineSuggestionFlow 生成命令行建议流程
type GenCmdlineSuggestionFlow = *core.Flow[GenCmdlineSuggestionInput, GenCmdlineSuggestionOutput, struct{}]

// handleGenCmdlineSuggestion 处理命令行建议生成
func (a *GoshAgent) handleGenCmdlineSuggestion(
	ctx context.Context,
	in GenCmdlineSuggestionInput,
) (GenCmdlineSuggestionOutput, error) {
	inRaw, _ := json.MarshalIndent(in, "", "  ")

	resp, err := genkit.Generate(
		ctx, a.g,
		ai.WithModelName(a.currentModels.GetLite()),
		ai.WithConfig(oai.GenerateConfig{ReasoningLevel: 0}),
		ai.WithUse(a.tokenTracker.Middleware()),
		ai.WithSystem(`根据输入的历史命令执行记录和当前输入的命令前缀推测当前最有可能的后续命令内容作为推荐。

比如输入命令前缀 "ec" ，后续内容可能是 "ho ..." ，比较确定的部分是 "echo" 命令，因此推荐 "ho " 。建议应尽量包含直到行尾的内容，而不是仅建议补全一个命令或参数，除非后续参数随机性过大没有推荐意义。

推荐也应该参考历史执行记录。比如之前用户执行了 "kubectl get pod" ，其输出包含一个名为 "example-pod" 的 Pod 名。
当前输入前缀 "kubectl get pod ex" ，那么用户很可能是希望查看这个 Pod 的详细信息，因此可以推荐 "ample-pod -o yaml" 。

## 输入格式

输入为一个 JSON 格式字符串，其各字段含义如下：

- **history** ([]History) 命令执行历史

  其中每一项包含以下字段：

  - **index** (int) 命令序号
  - **cmdline** (string) 命令行
  - **startTime** (Time) 命令执行开始时间
  - **duration** (Duration) 命令执行耗时
  - **exitCode** (int) 退出码
  - **output** (string) 命令执行期间 stdout/stderr 输出内容，在 outputOmitted=true 时该值为空
  - **outputOmitted** (bool) 当前命令记录是否省略了命令执行输出
  - **outputLength** (int) 命令执行输出内容字节长度

- **prefix** (string) 当前输入的命令前缀，推荐应基于该前缀推荐后续内容

## 输出格式

输出 **必须** 包含 <suggestion> </suggestion> 标签，中间是建议的命令行剩余内容

比如输入前缀 "ec" ，输出 "<suggestion>ho </suggestion>" ，表示建议在 "ec" 后跟随内容 "ho "

如果没有建议的内容可以输出 "<suggestion></suggestion>"

`),
		ai.WithPrompt(string(inRaw)),
	)
	if err != nil {
		return GenCmdlineSuggestionOutput{}, err
	}

	respContent := resp.Text()
	divided := strings.SplitN(respContent, "<suggestion>", 2)
	if len(divided) != 2 {
		return GenCmdlineSuggestionOutput{}, nil
	}
	suggestion := divided[1]

	divided = strings.SplitN(suggestion, "</suggestion>", 2)
	if len(divided) != 2 {
		return GenCmdlineSuggestionOutput{
			Suggestion: suggestion,
		}, nil
	}

	return GenCmdlineSuggestionOutput{
		Suggestion: divided[0],
	}, nil
}

// CommandLineSuggestionCache 命令行推荐缓存
type CommandLineSuggestionCache struct {
	lock             sync.RWMutex
	lastCommandIndex int
	cachePrefixes    map[string]string
}

// Get 获取缓存
func (sc *CommandLineSuggestionCache) Get(lastCommandIndex int, prefix string) (string, bool) {
	sc.lock.RLock()
	defer sc.lock.RUnlock()

	if lastCommandIndex != sc.lastCommandIndex {
		return "", false
	}

	ret, ok := sc.cachePrefixes[prefix]
	return ret, ok
}

// Add 添加内容到缓存
func (sc *CommandLineSuggestionCache) Add(lastCommandIndex int, prefix, suggestion string) {
	sc.lock.Lock()
	defer sc.lock.Unlock()

	if lastCommandIndex != sc.lastCommandIndex {
		// 执行了新的命令，清空缓存
		sc.lastCommandIndex = lastCommandIndex
		sc.cachePrefixes = make(map[string]string)
	}

	if sc.cachePrefixes == nil {
		sc.cachePrefixes = make(map[string]string)
	}

	sc.cachePrefixes[prefix] = suggestion
}
