package configs

import (
	"github.com/yhlooo/gosh/pkg/models"
)

// Config 配置
type Config struct {
	// 模型提供商配置
	ModelProviders []models.ModelProvider `json:"modelProviders,omitempty"`
	// 默认模型
	DefaultModels models.Models `json:"defaultModels,omitempty"`
	// 语言，可选 en, zh
	Language string `json:"language,omitempty"`
	// 最大上下文窗口
	// 默认 200K
	MaxContextWindow int64 `json:"maxContextWindow,omitempty"`
}
