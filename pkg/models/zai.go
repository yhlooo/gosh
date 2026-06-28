package models

const (
	// ZAIProviderName 智谱模型供应商名
	ZAIProviderName = "z-ai"
	// ZAIBaseURL 智谱默认 API 地址
	ZAIBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
)

var (
	GLM51 = ModelConfig{
		Name:      "glm-5.1",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  8, // > 32K
			Output: 28,
			Cached: 2,
		},
		ContextWindow: 200000,
		Score:         7,
	}
	GLM5VTurbo = ModelConfig{
		Name:      "glm-5v-turbo",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  7, // > 32K
			Output: 26,
			Cached: 1.8,
		},
		ContextWindow: 200000,
		Score:         6,
	}
	GLM5 = ModelConfig{
		Name:      "glm-5",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  6, // > 32K
			Output: 22,
			Cached: 1.5,
		},
		ContextWindow: 200000,
		Score:         6,
	}
)

// ZAIModels 建议的智谱 AI 模型
var ZAIModels = []ModelConfig{
	GLM51,
	GLM5VTurbo,
	GLM5,
}
