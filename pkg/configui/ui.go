package configui

import (
	"strconv"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/yhlooo/gosh/pkg/configs"
)

// NewConfigUI 创建配置 UI
func NewConfigUI(cfg configs.Config, cfgPath string) *ConfigUI {
	ui := &ConfigUI{
		cfg:     &cfg,
		cfgPath: cfgPath,
	}

	cfgList := list.New(
		[]list.Item{
			Form{
				Key:  "defaultModels",
				Name: "Models",
				Form: huh.NewForm(huh.NewGroup(
					huh.NewSelect[string]().
						Title("Primary").
						OptionsFunc(ui.availableModels, &cfg.ModelProviders).
						Value(new(cfg.DefaultModels.Primary)),
					huh.NewSelect[string]().
						Title("Vision").
						OptionsFunc(ui.availableModels, &cfg.ModelProviders).
						Value(new(cfg.DefaultModels.Vision)),
					huh.NewSelect[*int]().
						Title("Reasoning Level").
						Options(
							huh.NewOption("Default", (*int)(nil)),
							huh.NewOption("Disable", new(0)),
							huh.NewOption("Low", new(1)),
							huh.NewOption("High", new(2)),
						).
						Value(new(cfg.DefaultModels.ReasoningLevel)),
				)),
			},
			Form{
				Key:  "language",
				Name: "Language",
				Form: huh.NewForm(huh.NewGroup(
					huh.NewSelect[string]().
						Title("Language").
						Options(
							huh.NewOption("中文", "zh"),
							huh.NewOption("English", "en"),
						).
						Value(new(cfg.Language)),
				)),
			},
			Form{
				Key:  "maxContextWindow",
				Name: "Max Context Window",
				Form: huh.NewForm(huh.NewGroup(
					huh.NewInput().
						Title("Max Context Window").
						Value(new(strconv.FormatInt(cfg.MaxContextWindow, 10))),
				)),
			},
		},
		ConfigItemDelegate{},
		60, 20,
	)
	cfgList.Title = "Config"
	cfgList.SetShowTitle(true)
	cfgList.SetShowFilter(false)
	cfgList.SetShowStatusBar(false)
	cfgList.SetShowHelp(true)

	ui.cfgList = cfgList

	return ui
}

// SubView 子视图
type SubView interface {
	// Init 返回第一个命令
	Init() tea.Cmd
	// Update 处理更新事件
	Update(msg tea.Msg) (SubView, tea.Cmd)
	// View 返回显示内容
	View() tea.View
}

// ConfigUI 配置 UI
type ConfigUI struct {
	cfg     *configs.Config
	cfgPath string

	cfgList  list.Model
	nestView SubView
}

var _ tea.Model = (*ConfigUI)(nil)

// Init 返回第一个命令
func (ui *ConfigUI) Init() tea.Cmd {
	return nil
}

// Update 处理更新事件
func (ui *ConfigUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if ui.nestView != nil {
		ui.nestView, cmd = ui.nestView.Update(msg)
		return ui, cmd
	}

	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "enter":
			if subView, ok := ui.cfgList.SelectedItem().(SubView); ok {
				ui.nestView = subView
				return ui, ui.nestView.Init()
			}
			return ui, nil
		case "ctrl+c", "esc":
			return ui, tea.Quit
		}
	}

	ui.cfgList, cmd = ui.cfgList.Update(msg)
	return ui, cmd
}

// View 返回显示内容
func (ui *ConfigUI) View() tea.View {
	if ui.nestView != nil {
		return ui.nestView.View()
	}
	return tea.NewView(ui.cfgList.View())
}

// availableModels 返回可用模型列表
func (ui *ConfigUI) availableModels() []huh.Option[string] {
	// TODO: 根据模型供应商配置实时确定
	return []huh.Option[string]{
		huh.NewOption("deepseek/deepseek-v4-pro", "deepseek/deepseek-v4-pro"),
		huh.NewOption("deepseek/deepseek-v4-flash", "deepseek/deepseek-v4-flash"),
		huh.NewOption("tencent-cloud/hy3", "tencent-cloud/hy3"),
	}
}
