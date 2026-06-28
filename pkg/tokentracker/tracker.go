package tokentracker

import (
	"context"
	"fmt"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"

	"github.com/yhlooo/gosh/pkg/models"
)

// NewTracker 创建 Token 跟踪器
func NewTracker(allModels []models.ModelConfig) *TokenTracker {
	prices := make(map[string]models.ModelPrices)
	for _, model := range allModels {
		prices[model.Name] = model.Prices
	}
	return &TokenTracker{
		prices: prices,
	}
}

// TokenTracker Token 跟踪器
type TokenTracker struct {
	lock sync.RWMutex

	totalUsage TokenUsage
	usages     map[string]*TokenUsage
	prices     map[string]models.ModelPrices
}

var million = decimal.New(1, 6)

// Summary 用量摘要
type Summary struct {
	TotalUsage TokenUsage      `json:"totalUsage"`
	TotalCost  decimal.Decimal `json:"totalCost"`
}

// TokenUsage Token 用量
type TokenUsage struct {
	// 总输入 Token
	InputTokens int64 `json:"inputTokens,omitempty"`
	// 总输出 Token
	OutputTokens int64 `json:"outputTokens,omitempty"`
	// 输入 Token 中命中缓存的 Token
	CacheReadTokens int64 `json:"cacheReadTokens,omitempty"`
}

// Summary 获取当前摘要
func (tracker *TokenTracker) Summary() Summary {
	tracker.lock.RLock()
	defer tracker.lock.RUnlock()

	cost := decimal.Zero
	for m, usage := range tracker.usages {
		prices, ok := tracker.prices[m]
		if !ok {
			continue
		}
		cost = cost.Add(
			decimal.NewFromFloat(prices.Input).
				Mul(decimal.NewFromInt(usage.InputTokens-usage.CacheReadTokens)).
				DivRound(million, 6),
		)
		cost = cost.Add(
			decimal.NewFromFloat(prices.Output).
				Mul(decimal.NewFromInt(usage.OutputTokens)).
				DivRound(million, 6),
		)
		cacheReadPrice := prices.Cached
		if cacheReadPrice == 0 {
			cacheReadPrice = prices.Input
		}
		cost = cost.Add(
			decimal.NewFromFloat(cacheReadPrice).
				Mul(decimal.NewFromInt(usage.CacheReadTokens)).
				DivRound(million, 6),
		)
	}

	return Summary{
		TotalUsage: tracker.totalUsage,
		TotalCost:  cost,
	}
}

// Middleware 返回作为中间件的 Token 跟踪器
func (tracker *TokenTracker) Middleware() ai.Middleware {
	return (*TokenTrackingMiddleware)(tracker)
}

// TokenTrackingMiddleware Token 跟踪中间件
type TokenTrackingMiddleware TokenTracker

var _ ai.Middleware = (*TokenTrackingMiddleware)(nil)

// Name 返回中间件名
func (tracker *TokenTrackingMiddleware) Name() string {
	return "TokenTracker"
}

// New 创建 hook
func (tracker *TokenTrackingMiddleware) New(_ context.Context) (*ai.Hooks, error) {
	if tracker == nil {
		return &ai.Hooks{}, nil
	}
	return &ai.Hooks{
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
			resp, err := next(ctx, params)
			if resp == nil || resp.Usage == nil {
				return resp, err
			}

			tracker.lock.Lock()
			defer tracker.lock.Unlock()

			modelName := ""
			if params != nil && params.Options != nil {
				modelName = params.Options.Model
			}
			if modelName == "" {
				modelName = "unknown"
			}

			logger := logr.FromContextOrDiscard(ctx)
			logger.V(1).Info(fmt.Sprintf(
				"token usage: %s i:%d o:%d c:%d",
				modelName, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.CachedContentTokens,
			))

			// 统计用量
			tracker.totalUsage.InputTokens += int64(resp.Usage.InputTokens)
			tracker.totalUsage.OutputTokens += int64(resp.Usage.OutputTokens)
			tracker.totalUsage.CacheReadTokens += int64(resp.Usage.CachedContentTokens)
			if tracker.usages == nil {
				tracker.usages = make(map[string]*TokenUsage)
			}
			if tracker.usages[modelName] == nil {
				tracker.usages[modelName] = &TokenUsage{}
			}
			tracker.usages[modelName].InputTokens += int64(resp.Usage.InputTokens)
			tracker.usages[modelName].OutputTokens += int64(resp.Usage.OutputTokens)
			tracker.usages[modelName].CacheReadTokens += int64(resp.Usage.CachedContentTokens)

			return resp, err
		},
	}, nil
}
