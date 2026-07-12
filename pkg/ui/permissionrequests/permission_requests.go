package permissionrequests

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// PermissionRequest 权限请求
type PermissionRequest struct {
	Title       string
	Description string

	allowed bool
}

var _ tea.Model = (*PermissionRequest)(nil)

// Init 返回第一个命令
func (ui *PermissionRequest) Init() tea.Cmd {
	return nil
}

// Update 处理更新事件
func (ui *PermissionRequest) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "up", "down", "tab":
			ui.allowed = !ui.allowed
		case "ctrl+c", "esc":
			ui.allowed = false
			return ui, tea.Quit
		case "enter":
			return ui, tea.Quit
		}
	}
	return ui, nil
}

// View 渲染视图
func (ui *PermissionRequest) View() tea.View {
	content := &strings.Builder{}

	content.WriteString("────────\n")
	content.WriteString(" " + ui.Title + "\n\n")
	content.WriteString("   " + strings.ReplaceAll(ui.Description, "\n", "\n   ") + "\n\n")

	content.WriteString(" Do you allow this operation?\n")
	if ui.allowed {
		content.WriteString(" ❯ \x1b[1m1.\x1b[0m Yes\n")
		content.WriteString("   \x1b[1m2.\x1b[0m No\n\n")
	} else {
		content.WriteString("   \x1b[1m1.\x1b[0m Yes\n")
		content.WriteString(" ❯ \x1b[1m2.\x1b[0m No\n\n")
	}
	content.WriteString(" \x1b[1mEsc to cancel · Enter to confirm\x1b[0m\n")

	return tea.NewView(content.String())
}

// Allowed 返回是否允许
func (ui *PermissionRequest) Allowed() bool {
	return ui.allowed
}
