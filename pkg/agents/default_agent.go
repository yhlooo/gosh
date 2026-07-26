package agents

import (
	"context"
	"fmt"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/gosh/pkg/agents/generic"
	"github.com/yhlooo/gosh/pkg/agents/tools"
	"github.com/yhlooo/gosh/pkg/models"
	"github.com/yhlooo/gosh/pkg/term"
	"github.com/yhlooo/gosh/pkg/tokentracker"
)

const loggerName = "agent"

// GoshAgentOptions Agent 运行选项
type GoshAgentOptions struct {
	ModelProviders   []models.ModelProvider
	DefaultModels    models.Models
	MaxContextWindow int64
}

// Complete 使用默认值补全选项
func (opts *GoshAgentOptions) Complete() {
	if len(opts.ModelProviders) == 0 {
		opts.ModelProviders = append(opts.ModelProviders, models.ModelProvider{Ollama: &models.OllamaOptions{}})
	}
	if opts.MaxContextWindow == 0 {
		opts.MaxContextWindow = 200000
	}
}

// NewGoshAgent 创建 GoshAgent
func NewGoshAgent(opts GoshAgentOptions) *GoshAgent {
	opts.Complete()
	return &GoshAgent{
		opts:                   opts,
		cmdlineSuggestionCache: &CommandLineSuggestionCache{},
	}
}

// GoshAgent 是 gosh 内置的 Agent 的默认实现
type GoshAgent struct {
	opts        GoshAgentOptions
	genericOpts generic.Options

	g                *genkit.Genkit
	tokenTracker     *tokentracker.TokenTracker
	shellController  generic.ShellController
	commandCollector *term.CommandCollector

	chatTurnFlow             ChatTurnFlow
	genCmdlineSuggestionFlow GenCmdlineSuggestionFlow

	availableModels []models.ModelConfig
	availableTools  []ai.ToolRef
	allowTools      []string

	currentModels models.Models
	session       *Session

	cmdlineSuggestionCache *CommandLineSuggestionCache
}

// Session 会话
type Session struct {
	lock         sync.RWMutex
	cancelPrompt context.CancelFunc

	lastCmdIndex      int
	history           []*ai.Message
	lastContextWindow int64
}

// Initialize 初始化
func (a *GoshAgent) Initialize(ctx context.Context, opts generic.Options) error {
	logger := logr.FromContextOrDiscard(ctx).WithName(loggerName)

	if a.g != nil {
		return nil
	}

	a.shellController = opts.ShellController
	a.commandCollector = a.shellController.CommandCollector()

	// 创建 genkit 对象，注册模型
	a.g, a.availableModels = newGenkitWithModels(ctx, a.opts.ModelProviders, a.opts.DefaultModels)
	a.currentModels = a.opts.DefaultModels
	if a.currentModels.Primary == "" && len(a.availableModels) > 0 {
		a.currentModels.Primary = a.availableModels[0].Name
	}
	for _, m := range a.availableModels {
		logger.Info(fmt.Sprintf("registered model: %s", m.Name))
	}

	// 注册工具
	a.availableTools = append(a.availableTools, ai.ToolRef(tools.DefineToolGetHistory(a.g, a.commandCollector)))
	a.availableTools = append(a.availableTools, ai.ToolRef(tools.DefineToolReadExecOutput(a.g, a.commandCollector)))
	a.availableTools = append(a.availableTools, ai.ToolRef(tools.DefineToolExec(a.g, a.shellController)))

	for _, t := range a.availableTools {
		logger.Info(fmt.Sprintf("registered tool: %s", t.Name()))
	}

	// 始终允许只读操作
	a.allowTools = []string{
		tools.GetHistoryName,
		tools.ReadExecOutputName,
	}

	// 注册 flows
	a.chatTurnFlow = genkit.DefineFlow(a.g, "ChatTurn", a.handleChatTurn)
	a.genCmdlineSuggestionFlow = genkit.DefineFlow(a.g, "GenCmdlineSuggestion", a.handleGenCmdlineSuggestion)

	// 设置其它属性
	a.genericOpts = opts
	a.tokenTracker = tokentracker.NewTracker(a.availableModels)
	a.session = &Session{lastCmdIndex: a.commandCollector.LastCommandIndex()}

	return nil
}

// newGenkitWithModels 创建 genkit 对象并注册模型
func newGenkitWithModels(
	ctx context.Context,
	providers []models.ModelProvider,
	defaultModels models.Models,
) (*genkit.Genkit, []models.ModelConfig) {
	logger := logr.FromContextOrDiscard(ctx)

	// 确定插件
	var (
		modelRegisters []models.ModelRegister
		plugins        []api.Plugin
		modelConfigs   []models.ModelConfig
	)
	for _, p := range providers {
		modelRegister := p.Register()
		if modelRegister == nil {
			continue
		}
		plugins = append(plugins, modelRegister.GenkitPlugin())
		modelRegisters = append(modelRegisters, modelRegister)
	}

	genkitOpts := []genkit.GenkitOption{
		genkit.WithPlugins(plugins...),
	}
	if defaultModels.GetPrimary() != "" {
		genkitOpts = append(genkitOpts, genkit.WithDefaultModel(defaultModels.GetPrimary()))
	}
	g := genkit.Init(ctx, genkitOpts...)

	// 注册模型
	for i, reg := range modelRegisters {
		registeredModels, err := reg.RegisterModels(ctx, g)
		if err != nil {
			logger.Error(err, fmt.Sprintf("register model for provider %d error", i))
			continue
		}
		modelConfigs = append(modelConfigs, registeredModels...)
	}

	// 警告：如果没有配置任何模型
	if len(modelConfigs) == 0 {
		logger.Info("no models configured, please configure models in your config file")
	}

	return g, modelConfigs
}

// Cancel 取消当前正在处理的指令
func (a *GoshAgent) Cancel() error {
	a.session.lock.Lock()
	defer a.session.lock.Unlock()
	if a.session.cancelPrompt != nil {
		a.session.cancelPrompt()
	}
	return nil
}

// AvailableModels 获取可用模型列表
func (a *GoshAgent) AvailableModels() []models.ModelConfig {
	if len(a.availableModels) == 0 {
		return nil
	}

	ret := make([]models.ModelConfig, len(a.availableModels))
	copy(ret, a.availableModels)
	return ret
}
