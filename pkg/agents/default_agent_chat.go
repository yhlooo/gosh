package agents

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/gosh/pkg/genkitplugins/oai"
	"github.com/yhlooo/gosh/pkg/models"
	"github.com/yhlooo/gosh/pkg/tokentracker"
)

// Chat 发送指令开始一轮对话并等待指令处理完成
func (a *GoshAgent) Chat(ctx context.Context, prompt string) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)
	ctx = logr.NewContext(ctx, logger)

	a.session.lock.Lock()

	// 取消当前正在进行的对话
	if a.session.cancelPrompt != nil {
		logger.Info("cancel processing prompt")
		a.session.cancelPrompt()
	}
	ctx, cancel := context.WithCancel(ctx)
	a.session.cancelPrompt = cancel

	// 获取历史记录
	messages := a.session.history
	lastContextWindow := a.session.lastContextWindow

	defer func() {
		a.session.lock.Lock()
		a.session.cancelPrompt = nil
		a.session.history = messages
		a.session.lastContextWindow = lastContextWindow
		a.session.lock.Unlock()
		cancel()
	}()

	a.session.lock.Unlock()

	if lastContextWindow > a.opts.MaxContextWindow {
		return fmt.Errorf(
			"%w: context window size exceeds limit: %d (>%d)",
			ErrContextWindowLimit, lastContextWindow, a.opts.MaxContextWindow,
		)
	}
	if a.currentModels.Primary == "" {
		return fmt.Errorf("no available model")
	}
	if prompt == "" {
		return nil
	}

	ctx = models.ContextWithModels(ctx, a.currentModels)
	ctx = tokentracker.ContextWithTokenTracker(ctx, a.tokenTracker)

	logger.Info("chat turn start")
	defer logger.Info("chat turn end")

	history := make([]*ai.Message, len(messages))
	copy(history, messages)
	messages = append(messages, ai.NewUserTextMessage(prompt))

	chatOut, err := a.chatTurnFlow.Run(ctx, ChatTurnInput{
		Prompt:  prompt,
		History: history,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = fmt.Errorf("%w: user cancelled", ErrUserCancelled)
			messages = append(messages, ai.NewModelTextMessage("[System] user cancelled"))
		} else {
			messages = append(messages, ai.NewModelTextMessage("[Error] "+err.Error()))
		}
		return err
	}

	messages = append(messages, chatOut.Messages...)
	lastContextWindow = chatOut.LastContextWindow
	if lastContextWindow > a.opts.MaxContextWindow {
		return fmt.Errorf(
			"%w: context window size exceeds limit: %d (>%d)",
			ErrContextWindowLimit, lastContextWindow, a.opts.MaxContextWindow,
		)
	}

	return nil
}

// ChatTurnInput 对话输入
type ChatTurnInput struct {
	Prompt  string        `json:"prompt"`
	History []*ai.Message `json:"history,omitempty"`
}

// ChatTurnOutput 对话输出
type ChatTurnOutput struct {
	Messages          []*ai.Message `json:"messages"`
	LastContextWindow int64         `json:"lastContextWindow,omitempty"`
}

// ChatTurnFlow 对话流程
type ChatTurnFlow = *core.Flow[ChatTurnInput, ChatTurnOutput, struct{}]

// ToolCallError 工具调用错误
type ToolCallError struct {
	Err string `json:"error"`
}

var _ error = ToolCallError{}

// Error 返回错误描述
func (e ToolCallError) Error() string {
	return e.Err
}

// handleChatTurn 处理一轮对话
func (a *GoshAgent) handleChatTurn(ctx context.Context, in ChatTurnInput) (ChatTurnOutput, error) {
	opts := []ai.GenerateOption{
		ai.WithSystem(`你是一个 shell 专家，负责解决用户关于 shell 的问题

## 严格遵循以下要求进行回答
- 以用户提问的语言回答问题，比如用户用中文提问就用中文回答，用户用英文提问就用英文回答；
`), // TODO: 待完善
		ai.WithReturnToolRequests(true),
		ai.WithUse(tokentracker.TrackerFromContext(ctx).Middleware()),
		ai.WithStreaming(handleTextStream(a.handleChatOutputStream, true, true)),
		ai.WithTools(a.availableTools...),
	}
	modelName := ""
	reasoningLevel := 0
	if m, ok := models.FromContext(ctx); ok {
		modelName = m.GetPrimary()
		reasoningLevel = m.GetReasoningLevel()
	}
	if modelName != "" {
		opts = append(opts,
			ai.WithModelName(modelName),
			ai.WithConfig(oai.GenerateConfig{ReasoningLevel: reasoningLevel}),
		)
	}

	output := ChatTurnOutput{}
	messages := slices.Clone(in.History)
	promptMsg := ai.NewUserTextMessage(in.Prompt)
	messages = append(messages, promptMsg)
	for {
		subTurnOpts := append([]ai.GenerateOption{ai.WithMessages(messages...)}, opts...)

		// 检查上下文限制
		if output.LastContextWindow > a.opts.MaxContextWindow {
			return output, nil
		}

		// 进行一轮生成
		resp, err := genkit.Generate(ctx, a.g, subTurnOpts...)
		if err != nil {
			return output, err
		}
		messages = append(messages, resp.Message)
		if resp.Usage != nil {
			output.LastContextWindow = int64(resp.Usage.InputTokens)
		}

		toolRequests := resp.ToolRequests()
		if len(toolRequests) == 0 {
			// 结束对话
			output.Messages = append(output.Messages, resp.Message)
			return output, nil
		}

		output.Messages = append(output.Messages, resp.Message)

		// 调用工具
		var parts []*ai.Part
		for _, toolReq := range toolRequests {
			if err := a.handleChatOutputStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{ai.NewToolRequestPart(toolReq)},
				Role:    resp.Message.Role,
			}); err != nil {
				return output, fmt.Errorf("handle stream error: %w", err)
			}

			toolResp := handleToolCall(ctx, a.g, toolReq)
			parts = append(parts, toolResp)

			if err := a.handleChatOutputStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{toolResp},
				Role:    ai.RoleTool,
			}); err != nil {
				return output, fmt.Errorf("handle stream error: %w", err)
			}
		}
		toolRespMessage := ai.NewMessage(ai.RoleTool, nil, parts...)
		messages = append(messages, toolRespMessage)
		output.Messages = append(output.Messages, toolRespMessage)
	}
}

// handleChatOutputStream 处理对话轮流式输出
func (a *GoshAgent) handleChatOutputStream(ctx context.Context, chunk *ai.ModelResponseChunk) error {
	if a.genericOpts.ChatOutputStreamHandler != nil {
		return a.genericOpts.ChatOutputStreamHandler(ctx, chunk)
	}
	return nil
}

// handleTextStream 处理文本流
func handleTextStream(handler ai.ModelStreamCallback, reasoning, text bool) ai.ModelStreamCallback {
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		content := make([]*ai.Part, 0, len(chunk.Content))
		for _, part := range chunk.Content {
			if reasoning && part.IsReasoning() {
				content = append(content, part)
			} else if text && (part.IsText() || part.IsData()) {
				content = append(content, part)
			}
		}
		chunk.Content = content
		if len(chunk.Content) == 0 {
			return nil
		}
		return handler(ctx, chunk)
	}
}

// handleToolCall 处理工具调用
func handleToolCall(ctx context.Context, g *genkit.Genkit, req *ai.ToolRequest) *ai.Part {
	tool := genkit.LookupTool(g, req.Name)
	if tool == nil {
		// 找不到工具
		return ai.NewToolResponsePart(&ai.ToolResponse{
			Name:   req.Name,
			Ref:    req.Ref,
			Output: ToolCallError{Err: fmt.Sprintf("tool %q not found", req.Name)},
		})
	}

	output, err := tool.RunRaw(ctx, req.Input)
	if err != nil {
		return ai.NewToolResponsePart(&ai.ToolResponse{
			Name:   req.Name,
			Ref:    req.Ref,
			Output: ToolCallError{Err: fmt.Sprintf("call tool %q error: %s", req.Name, err.Error())},
		})
	}

	return ai.NewToolResponsePart(&ai.ToolResponse{
		Name:   req.Name,
		Ref:    req.Ref,
		Output: output,
	})
}
