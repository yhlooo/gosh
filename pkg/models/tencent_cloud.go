package models

const (
	// TencentCloudProviderName 腾讯云 TokenHub 模型供应商名
	TencentCloudProviderName = "tencent-cloud"
	// TencentCloudBaseURL 腾讯云 TokenHub 默认 API 地址
	TencentCloudBaseURL = "https://tokenhub.tencentmaas.com/v1"
)

var (
	HY3Preview = ModelConfig{
		Name:      "hy3-preview",
		Reasoning: true,
		Prices: ModelPrices{
			Input:  2, // > 32K
			Output: 8,
			Cached: 0.8,
		},
		ContextWindow: 256000,
		Score:         3,
	}
)

// TencentCloudModels 腾讯云推荐模型
var TencentCloudModels = []ModelConfig{
	HY3Preview,
	DeepseekV4Pro,
	DeepseekV4Flash,
	KimiK26,
	KimiK25,
	GLM51,
	GLM5VTurbo,
	GLM5,
	MinimaxM27,
}
