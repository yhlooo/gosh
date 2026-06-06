package oai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

// NewModelGenerator 创建模型生成器
func NewModelGenerator(client *openai.Client, req *ai.ModelRequest, opts ModelOptions) *ModelGenerator {
	rawReq := &openai.ChatCompletionNewParams{
		Model: opts.Label,
	}

	cfg, hasCfg := req.Config.(GenerateConfig)

	reasoningLevel := 2
	if hasCfg {
		reasoningLevel = cfg.ReasoningLevel
	}
	if !opts.Reasoning {
		reasoningLevel = 0
	}
	switch reasoningLevel {
	case 1:
		rawReq.SetExtraFields(opts.ReasoningEffortFields[1])
	case 2:
		rawReq.SetExtraFields(opts.ReasoningEffortFields[2])
	default: // 0
		rawReq.SetExtraFields(opts.ReasoningEffortFields[0])
	}

	return &ModelGenerator{
		client:                client,
		modelName:             opts.Label,
		request:               rawReq,
		reasoningContentField: opts.ReasoningContentField,
	}
}

// ModelGenerator 模型生成器
type ModelGenerator struct {
	client                *openai.Client
	modelName             string
	reasoningContentField string

	request *openai.ChatCompletionNewParams

	messages   []openai.ChatCompletionMessageParamUnion
	tools      []openai.ChatCompletionToolParam
	toolChoice openai.ChatCompletionToolChoiceOptionUnionParam
	// 存储初始化期间的错误
	err error
}

// WithMessages 添加消息到生成器
func (g *ModelGenerator) WithMessages(messages []*ai.Message) *ModelGenerator {
	if g.err != nil {
		return g
	}

	if messages == nil {
		return g
	}

	oaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		content := concatenateContent(msg.Content)
		switch msg.Role {
		case ai.RoleSystem:
			oaiMessages = append(oaiMessages, openai.SystemMessage(content))
		case ai.RoleModel:
			am := openai.ChatCompletionAssistantMessageParam{}
			am.Content.OfString = param.NewOpt(content)
			toolCalls, err := convertToolCalls(msg.Content)
			if err != nil {
				g.err = err
				return g
			}
			if len(toolCalls) > 0 {
				am.ToolCalls = toolCalls
			}
			am.SetExtraFields(map[string]any{g.reasoningContentField: concatenateReasoningContent(msg.Content)})
			oaiMessages = append(oaiMessages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &am,
			})
		case ai.RoleTool:
			for _, p := range msg.Content {
				if !p.IsToolResponse() {
					continue
				}
				// Use the captured tool call ID (Ref) if available, otherwise fall back to tool name
				toolCallID := p.ToolResponse.Ref
				if toolCallID == "" {
					toolCallID = p.ToolResponse.Name
				}

				toolOutput, err := anyToJSONString(p.ToolResponse.Output)
				if err != nil {
					g.err = err
					return g
				}
				tm := openai.ToolMessage(toolOutput, toolCallID)
				oaiMessages = append(oaiMessages, tm)
			}
		case ai.RoleUser:
			var parts []openai.ChatCompletionContentPartUnionParam
			for _, p := range msg.Content {
				if p.IsText() {
					parts = append(parts, openai.TextContentPart(p.Text))
				}
				if p.IsMedia() {
					part := openai.ImageContentPart(
						openai.ChatCompletionContentPartImageImageURLParam{
							URL: p.Text,
						})
					parts = append(parts, part)
					continue
				}
			}
			if len(parts) > 0 {
				oaiMessages = append(oaiMessages, openai.ChatCompletionMessageParamUnion{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{OfArrayOfContentParts: parts},
					},
				})
			}
		default:
			// 不支持的角色
			continue
		}
	}
	g.messages = oaiMessages
	return g
}

// WithTools 添加工具到生成器
func (g *ModelGenerator) WithTools(tools []*ai.ToolDefinition) *ModelGenerator {
	if g.err != nil {
		return g
	}

	if tools == nil {
		return g
	}

	toolParams := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, tool := range tools {
		if tool == nil || tool.Name == "" {
			continue
		}

		toolParams = append(toolParams, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  tool.InputSchema,
				Strict:      openai.Bool(false), // TODO: implement strict mode
			},
		})
	}

	if len(toolParams) > 0 {
		g.tools = toolParams
	}
	return g
}

// Generate 生成
func (g *ModelGenerator) Generate(
	ctx context.Context,
	req *ai.ModelRequest,
	handleChunk core.StreamCallback[*ai.ModelResponseChunk],
) (*ai.ModelResponse, error) {
	if g.err != nil {
		return nil, g.err
	}

	if len(g.messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}
	g.request.Messages = g.messages

	if len(g.tools) > 0 {
		g.request.Tools = g.tools
	}

	if handleChunk != nil {
		return g.generateStream(ctx, handleChunk)
	}
	return g.generateComplete(ctx, req)
}

// generateStream 对话补全流式生成
func (g *ModelGenerator) generateStream(ctx context.Context, handleChunk core.StreamCallback[*ai.ModelResponseChunk]) (*ai.ModelResponse, error) {
	stream := g.client.Chat.Completions.NewStreaming(
		ctx, *g.request,
		option.WithJSONSet("stream_options.include_usage", true),
	)
	defer func() { _ = stream.Close() }()

	var fullResponse ai.ModelResponse
	fullResponse.Message = &ai.Message{
		Role: ai.RoleModel,
	}

	// Initialize request and usage
	fullResponse.Request = &ai.ModelRequest{}
	fullResponse.Usage = &ai.GenerationUsage{
		InputTokens:  0,
		OutputTokens: 0,
		TotalTokens:  0,
	}

	var currentToolCall *ai.ToolRequest
	var currentArguments string
	var toolCallCollects []struct {
		toolCall *ai.ToolRequest
		args     string
	}

	for stream.Next() {
		chunk := stream.Current()

		fullResponse.Usage.InputTokens += int(chunk.Usage.PromptTokens)
		fullResponse.Usage.OutputTokens += int(chunk.Usage.CompletionTokens)
		fullResponse.Usage.ThoughtsTokens += int(chunk.Usage.CompletionTokensDetails.ReasoningTokens)
		fullResponse.Usage.CachedContentTokens += int(chunk.Usage.PromptTokensDetails.CachedTokens)
		fullResponse.Usage.TotalTokens += int(chunk.Usage.TotalTokens)

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		modelChunk := &ai.ModelResponseChunk{}

		switch choice.FinishReason {
		case "tool_calls", "stop":
			fullResponse.FinishReason = ai.FinishReasonStop
		case "length":
			fullResponse.FinishReason = ai.FinishReasonLength
		case "content_filter":
			fullResponse.FinishReason = ai.FinishReasonBlocked
		case "function_call":
			fullResponse.FinishReason = ai.FinishReasonOther
		default:
			fullResponse.FinishReason = ai.FinishReasonUnknown
		}

		// handle tool calls
		for _, toolCall := range choice.Delta.ToolCalls {
			// first tool call (= current tool call is nil) contains the tool call name
			if currentToolCall != nil && toolCall.ID != "" && currentToolCall.Ref != toolCall.ID {
				toolCallCollects = append(toolCallCollects, struct {
					toolCall *ai.ToolRequest
					args     string
				}{
					toolCall: currentToolCall,
					args:     currentArguments,
				})
				currentToolCall = nil
				currentArguments = ""
			}

			if currentToolCall == nil {
				currentToolCall = &ai.ToolRequest{
					Name: toolCall.Function.Name,
					Ref:  toolCall.ID,
				}
			}

			if toolCall.Function.Arguments != "" {
				currentArguments += toolCall.Function.Arguments
			}

			modelChunk.Content = append(modelChunk.Content, ai.NewToolRequestPart(&ai.ToolRequest{
				Name:  currentToolCall.Name,
				Input: toolCall.Function.Arguments,
				Ref:   currentToolCall.Ref,
			}))
		}

		// when tool call is complete
		if choice.FinishReason == "tool_calls" && currentToolCall != nil {
			// parse accumulated arguments string
			for _, toolcall := range toolCallCollects {
				args, err := jsonStringToMap(toolcall.args)
				if err != nil {
					return nil, fmt.Errorf("generate error: could not parse tool args: %w", err)
				}
				toolcall.toolCall.Input = args
				fullResponse.Message.Content = append(fullResponse.Message.Content, ai.NewToolRequestPart(toolcall.toolCall))
			}
			if currentArguments != "" {
				args, err := jsonStringToMap(currentArguments)
				if err != nil {
					return nil, fmt.Errorf("generate error: could not parse tool args: %w", err)
				}
				currentToolCall.Input = args
			}
			fullResponse.Message.Content = append(fullResponse.Message.Content, ai.NewToolRequestPart(currentToolCall))
			currentToolCall = nil
			currentArguments = ""
		}

		// 思考
		reasoningField := choice.Delta.JSON.ExtraFields[g.reasoningContentField]
		reasoningContent := ""
		if reasoningField.Raw() != "" {
			_ = json.Unmarshal([]byte(reasoningField.Raw()), &reasoningContent)
		}
		if reasoningContent != "" {
			part := &ai.Part{Kind: ai.PartReasoning, ContentType: "plain/text", Text: reasoningContent}
			modelChunk.Content = append(modelChunk.Content, part)
			fullResponse.Message.Content = append(fullResponse.Message.Content, part)
		}
		// 普通文本
		if content := choice.Delta.Content; content != "" {
			part := ai.NewTextPart(content)
			modelChunk.Content = append(modelChunk.Content, part)
			fullResponse.Message.Content = append(fullResponse.Message.Content, part)
		}

		if err := handleChunk(ctx, modelChunk); err != nil {
			return nil, fmt.Errorf("generate error: callback error: %w", err)
		}

	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("generate error: stream error: %w", err)
	}

	return &fullResponse, nil
}

// generateComplete 对话补全
func (g *ModelGenerator) generateComplete(ctx context.Context, req *ai.ModelRequest) (*ai.ModelResponse, error) {
	completion, err := g.client.Chat.Completions.New(ctx, *g.request)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion: %w", err)
	}

	resp := &ai.ModelResponse{
		Request: req,
		Usage: &ai.GenerationUsage{
			InputTokens:         int(completion.Usage.PromptTokens),
			OutputTokens:        int(completion.Usage.CompletionTokens),
			ThoughtsTokens:      int(completion.Usage.CompletionTokensDetails.ReasoningTokens),
			CachedContentTokens: int(completion.Usage.PromptTokensDetails.CachedTokens),
			TotalTokens:         int(completion.Usage.TotalTokens),
		},
		Message: &ai.Message{
			Role: ai.RoleModel,
		},
	}

	if len(completion.Choices) == 0 {
		resp.FinishReason = ai.FinishReasonUnknown
		return resp, nil
	}

	choice := completion.Choices[0]

	switch choice.FinishReason {
	case "stop", "tool_calls":
		resp.FinishReason = ai.FinishReasonStop
	case "length":
		resp.FinishReason = ai.FinishReasonLength
	case "content_filter":
		resp.FinishReason = ai.FinishReasonBlocked
	case "function_call":
		resp.FinishReason = ai.FinishReasonOther
	default:
		resp.FinishReason = ai.FinishReasonUnknown
	}

	msgRaw := choice.Message.RawJSON()
	var msgRawMap map[string]any
	if err := json.Unmarshal([]byte(msgRaw), &msgRawMap); err != nil {
		return nil, fmt.Errorf("unmarshal choices[0].message error: %w", err)
	}

	// 思考内容
	if reasoningContent, ok := msgRawMap[g.reasoningContentField].(string); ok {
		resp.Message.Content = append(
			resp.Message.Content,
			&ai.Part{Kind: ai.PartReasoning, ContentType: "plain/text", Text: reasoningContent},
		)
	}

	// 普通文本
	if choice.Message.Content != "" {
		resp.Message.Content = append(resp.Message.Content, ai.NewTextPart(choice.Message.Content))
	}

	// 工具调用
	var toolRequestParts []*ai.Part
	for _, toolCall := range choice.Message.ToolCalls {
		args, err := jsonStringToMap(toolCall.Function.Arguments)
		if err != nil {
			return nil, err
		}
		toolRequestParts = append(toolRequestParts, ai.NewToolRequestPart(&ai.ToolRequest{
			Ref:   toolCall.ID,
			Name:  toolCall.Function.Name,
			Input: args,
		}))
	}
	if len(toolRequestParts) > 0 {
		resp.Message.Content = append(resp.Message.Content, toolRequestParts...)
		return resp, nil
	}

	return resp, nil
}

func concatenateContent(parts []*ai.Part) string {
	var content strings.Builder
	for _, part := range parts {
		if part.IsText() || part.IsData() {
			content.WriteString(part.Text)
		}
	}
	return content.String()
}

func concatenateReasoningContent(parts []*ai.Part) string {
	var content strings.Builder
	for _, part := range parts {
		if part.IsReasoning() {
			content.WriteString(part.Text)
		}
	}
	return content.String()
}

func convertToolCalls(content []*ai.Part) ([]openai.ChatCompletionMessageToolCallParam, error) {
	var toolCalls []openai.ChatCompletionMessageToolCallParam
	for _, p := range content {
		if !p.IsToolRequest() {
			continue
		}
		toolCall, err := convertToolCall(p)
		if err != nil {
			return nil, err
		}
		toolCalls = append(toolCalls, *toolCall)
	}
	return toolCalls, nil
}

func convertToolCall(part *ai.Part) (*openai.ChatCompletionMessageToolCallParam, error) {
	toolCallID := part.ToolRequest.Ref
	if toolCallID == "" {
		toolCallID = part.ToolRequest.Name
	}

	p := &openai.ChatCompletionMessageToolCallParam{
		ID: toolCallID,
		Function: openai.ChatCompletionMessageToolCallFunctionParam{
			Name: part.ToolRequest.Name,
		},
	}

	args, err := anyToJSONString(part.ToolRequest.Input)
	if err != nil {
		return nil, err
	}
	if part.ToolRequest.Input != nil {
		p.Function.Arguments = args
	}

	return p, nil
}

func jsonStringToMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed to parse json string %s: %w", jsonString, err)
	}
	return result, nil
}

func anyToJSONString(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal any to JSON string: data, %#v %w", data, err)
	}
	return string(jsonBytes), nil
}
