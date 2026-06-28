package models

import "context"

type modelsContextKey struct{}

// ContextWithModels 返回携带指定模型配置的上下文
func ContextWithModels(ctx context.Context, m Models) context.Context {
	return context.WithValue(ctx, modelsContextKey{}, m)
}

// FromContext 从上下文获取模型配置
func FromContext(ctx context.Context) (Models, bool) {
	m, ok := ctx.Value(modelsContextKey{}).(Models)
	return m, ok
}
