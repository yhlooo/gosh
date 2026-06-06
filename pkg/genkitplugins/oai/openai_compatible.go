package oai

import (
	"context"
	"net/http"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAICompatible OpenAI 兼容模型插件
type OpenAICompatible struct {
	Provider string
	BaseURL  string
	APIKey   string

	lock    sync.Mutex
	initted bool

	client *openai.Client
	models []string
}

var _ api.Plugin = (*OpenAICompatible)(nil)

// Name 返回插件名
func (d *OpenAICompatible) Name() string {
	if d.Provider != "" {
		return d.Provider
	}
	return "openai-compatible"
}

// Init 初始化插件
func (d *OpenAICompatible) Init(_ context.Context) []api.Action {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.initted {
		panic("plugin already initialized")
	}

	opts := []option.RequestOption{
		option.WithHTTPClient(http.DefaultClient),
		option.WithBaseURL(d.BaseURL),
	}
	if d.APIKey != "" {
		opts = append(opts, option.WithAPIKey(d.APIKey))
	}

	client := openai.NewClient(opts...)
	d.client = &client
	d.initted = true

	return nil
}

// ModelOptions 模型选项
type ModelOptions struct {
	ai.ModelOptions

	// 是否支持思考模式
	Reasoning bool
	// 三档思考程度字段
	// 关闭 / 中档 / 最高
	ReasoningEffortFields [3]map[string]any
	// 思考内容字段
	ReasoningContentField string
}

// Complete 使用默认值补全参数
func (opts *ModelOptions) Complete() {
	if opts.Reasoning {
		if opts.ReasoningEffortFields[0] == nil {
			opts.ReasoningEffortFields[0] = map[string]any{
				"reasoning_effort": "none",
			}
		}
		if opts.ReasoningEffortFields[1] == nil {
			opts.ReasoningEffortFields[1] = map[string]any{
				"reasoning_effort": "medium",
			}
		}
		if opts.ReasoningEffortFields[2] == nil {
			opts.ReasoningEffortFields[2] = map[string]any{
				"reasoning_effort": "xhigh",
			}
		}
	}
	if opts.ReasoningContentField == "" {
		opts.ReasoningContentField = "reasoning_content"
	}
}

// DefineModel 定义模型
func (d *OpenAICompatible) DefineModel(g *genkit.Genkit, opts ModelOptions) ai.Model {
	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.initted {
		panic("plugin not initialized")
	}

	model := d.defineModel(opts)
	genkit.DefineModel(g, model.Name(), &opts.ModelOptions, model.Generate)
	return model
}

// defineModel 定义模型
func (d *OpenAICompatible) defineModel(opts ModelOptions) ai.Model {
	opts.Complete()
	return ai.NewModel(api.NewName(d.Name(), opts.Label), &opts.ModelOptions, func(
		ctx context.Context,
		req *ai.ModelRequest,
		cb core.StreamCallback[*ai.ModelResponseChunk],
	) (*ai.ModelResponse, error) {
		generator := NewModelGenerator(d.client, req, opts).
			WithMessages(req.Messages).
			WithTools(req.Tools)

		resp, err := generator.Generate(ctx, req, cb)
		if err != nil {
			return nil, err
		}

		return resp, nil
	})
}
