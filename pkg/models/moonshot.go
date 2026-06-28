package models

const (
	// MoonshotProviderName 月之暗面模型供应商名
	MoonshotProviderName = "moonshotai"
	// MoonshotBaseURL 月之暗面默认 API 地址
	MoonshotBaseURL = "https://api.moonshot.cn/v1"
)

var (
	KimiK26 = ModelConfig{
		Name:      "kimi-k2.6",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  6.5,
			Output: 27,
			Cached: 1.1,
		},
		ContextWindow: 256000,
		Score:         7,
	}
	KimiK25 = ModelConfig{
		Name:      "kimi-k2.5",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  4,
			Output: 21,
			Cached: 0.7,
		},
		ContextWindow: 256000,
		Score:         6,
	}
)

// MoonshotModels 建议的月之暗面模型
var MoonshotModels = []ModelConfig{
	KimiK26,
	KimiK25,
}
