package models

// Models 模型配置
type Models struct {
	// 主模型，用于回答用户问题
	Primary string `json:"primary,omitempty"`
	// 视觉模型，用于处理图片理解任务
	Vision string `json:"vision,omitempty"`
	// 思考级别
	ReasoningLevel *int `json:"reasoningLevel,omitempty"`
}

// GetPrimary 获取主模型
func (m Models) GetPrimary() string {
	return m.Primary
}

// GetVision 获取视觉模型
func (m Models) GetVision() string {
	if m.Vision != "" {
		return m.Vision
	}
	return m.Primary
}

// GetReasoningLevel 获取思考级别
func (m Models) GetReasoningLevel() int {
	if m.ReasoningLevel != nil {
		return *m.ReasoningLevel
	}
	return 2
}
