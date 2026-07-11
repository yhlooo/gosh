package configui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/yhlooo/gosh/pkg/configs"
)

// ConfigItemDelegate 配置列表项
type ConfigItemDelegate struct{}

var _ list.ItemDelegate = ConfigItemDelegate{}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// Render 渲染
func (ConfigItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	content := ""
	switch typed := item.(type) {
	case StringListItem:
		content = string(typed)
	case ConfigItem:
		content = typed.Title()
	default:
		return
	}

	renderFn := itemStyle.Render
	if m.Index() == index {
		content = "> " + content
		renderFn = selectedItemStyle.Render
	}

	_, _ = fmt.Fprint(w, renderFn(content))
}

// Height 返回高度
func (ConfigItemDelegate) Height() int { return 1 }

// Spacing 返回项间行数
func (ConfigItemDelegate) Spacing() int { return 0 }

// Update 处理更新事件
func (ConfigItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// ConfigItem 配置项
type ConfigItem interface {
	list.Item
	SubView

	// Title 返回配置标题
	Title() string
}

// StringListItem 字符串列表项
type StringListItem string

var _ list.Item = StringListItem("")

// FilterValue 返回过滤值
func (s StringListItem) FilterValue() string {
	return string(s)
}

// ConfigPatch 配置修改
type ConfigPatch func(cfg *configs.Config)

// ConfigUpdateFn 配置更新方法
type ConfigUpdateFn func(patch ConfigPatch) error

// Form 表单
type Form struct {
	Key  string
	Name string
	Form *huh.Form
}

var _ ConfigItem = (*Form)(nil)

// Init 返回第一个命令
func (f Form) Init() tea.Cmd {
	return f.Form.Init()
}

// Update 处理更新事件
func (f Form) Update(msg tea.Msg) (SubView, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c", "esc":
			return nil, nil
		case "enter":
			// TODO: ...
			return nil, nil
		}
	}

	form, cmd := f.Form.Update(msg)
	if typedForm, ok := form.(*huh.Form); ok {
		f.Form = typedForm
	}
	return f, cmd
}

// View 返回显示内容
func (f Form) View() tea.View {
	return tea.NewView(f.Form.View())
}

// FilterValue 返回过滤值
func (f Form) FilterValue() string { return f.Key }

// Title 返回配置标题
func (f Form) Title() string { return f.Name }
