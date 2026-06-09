package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/yhlooo/gosh/pkg/controllers"
	"github.com/yhlooo/gosh/pkg/i18n"
	"github.com/yhlooo/gosh/pkg/version"
)

// NewGlobalOptions 创建默认 GlobalOptions
func NewGlobalOptions() GlobalOptions {
	userHomeDir, _ := os.UserHomeDir()
	home := os.Getenv("GOSH_HOME")
	if home == "" {
		home = filepath.Join(userHomeDir, ".gosh")
	}
	return GlobalOptions{
		Verbosity: 0,
		Home:      home,
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
	// 开启调试模式
	Debug bool
	// 数据存储根目录
	Home string
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.Uint32VarP(&o.Verbosity, "verbose", "v", o.Verbosity, i18n.T(MsgGlobalOptsVerbosityDesc))
	fs.BoolVar(&o.Debug, "debug", o.Debug, i18n.T(MsgGlobalOptsDebugDesc))
}

type globalOptsContextKey struct{}

// ContextWithGlobalOptions 创建带全局选项的 context.Context
func ContextWithGlobalOptions(parent context.Context, opts GlobalOptions) context.Context {
	return context.WithValue(parent, globalOptsContextKey{}, opts)
}

// GlobalOptionsFromContext 从 ctx 获取全局选项
func GlobalOptionsFromContext(ctx context.Context) GlobalOptions {
	opts, _ := ctx.Value(globalOptsContextKey{}).(GlobalOptions)
	return opts
}

// NewOptions 创建默认 Options
func NewOptions() Options {
	defaultShell := os.Getenv("SHELL")
	if defaultShell == "" {
		defaultShell = "bash"
	}
	return Options{
		Shell: defaultShell,
	}
}

// Options 运行选项
type Options struct {
	Shell string
}

// AddPFlags 将选项绑定到命令行参数
func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Shell, "shell", "s", o.Shell, i18n.T(MsgOptsShellDesc))
}

// NewCommand 创建根命令
func NewCommand(name string) *cobra.Command {
	globalOpts := NewGlobalOptions()
	opts := NewOptions()

	var keylog *os.File
	cmd := &cobra.Command{
		Use:           name,
		Short:         i18n.T(MsgCmdShortDesc),
		Long:          i18n.T(MsgCmdLongDesc),
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := globalOpts.Validate(); err != nil {
				return err
			}
			ctx = ContextWithGlobalOptions(ctx, globalOpts)

			// 创建日志目录
			logDir := filepath.Join(globalOpts.Home, "logs")
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("create log directory %q error: %w", logDir, err)
			}

			// 初始化 logger
			logrusLogger := logrus.New()
			logrusLogger.SetOutput(&lumberjack.Logger{
				Filename:   filepath.Join(logDir, "gosh.log"),
				MaxSize:    500, // MB
				MaxBackups: 7,
				MaxAge:     30, // 天
			})
			switch globalOpts.Verbosity {
			case 0:
				logrusLogger.Level = logrus.InfoLevel
			case 1:
				logrusLogger.Level = logrus.DebugLevel
			default:
				logrusLogger.Level = logrus.TraceLevel
			}
			logger := logrusr.New(logrusLogger)
			ctx = logr.NewContext(ctx, logger)

			// 设置本地化器
			ctx = i18n.ContextWithLocalizer(ctx, i18n.NewLocalizer(i18n.GetEnvLanguage()))

			var err error
			keylog, err = setKeyLog()
			if err != nil {
				return fmt.Errorf("set tls key log error: %w", err)
			}

			cmd.SetContext(ctx)

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if keylog != nil {
				_ = keylog.Close()
			}
			return nil
		},
	}

	globalOpts.AddPFlags(cmd.PersistentFlags())
	opts.AddPFlags(cmd.Flags())

	cmd.AddCommand(
		newVersionCommand(),
	)

	return cmd
}

// run 运行
func run(ctx context.Context, opts Options) error {
	globalOpts := GlobalOptionsFromContext(ctx)

	logDir := ""
	if globalOpts.Debug {
		traceID := fmt.Sprintf("%x", rand.Uint64())
		_, _ = fmt.Fprintf(os.Stderr, "[DEBUG] trace id: %s\n", traceID)
		logDir = filepath.Join(globalOpts.Home, "logs", traceID)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("create trace log directory %q error: %w", logDir, err)
		}
	}

	ctl, err := controllers.New(controllers.Options{
		Command:     opts.Shell,
		Args:        nil,
		Env:         nil,
		TraceLogDir: logDir,
	})
	if err != nil {
		return err
	}

	return ctl.Run(ctx)
}

// setKeyLog 设置 TLS keylog
func setKeyLog() (*os.File, error) {
	keylog := os.Getenv("SSLKEYLOGFILE")
	if keylog == "" {
		return nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(keylog), 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(keylog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// 设置输出 keylog 文件
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,

			TLSClientConfig: &tls.Config{KeyLogWriter: f},
		},
	}

	return f, nil
}
