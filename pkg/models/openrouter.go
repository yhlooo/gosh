package models

const (
	// OpenRouterProviderName OpenRouter 模型供应商名
	OpenRouterProviderName = "openrouter"
	// OpenRouterBaseURL OpenRouter 默认 API 地址
	OpenRouterBaseURL = "https://openrouter.ai/api/v1"
)

var (
	Gemini31ProPreview = ModelConfig{
		Name:      "gemini-3.1-pro-preview",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  2, // <= 200K
			Output: 12,
			Cached: 0.2,
		},
		ContextWindow: 1050000,
		Score:         9,
	}
	GPT54 = ModelConfig{
		Name:      "gpt-5.4",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  2.5, // <= 272K
			Output: 15,
			Cached: 0.25,
		},
		ContextWindow: 1050000,
		Score:         10,
	}
	ClaudeSonnet46 = ModelConfig{
		Name:      "claude-sonnet-4.6",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  3,
			Output: 15,
			Cached: 0.3,
		},
		ContextWindow: 1000000,
		Score:         9,
	}
	ClaudeOpus47 = ModelConfig{
		Name:      "claude-opus-4.7",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  5,
			Output: 25,
			Cached: 0.5,
		},
		ContextWindow: 1000000,
		Score:         10,
	}
	Grok420 = ModelConfig{
		Name:      "grok-4.20",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  2,
			Output: 6,
			Cached: 0.2,
		},
		ContextWindow: 2000000,
		Score:         9,
	}
)

// OpenRouterModels 建议的 OpenRouter 模型
var OpenRouterModels = []ModelConfig{
	Gemini31ProPreview.WithName("google/gemini-3.1-pro-preview"),
	GPT54.WithName("openai/gpt-5.4"),
	ClaudeSonnet46.WithName("anthropic/claude-sonnet-4.6"),
	ClaudeOpus47.WithName("anthropic/claude-opus-4.7"),
	Grok420.WithName("x-ai/grok-4.20"),
}
