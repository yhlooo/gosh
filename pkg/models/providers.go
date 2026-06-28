package models

import (
	"context"

	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
)

// ModelProvider 模型供应商配置
type ModelProvider struct {
	Ollama *OllamaOptions `json:"ollama,omitempty"`

	OpenAICompatible *OpenAICompatibleOptions `json:"openai-compatible,omitempty"`
	OpenRouter       *OpenAICompatibleOptions `json:"openrouter,omitempty"`
	OpenCode         *OpenAICompatibleOptions `json:"opencode,omitempty"`
	OpenCodeGo       *OpenAICompatibleOptions `json:"opencode-go,omitempty"`
	Deepseek         *OpenAICompatibleOptions `json:"deepseek,omitempty"`
	Qwen             *OpenAICompatibleOptions `json:"qwen,omitempty"`
	MoonshotAI       *OpenAICompatibleOptions `json:"moonshotai,omitempty"`
	ZAI              *OpenAICompatibleOptions `json:"z-ai,omitempty"`
	TencentCloud     *OpenAICompatibleOptions `json:"tencent-cloud,omitempty"`
	Minimax          *OpenAICompatibleOptions `json:"minimax,omitempty"`
}

// Register 返回对应的模型注册器
func (p ModelProvider) Register() ModelRegister {
	switch {
	case p.Ollama != nil:
		return NewOllamaRegister(*p.Ollama)
	case p.OpenAICompatible != nil:
		return NewOpenAICompatibleRegister(
			*p.OpenAICompatible,
			"", "",
			nil, DefaultOpenAIExtension,
		)
	case p.OpenRouter != nil:
		return NewOpenAICompatibleRegister(
			*p.OpenRouter,
			OpenRouterProviderName, OpenRouterBaseURL,
			OpenRouterModels, OpenRouterOpenAIExtension,
		)
	case p.OpenCode != nil:
		return NewOpenAICompatibleRegister(
			*p.OpenCode,
			OpenCodeProviderName, OpenCodeBaseURL,
			OpenCodeModels, DefaultOpenAIExtension,
		)
	case p.OpenCodeGo != nil:
		return NewOpenAICompatibleRegister(
			*p.OpenCodeGo,
			OpenCodeGoProviderName, OpenCodeGoBaseURL,
			OpenCodeGoModels, DefaultOpenAIExtension,
		)
	case p.Deepseek != nil:
		return NewOpenAICompatibleRegister(
			*p.Deepseek,
			DeepseekProviderName, DeepseekBaseURL,
			DeepseekModels, DeepseekOpenAIExtension,
		)
	case p.Qwen != nil:
		return NewOpenAICompatibleRegister(
			*p.Qwen,
			QwenProviderName, QwenBaseURL,
			QwenModels, QwenOpenAIExtension,
		)
	case p.MoonshotAI != nil:
		return NewOpenAICompatibleRegister(
			*p.MoonshotAI,
			MoonshotProviderName, MoonshotBaseURL,
			MoonshotModels, MoonshotOpenAIExtension,
		)
	case p.ZAI != nil:
		return NewOpenAICompatibleRegister(
			*p.ZAI,
			ZAIProviderName, ZAIBaseURL,
			ZAIModels, DefaultOpenAIExtension,
		)
	case p.TencentCloud != nil:
		return NewOpenAICompatibleRegister(
			*p.TencentCloud,
			TencentCloudProviderName, TencentCloudBaseURL,
			TencentCloudModels, DefaultOpenAIExtension,
		)
	case p.Minimax != nil:
		return NewOpenAICompatibleRegister(
			*p.Minimax,
			MinimaxProviderName, MinimaxBaseURL,
			MinimaxModels, DefaultOpenAIExtension,
		)
	}
	return nil
}

// ModelRegister 模型注册器
type ModelRegister interface {
	// GenkitPlugin 获取对应 Genkit 插件
	GenkitPlugin() api.Plugin
	// RegisterModels 注册模型
	RegisterModels(ctx context.Context, g *genkit.Genkit) ([]ModelConfig, error)
}

// ModelConfig 模型配置
type ModelConfig struct {
	// 模型名称
	Name string `json:"name"`
	// 供应商名称
	Provider string `json:"provider,omitempty"`

	// 是否支持推理、思考模式
	Reasoning bool `json:"reasoning,omitempty"`
	// 是否支持视觉、图片理解
	Vision bool `json:"vision,omitempty"`
	// 上下文窗口大小
	ContextWindow int64 `json:"contextWindow,omitempty"`

	// 价格信息
	Prices ModelPrices `json:"prices,omitempty"`

	// 效果评分，0-10
	Score int `json:"score,omitempty"`
}

// WithName 返回带指定名字的该模型
func (cfg ModelConfig) WithName(name string) ModelConfig {
	out := cfg
	out.Name = name
	return out
}

// WithPrices 返回带指定价格的该模型
func (cfg ModelConfig) WithPrices(prices ModelPrices) ModelConfig {
	out := cfg
	out.Prices = prices
	return out
}

// ModelPrices 价格信息
type ModelPrices struct {
	// 每百万输入 Token 价格
	Input float64 `json:"input,omitempty"`
	// 每百万输出 Token 价格
	Output float64 `json:"output,omitempty"`
	// 每百万缓存 Token 价格
	Cached float64 `json:"cached,omitempty"`
}
