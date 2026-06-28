package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yhlooo/gosh/pkg/i18n"
)

// newDebugCommand 创建 debug 子命令
func newDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "debug",
		Short:  i18n.T(MsgCmdShortDescDebug),
		Hidden: true,
	}

	cmd.AddCommand(
		newDebugParseANSICommand(),
	)

	return cmd
}

// newDebugParseANSICommand 创建 debug parse-ansi 子命令
func newDebugParseANSICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse-ansi",
		Short: i18n.T(MsgCmdShortDescDebugParseANSI),
		RunE: func(cmd *cobra.Command, args []string) error {
			parser := ansi.NewParser()
			parser.SetHandler(ansi.Handler{
				Print: func(r rune) {
					fmt.Printf("[Print] %q", r)
				},
				Execute: func(b byte) {
					fmt.Printf("[Execute] %q", b)
				},
				HandleCsi: func(cmd ansi.Cmd, params ansi.Params) {
					fmt.Printf(
						"[CSI] prefix=%q intermediate=%q final=%q params=%d",
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(), params,
					)
				},
				HandleEsc: func(cmd ansi.Cmd) {
					fmt.Printf(
						"[ESC] prefix=%q intermediate=%q final=%q",
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(),
					)
				},
				HandleDcs: func(cmd ansi.Cmd, params ansi.Params, data []byte) {
					fmt.Printf(
						"[DCS] prefix=%q intermediate=%q final=%q params=%d data=%q",
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(), params, string(data),
					)
				},
				HandleOsc: func(cmd int, data []byte) {
					fmt.Printf(
						"[OSC] cmd=%d data=%q",
						cmd, string(data),
					)
				},
				HandlePm: func(data []byte) {
					fmt.Printf("[PM] data=%q", string(data))
				},
				HandleApc: func(data []byte) {
					fmt.Printf("[APC] data=%q", string(data))
				},
				HandleSos: func(data []byte) {
					fmt.Printf("[SOS] data=%q", string(data))
				},
			})

			// 设置输入流为 raw 格式
			if term.IsTerminal(int(os.Stdin.Fd())) {
				oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
				if err != nil {
					panic(fmt.Errorf("set stdin raw error: %w", err))
				}
				defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
			}

			buff := make([]byte, 2048)
			for {
				n, err := os.Stdin.Read(buff)

				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}

				for _, b := range buff[:n] {
					parser.Advance(b)
					fmt.Printf(" (%s)\r\n", parser.StateName())
					switch b {
					case '\x03', '\x04':
						return nil
					}
				}
			}

			return nil
		},
	}
	cmd.AddCommand()
	return cmd
}
