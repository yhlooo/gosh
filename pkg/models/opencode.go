package models

const (
	// OpenCodeProviderName OpenCode Zen 模型供应商名
	OpenCodeProviderName = "opencode"
	// OpenCodeBaseURL OpenCode Zen 默认 API 地址
	OpenCodeBaseURL = "https://opencode.ai/zen/v1"
	// OpenCodeGoProviderName OpenCode Go 模型供应商名
	OpenCodeGoProviderName = "opencode-go"
	// OpenCodeGoBaseURL OpenCode Go 默认 API 地址
	OpenCodeGoBaseURL = "https://opencode.ai/zen/go/v1"
)

// OpenCodeModels 建议的 OpenCode Zen 模型
var OpenCodeModels = []ModelConfig{
	ClaudeOpus47.WithName("claude-opus-4-7"),
	ClaudeSonnet46.WithName("claude-sonnet-4-6"),
	GPT54,
	GLM51,
	GLM5,
	Qwen36Plus,
	KimiK26,
	KimiK25,
	MinimaxM27,
}

// OpenCodeGoModels 建议的 OpenCode Go 模型
var OpenCodeGoModels = []ModelConfig{
	DeepseekV4Pro,
	DeepseekV4Flash,
	GLM51,
	GLM5,
	Qwen36Plus,
	KimiK26,
	KimiK25,
	MinimaxM27,
}
