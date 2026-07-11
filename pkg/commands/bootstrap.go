package commands

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/gosh/pkg/bootstrap"
	"github.com/yhlooo/gosh/pkg/configs"
	"github.com/yhlooo/gosh/pkg/i18n"
)

// BootstrapOptions bootstrap 子命令
type BootstrapOptions struct {
	Force bool
}

// AddPFlags 绑定选项到命令行参数
func (opts *BootstrapOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&opts.Force, "force", opts.Force, i18n.T(MsgBootstrapOptsForceDesc))
}

// NewBootstrapOptions 创建默认 BootstrapOptions
func NewBootstrapOptions() BootstrapOptions {
	return BootstrapOptions{
		Force: false,
	}
}

// newBootstrapCommand 创建 bootstrap 子命令
func newBootstrapCommand() *cobra.Command {
	opts := NewBootstrapOptions()
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: i18n.T(MsgCmdShortDescBootstrap),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := configs.ConfigFromContext(ctx)
			cfgPath := configs.ConfigPathFromContext(ctx)

			if (len(cfg.ModelProviders) > 0 || cfg.DefaultModels.Primary != "") && !opts.Force {
				fmt.Println("\x1b[33m" + i18n.TContext(ctx, MsgBootstrapConfigAlreadyExists) + "\x1b[0m")
				return nil
			}

			// 配置
			bs := bootstrap.New()
			if _, err := tea.NewProgram(bs).Run(); err != nil {
				return fmt.Errorf("configure gosh error: %w", err)
			}

			// 保存
			if ok := bs.ApplyConfig(&cfg); !ok {
				return nil
			}
			if err := configs.SaveConfig(cfgPath, cfg); err != nil {
				return fmt.Errorf("save configuration error: %w", err)
			}

			return nil
		},
	}
	opts.AddPFlags(cmd.Flags())
	return cmd
}
