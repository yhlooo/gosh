package controllers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/creack/pty"
	"github.com/go-logr/logr"
	"golang.org/x/term"

	"github.com/yhlooo/gosh/pkg/iotrace"
)

// Options 运行选项
type Options struct {
	// 启动 shell 的命令
	Command string
	// 启动 shell 的参数
	Args []string
	// shell 额外环境变量
	Env []string

	// 跟踪日志输出目录
	TraceLogDir string
}

// Validate 校验选项
func (opts *Options) Validate() error {
	if opts.Command == "" {
		return fmt.Errorf("%w: .Command is required", ErrInvalidArguments)
	}
	return nil
}

// New 创建控制器
func New(opts Options) (*Controller, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &Controller{
		opts: opts,
	}, nil
}

// Controller 控制器
type Controller struct {
	opts Options

	started atomic.Int32

	cmd    *exec.Cmd
	ptmx   *os.File
	output *os.File
}

const (
	TraceInputLogFile  = "input.log"
	TraceInputRawFile  = "input.raw"
	TraceOutputLogFile = "output.log"
	TraceOutputRawFile = "output.raw"
)

// Run 运行直到 shell 运行结束或 pty 关闭
//
// NOTE: 只能执行一次
func (ctl *Controller) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("controller")

	started := ctl.started.Add(1)
	if started > 1 {
		return fmt.Errorf("%w: controller already started", ErrAlreadyStarted)
	}

	// 构造执行 shell 命令
	ctl.cmd = exec.CommandContext(ctx, ctl.opts.Command, ctl.opts.Args...)
	if len(ctl.opts.Env) > 0 {
		ctl.cmd.Env = append(ctl.cmd.Env, os.Environ()...)
		ctl.cmd.Env = append(ctl.cmd.Env, ctl.opts.Env...)
	}

	// 设置输入输出流并启动 shell
	var err error
	ctl.ptmx, err = pty.Start(ctl.cmd)
	if err != nil {
		return fmt.Errorf("start %q error: %w", ctl.cmd.Args, err)
	}
	defer func() {
		_ = ctl.ptmx.Close()
		if ctl.cmd.Process != nil {
			_ = ctl.cmd.Process.Kill()
		}
	}()

	// 处理窗口大小变化信号
	winchCh := make(chan os.Signal, 1)
	signal.Notify(winchCh, syscall.SIGWINCH)
	go func() {
		for range winchCh {
			if err := pty.InheritSize(os.Stdin, ctl.ptmx); err != nil {
				logger.Error(err, "resize pty error")
			}
		}
	}()
	winchCh <- syscall.SIGWINCH
	defer func() {
		signal.Stop(winchCh)
		close(winchCh)
	}()

	ctl.output = os.Stdout

	ptyInW := io.Writer(WriterFn(ctl.handleInput))
	ptyOutW := io.Writer(WriterFn(ctl.handleOutput))
	if ctl.opts.TraceLogDir != "" {
		inTracer, err := iotrace.NewFileTracer(
			filepath.Join(ctl.opts.TraceLogDir, TraceInputLogFile),
			filepath.Join(ctl.opts.TraceLogDir, TraceInputRawFile),
		)
		if err != nil {
			return fmt.Errorf("create input tracer error: %w", err)
		}
		defer func() { _ = inTracer.Close() }()
		ptyInW = inTracer.TraceWriter(ptyInW)

		outTracer, err := iotrace.NewFileTracer(
			filepath.Join(ctl.opts.TraceLogDir, TraceOutputLogFile),
			filepath.Join(ctl.opts.TraceLogDir, TraceOutputRawFile),
		)
		if err != nil {
			return fmt.Errorf("create output tracer error: %w", err)
		}
		defer func() { _ = outTracer.Close() }()
		ptyOutW = outTracer.TraceWriter(ptyOutW)
	}

	// 设置输入流为 raw 格式
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("set stdin raw error: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() { _, _ = io.Copy(ptyInW, os.Stdin) }()
	_, _ = io.Copy(ptyOutW, ctl.ptmx)

	return nil
}
