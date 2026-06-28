package models

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/gosh/pkg/genkitplugins/oai"
)

// OpenAICompatibleOptions OpenAI 兼容选项
type OpenAICompatibleOptions struct {
	// 供应商名
	Name string `json:"name"`
	// API 地址
	BaseURL string `json:"baseURL"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// NewOpenAICompatibleRegister 创建 OpenAI 兼容模型注册器
func NewOpenAICompatibleRegister(
	opts OpenAICompatibleOptions,
	defaultProvider, defaultBaseURL string,
	defaultModels []ModelConfig,
	ext OpenAICompatibleExtension,
) *OpenAICompatibleRegister {
	if opts.Name == "" {
		opts.Name = defaultProvider
	}
	if opts.BaseURL == "" {
		opts.BaseURL = defaultBaseURL
	}

	return &OpenAICompatibleRegister{
		Plugin: &oai.OpenAICompatible{
			Provider: opts.Name,
			BaseURL:  opts.BaseURL,
			APIKey:   opts.APIKey,
		},
		Models:        opts.Models,
		DefaultModels: defaultModels,
		Extension:     ext,
	}
}

// OpenAICompatibleExtension OpenAI 兼容接口扩展
type OpenAICompatibleExtension struct {
	// 三档思考程度字段
	// 关闭 / 中档 / 最高
	ReasoningEffortFields [3]map[string]any
	// 思考内容字段
	ReasoningContentField string
}

// DefaultOpenAIExtension 默认 OpenAI 扩展
var DefaultOpenAIExtension = OpenAICompatibleExtension{
	ReasoningEffortFields: [3]map[string]any{
		{
			"thinking": map[string]any{"type": "disabled"},
		},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "medium",
		},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "xhigh",
		},
	},
	ReasoningContentField: "reasoning_content",
}

// MoonshotOpenAIExtension 月之暗面 OpenAI 扩展
var MoonshotOpenAIExtension = OpenAICompatibleExtension{
	ReasoningEffortFields: [3]map[string]any{
		{
			"thinking": map[string]any{"type": "disabled"},
		},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "medium",
		},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "high",
		},
	},
	ReasoningContentField: "reasoning_content",
}

// DeepseekOpenAIExtension Deepseek OpenAI 扩展
var DeepseekOpenAIExtension = OpenAICompatibleExtension{
	ReasoningEffortFields: [3]map[string]any{
		{"thinking": map[string]any{"type": "disabled"}},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "high",
		},
		{
			"thinking":         map[string]any{"type": "enabled"},
			"reasoning_effort": "max",
		},
	},
	ReasoningContentField: "reasoning_content",
}

// QwenOpenAIExtension 千问 OpenAI 扩展
var QwenOpenAIExtension = OpenAICompatibleExtension{
	ReasoningEffortFields: [3]map[string]any{
		{"enable_thinking": false},
		{"enable_thinking": true},
		{"enable_thinking": true},
	},
	ReasoningContentField: "reasoning_content",
}

// OpenRouterOpenAIExtension OpenRouter OpenAI 扩展
var OpenRouterOpenAIExtension = OpenAICompatibleExtension{
	ReasoningEffortFields: [3]map[string]any{
		{"reasoning": map[string]any{"effort": "none"}},
		{"reasoning": map[string]any{"effort": "medium"}},
		{"reasoning": map[string]any{"effort": "xhigh"}},
	},
	ReasoningContentField: "reasoning",
}

// OpenAICompatibleRegister OpenAI 兼容模型注册器
type OpenAICompatibleRegister struct {
	Plugin *oai.OpenAICompatible
	Models []ModelConfig
	// 默认添加的模型
	DefaultModels []ModelConfig
	// OpenAI 兼容接口扩展
	Extension OpenAICompatibleExtension
}

var _ ModelRegister = (*OpenAICompatibleRegister)(nil)

// GenkitPlugin 获取对应 Genkit 插件
func (r *OpenAICompatibleRegister) GenkitPlugin() api.Plugin {
	return r.Plugin
}

// RegisterModels 注册模型
func (r *OpenAICompatibleRegister) RegisterModels(_ context.Context, g *genkit.Genkit) ([]ModelConfig, error) {
	registerModels := r.Models
	registeredModelFlags := map[string]struct{}{}
	for _, m := range r.Models {
		registeredModelFlags[m.Name] = struct{}{}
	}

	// 注册默认模型
	for _, m := range r.DefaultModels {
		if _, ok := registeredModelFlags[m.Name]; !ok {
			registerModels = append(registerModels, m)
		}
	}

	var registeredModels []ModelConfig
	for _, cfg := range registerModels {
		m := r.Plugin.DefineModel(g, oai.ModelOptions{
			ModelOptions: ai.ModelOptions{
				Label: cfg.Name,
				Supports: &ai.ModelSupports{
					Multiturn:  true,
					Tools:      true,
					SystemRole: true,
					Media:      true,
					ToolChoice: true,
				},
			},
			Reasoning:             cfg.Reasoning,
			ReasoningEffortFields: r.Extension.ReasoningEffortFields,
			ReasoningContentField: r.Extension.ReasoningContentField,
		})

		registeredModel := cfg
		registeredModel.Name = m.Name()
		registeredModels = append(registeredModels, registeredModel)
	}

	return registeredModels, nil
}
