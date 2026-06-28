package models

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
)

// OllamaOptions Ollama 选项
type OllamaOptions struct {
	// Ollama 服务端地址
	//
	// 默认 http://localhost:11434
	BaseURL string `json:"baseURL,omitempty"`
	// 模型响应超时时间，秒
	//
	// 默认 300
	Timeout int `json:"timeout,omitempty"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *OllamaOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = "http://localhost:11434"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 300
	}
}

// NewOllamaRegister 创建 Ollama 模型注册器
func NewOllamaRegister(opts OllamaOptions) *OllamaRegister {
	opts.Complete()
	return &OllamaRegister{
		Plugin: &ollama.Ollama{
			ServerAddress: opts.BaseURL,
			Timeout:       opts.Timeout,
		},
	}
}

// OllamaRegister Ollama 模型注册器
type OllamaRegister struct {
	Plugin *ollama.Ollama
	Models []ModelConfig
}

var _ ModelRegister = (*OllamaRegister)(nil)

// GenkitPlugin 获取对应 Genkit 插件
func (r *OllamaRegister) GenkitPlugin() api.Plugin {
	return r.Plugin
}

// RegisterModels 注册模型
func (r *OllamaRegister) RegisterModels(_ context.Context, g *genkit.Genkit) ([]ModelConfig, error) {
	var registeredModels []ModelConfig
	for _, modelConfig := range r.Models {
		m := r.Plugin.DefineModel(g, ollama.ModelDefinition{
			Name: modelConfig.Name,
			Type: "chat",
		}, &ai.ModelOptions{
			Label: modelConfig.Name,
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
				Tools:      true,
			},
		})

		registeredModel := modelConfig
		registeredModel.Name = m.Name()
		registeredModels = append(registeredModels, registeredModel)
	}

	return registeredModels, nil
}
