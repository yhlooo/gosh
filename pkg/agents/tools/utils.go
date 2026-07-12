package tools

import (
	"encoding/json"

	"github.com/firebase/genkit/go/ai"

	"github.com/yhlooo/gosh/pkg/agents/generic"
)

// ToolInputWithPermissionRequest 带权限请求的工具输入
type ToolInputWithPermissionRequest interface {
	// PermissionRequest 获取工具输入对应权限请求
	PermissionRequest() generic.PermissionRequest
}

// PermissionRequestForToolRequest 获取工具请求对应权限请求
func PermissionRequestForToolRequest(req *ai.ToolRequest) generic.PermissionRequest {
	inRaw, _ := json.MarshalIndent(req.Input, "", "  ")

	switch req.Name {
	case ExecName:
		in := ExecInput{}
		_ = json.Unmarshal(inRaw, &in)
		return in.PermissionRequest()
	}

	return generic.PermissionRequest{
		Title:       req.Name,
		Description: string(inRaw),
	}
}
