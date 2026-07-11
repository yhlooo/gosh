package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/yhlooo/gosh/pkg/configs"
	"github.com/yhlooo/gosh/pkg/models"
)

// New 创建启动引导
func New() *Bootstrap {
	return &Bootstrap{}
}

// Bootstrap 启动引导
type Bootstrap struct {
	form *huh.Form

	abort         bool
	modelProvider ModelProviderOptions
	selectedModel string
	otherModel    string
}

var _ tea.Model = (*Bootstrap)(nil)

// Init 返回第一个命令
func (b *Bootstrap) Init() tea.Cmd {
	b.form = b.newForm()
	b.form.SubmitCmd = tea.Quit
	return b.form.Init()
}

// Update 处理更新事件
func (b *Bootstrap) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.String() {
		case "ctrl+c":
			b.abort = true
			return b, tea.Quit
		}
	}

	// 透传事件给表单
	m, cmd := b.form.Update(msg)
	if form, ok := m.(*huh.Form); ok {
		b.form = form
	}

	return b, cmd
}

// View 渲染视图
func (b *Bootstrap) View() tea.View {
	return tea.NewView(b.form.View())
}

// ApplyConfig 将引导配置合入到配置中
func (b *Bootstrap) ApplyConfig(cfg *configs.Config) bool {
	if b.abort {
		return false
	}

	providerName := ""
	switch b.modelProvider.Type {
	case models.OllamaProviderName:
		cfg.ModelProviders = []models.ModelProvider{{Ollama: &models.OllamaOptions{}}}
		providerName = models.OllamaProviderName
	case models.DeepseekProviderName:
		cfg.ModelProviders = []models.ModelProvider{{Deepseek: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.DeepseekProviderName
	case models.ZAIProviderName:
		cfg.ModelProviders = []models.ModelProvider{{ZAI: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.ZAIProviderName
	case models.MoonshotProviderName:
		cfg.ModelProviders = []models.ModelProvider{{MoonshotAI: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.MoonshotProviderName
	case models.MinimaxProviderName:
		cfg.ModelProviders = []models.ModelProvider{{Minimax: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.MinimaxProviderName
	case models.TokenHubProviderName:
		cfg.ModelProviders = []models.ModelProvider{{TokenHub: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.TokenHubProviderName
	case models.QwenProviderName:
		cfg.ModelProviders = []models.ModelProvider{{Qwen: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.QwenProviderName
	case models.OpenCodeProviderName:
		cfg.ModelProviders = []models.ModelProvider{{OpenCode: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.OpenCodeProviderName
	case models.OpenCodeGoProviderName:
		cfg.ModelProviders = []models.ModelProvider{{OpenCodeGo: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.OpenCodeGoProviderName
	case models.OpenRouterProviderName:
		cfg.ModelProviders = []models.ModelProvider{{OpenRouter: &models.OpenAICompatibleOptions{
			APIKey: b.modelProvider.APIKey,
		}}}
		providerName = models.OpenRouterProviderName
	case modelProviderTypeOpenAICompatible:
		cfg.ModelProviders = []models.ModelProvider{{OpenAICompatible: &models.OpenAICompatibleOptions{
			Name:    b.modelProvider.Name,
			BaseURL: b.modelProvider.BaseURL,
			APIKey:  b.modelProvider.APIKey,
		}}}
		providerName = b.modelProvider.Name
	}

	if providerName == "" {
		return false
	}

	modelName := b.selectedModel
	if b.selectedModel == "[others]" {
		modelName = b.otherModel
	}

	cfg.DefaultModels = models.Models{
		Primary: fmt.Sprintf("%s/%s", providerName, strings.TrimSpace(modelName)),
	}

	return true
}

func (b *Bootstrap) newForm() *huh.Form {
	return huh.NewForm(
		// 模型供应商选择
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(
					huh.NewOption("DeepSeek", models.DeepseekProviderName),
					huh.NewOption("Z.ai", models.ZAIProviderName),
					huh.NewOption("Moonshot AI", models.MoonshotProviderName),
					huh.NewOption("MiniMax", models.MinimaxProviderName),
					huh.NewOption("TokenHub (Tencent Cloud)", models.TokenHubProviderName),
					huh.NewOption("Qwen (Alibaba Cloud)", models.QwenProviderName),
					huh.NewOption("OpenCode Zen", models.OpenCodeProviderName),
					huh.NewOption("OpenCode Go", models.OpenCodeGoProviderName),
					huh.NewOption("OpenRouter", models.OpenRouterProviderName),
					huh.NewOption("Ollama", models.OllamaProviderName),
					huh.NewOption("OpenAI Compatible", modelProviderTypeOpenAICompatible),
				).
				Value(&b.modelProvider.Type),
		).
			Title("Model Provider"),

		// 模型供应商配置
		b.newKnownOpenAICompatibleGroup(models.DeepseekProviderName, "DeepSeek"),
		b.newKnownOpenAICompatibleGroup(models.ZAIProviderName, "Z.ai"),
		b.newKnownOpenAICompatibleGroup(models.MoonshotProviderName, "Moonshot AI"),
		b.newKnownOpenAICompatibleGroup(models.MinimaxProviderName, "MiniMax"),
		b.newKnownOpenAICompatibleGroup(models.TokenHubProviderName, "TokenHub (Tencent Cloud)"),
		b.newKnownOpenAICompatibleGroup(models.QwenProviderName, "Qwen (Alibaba Cloud)"),
		b.newKnownOpenAICompatibleGroup(models.OpenCodeProviderName, "OpenCode Zen"),
		b.newKnownOpenAICompatibleGroup(models.OpenCodeGoProviderName, "OpenCode Go"),
		b.newKnownOpenAICompatibleGroup(models.OpenRouterProviderName, "OpenRouter"),
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&b.modelProvider.Name).Validate(ValidateProviderName),
			huh.NewInput().Title("Base URL").Value(&b.modelProvider.BaseURL).Validate(ValidateURL),
			huh.NewInput().Title("API Key").Value(&b.modelProvider.APIKey),
		).
			Title("OpenAI Compatible").
			WithHideFunc(func() bool { return b.modelProvider.Type != modelProviderTypeOpenAICompatible }),

		// 模型
		huh.NewGroup(
			huh.NewSelect[string]().
				OptionsFunc(b.modelOptions, &b.modelProvider).
				Filtering(true).
				Height(10).
				Value(&b.selectedModel),
		).
			Title("Model"),
		huh.NewGroup(
			huh.NewInput().Value(&b.otherModel).Validate(ValidateModelName),
		).
			Title("Model").
			WithHideFunc(func() bool { return b.selectedModel != "[others]" }),
	)
}

// newKnownOpenAICompatibleGroup 创建已知 OpenAI 兼容模型供应商配置组
func (b *Bootstrap) newKnownOpenAICompatibleGroup(key, name string) *huh.Group {
	return huh.NewGroup(huh.NewInput().Title("API Key").Value(&b.modelProvider.APIKey)).
		Title(name).
		WithHideFunc(func() bool { return b.modelProvider.Type != key })
}

// modelOptions 可选模型选项
func (b *Bootstrap) modelOptions() []huh.Option[string] {
	var modelNames []string
	switch b.modelProvider.Type {
	case models.OllamaProviderName:
		modelNames, _ = models.ListOllamaModels(context.Background(), models.OllamaOptions{})
	case models.DeepseekProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.DeepseekBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.ZAIProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.ZAIBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.MoonshotProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.MoonshotBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.MinimaxProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.MinimaxBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.TokenHubProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.TokenHubBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.QwenProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.QwenBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.OpenCodeProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.OpenCodeBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.OpenCodeGoProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.OpenCodeGoBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case models.OpenRouterProviderName:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: models.OpenRouterBaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	case modelProviderTypeOpenAICompatible:
		modelNames, _ = models.ListOpenAICompatibleModels(context.Background(), models.OpenAICompatibleOptions{
			BaseURL: b.modelProvider.BaseURL,
			APIKey:  b.modelProvider.APIKey,
		})
	}

	var options []huh.Option[string]
	for _, modelName := range modelNames {
		options = append(options, huh.NewOption(modelName, modelName))
	}
	options = append(options, huh.NewOption("[Others]", "[others]"))

	return options
}

var providerNameRegexp = regexp.MustCompile("^[a-z][-a-z0-9]*$")

// ValidateProviderName 校验模型供应商名
func ValidateProviderName(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	if !providerNameRegexp.MatchString(s) {
		return fmt.Errorf("not match %q", providerNameRegexp.String())
	}
	return nil
}

// ValidateModelName 校验模型名
func ValidateModelName(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	return nil
}

// ValidateURL 校验 URL
func ValidateURL(s string) error {
	if s == "" {
		return errors.New("cannot be empty")
	}
	_, err := url.Parse(s)
	if err != nil {
		return err
	}
	return nil
}

const modelProviderTypeOpenAICompatible = "openai-compatible"

// ModelProviderOptions 模型供应商选项
type ModelProviderOptions struct {
	Type    string
	Name    string
	BaseURL string
	APIKey  string
}
