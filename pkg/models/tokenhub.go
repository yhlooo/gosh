package models

const (
	// TokenHubProviderName 腾讯云 TokenHub 模型供应商名
	TokenHubProviderName = "tokenhub"
	// TokenHubBaseURL 腾讯云 TokenHub 默认 API 地址
	TokenHubBaseURL = "https://tokenhub.tencentmaas.com/v1"
)

var (
	HY3 = ModelConfig{
		Name:      "hy3",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  1,
			Output: 4,
			Cached: 0.25,
		},
		ContextWindow: 256000,
		Score:         3,
	}
)

// TokenHubModels 腾讯云推荐模型
var TokenHubModels = []ModelConfig{
	HY3,
	DeepseekV4Pro,
	DeepseekV4Flash,
	KimiK26,
	KimiK25,
	GLM51,
	GLM5VTurbo,
	GLM5,
	MinimaxM27,
}
