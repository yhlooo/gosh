package commands

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/go-logr/logr"
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
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)

			logSuffix := "\r\n"
			if logger.V(1).Enabled() {
				logSuffix = "  "
			}
			startTime := time.Now()

			p := ansi.NewParser()

			lastPrintTime := 0
			var printBuff []rune
			p.SetHandler(ansi.Handler{
				Print: func(r rune) {
					printTime := int(time.Since(startTime) / time.Second)
					if len(printBuff) > 0 {
						if printTime == lastPrintTime && len(printBuff) < 80 {
							// 同行追加
							printBuff = append(printBuff, r)
							fmt.Printf("\x1b[A%04d  [Print] %q%s", printTime, string(printBuff), logSuffix)
							return
						}
						printBuff = nil
					}

					printBuff = append(printBuff, r)
					lastPrintTime = printTime
					fmt.Printf("%04d  [Print] %q%s", printTime, r, logSuffix)
				},
				Execute: func(b byte) {
					printBuff = nil
					fmt.Printf("%04d  [Execute] %q%s", time.Since(startTime)/time.Second, b, logSuffix)
				},
				HandleCsi: func(cmd ansi.Cmd, params ansi.Params) {
					printBuff = nil
					fmt.Printf(
						"%04d  [CSI] prefix=%q intermediate=%q final=%q params=%d%s",
						time.Since(startTime)/time.Second,
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(), params, logSuffix,
					)
				},
				HandleEsc: func(cmd ansi.Cmd) {
					printBuff = nil
					fmt.Printf(
						"%04d  [ESC] prefix=%q intermediate=%q final=%q%s",
						time.Since(startTime)/time.Second,
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(), logSuffix,
					)
				},
				HandleDcs: func(cmd ansi.Cmd, params ansi.Params, data []byte) {
					printBuff = nil
					fmt.Printf(
						"%04d  [DCS] prefix=%q intermediate=%q final=%q params=%d data=%q%s",
						time.Since(startTime)/time.Second,
						cmd.Prefix(), cmd.Intermediate(), cmd.Final(), params, string(data), logSuffix,
					)
				},
				HandleOsc: func(cmd int, data []byte) {
					printBuff = nil
					fmt.Printf(
						"%04d  [OSC] cmd=%d data=%q%s",
						time.Since(startTime)/time.Second, cmd, string(data), logSuffix,
					)
				},
				HandlePm: func(data []byte) {
					printBuff = nil
					fmt.Printf("%04d  [PM] data=%q%s", time.Since(startTime)/time.Second, string(data), logSuffix)
				},
				HandleApc: func(data []byte) {
					printBuff = nil
					fmt.Printf("%04d  [APC] data=%q%s", time.Since(startTime)/time.Second, string(data), logSuffix)
				},
				HandleSos: func(data []byte) {
					printBuff = nil
					fmt.Printf("%04d  [SOS] data=%q%s", time.Since(startTime)/time.Second, string(data), logSuffix)
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
					p.Advance(b)
					if logger.V(1).Enabled() {
						fmt.Printf("(%s)\r\n", p.StateName())
					}
					switch b {
					case '\x03', '\x04':
						return nil
					}
				}
			}
		},
	}
	cmd.AddCommand()
	return cmd
}
