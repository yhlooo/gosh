package models

const (
	// MinimaxProviderName MiniMax 模型供应商名
	MinimaxProviderName = "minimax"
	// MinimaxBaseURL MiniMax 默认 API 地址
	MinimaxBaseURL = "https://api.minimaxi.com/v1"
)

var MinimaxM27 = ModelConfig{
	Name:      "minimax-m2.7",
	Reasoning: true,
	Prices: ModelPrices{
		Input:  2.1,
		Output: 8.4,
		Cached: 0.42,
	},
	ContextWindow: 200000,
	Score:         5,
}

// MinimaxModels 建议的 MiniMax 模型
var MinimaxModels = []ModelConfig{
	MinimaxM27,
}
