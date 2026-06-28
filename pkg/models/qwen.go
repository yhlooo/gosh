package models

const (
	// QwenProviderName 通义千问模型供应商名
	QwenProviderName = "qwen"
	// QwenBaseURL 通义千问默认 API 地址（也可称 dashscope 、灵积）
	QwenBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

var (
	Qwen35_397b_a17b = ModelConfig{
		Name:      "qwen3.5-397b-a17b",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  3, // 128K-256K
			Output: 18,
		},
		ContextWindow: 254000,
		Score:         7,
	}
	Qwen36Plus = ModelConfig{
		Name:      "qwen3.6-plus",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  2, // <= 256K
			Output: 12,
			Cached: 0.2,
		},
		ContextWindow: 991000,
		Score:         8,
	}
	Qwen36_35b_a3b = ModelConfig{
		Name:      "qwen3.6-35b-a3b",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  1.8,
			Output: 10.8,
		},
		ContextWindow: 254000,
		Score:         4,
	}
	Qwen36Flash = ModelConfig{
		Name:      "qwen3.6-flash",
		Reasoning: true,
		Vision:    true,
		Prices: ModelPrices{
			Input:  1.2, // <= 256K
			Output: 7.2,
			Cached: 0.12,
		},
		ContextWindow: 991000,
		Score:         5,
	}
)

// QwenModels 建议的通义千问模型
var QwenModels = []ModelConfig{
	Qwen35_397b_a17b,
	Qwen36Plus,
	Qwen36_35b_a3b,
	Qwen36Flash,
}
