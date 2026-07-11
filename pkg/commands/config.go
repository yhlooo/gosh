package commands

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/yhlooo/gosh/pkg/configs"
	"github.com/yhlooo/gosh/pkg/configui"
	"github.com/yhlooo/gosh/pkg/i18n"
)

// newConfigCommand 创建 config 子命令
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "config",
		Short:  i18n.T(MsgCmdShortDescConfig),
		Hidden: true, // TODO: 尚未实现，暂时隐藏
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := configs.ConfigFromContext(ctx)
			cfgPath := configs.ConfigPathFromContext(ctx)
			p := tea.NewProgram(configui.NewConfigUI(cfg, cfgPath))
			_, err := p.Run()
			return err
		},
	}
	return cmd
}
