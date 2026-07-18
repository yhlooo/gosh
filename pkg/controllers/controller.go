package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/creack/pty"
	"github.com/go-logr/logr"
	goterm "golang.org/x/term"

	"github.com/yhlooo/gosh/pkg/agents/generic"
	agentsgeneric "github.com/yhlooo/gosh/pkg/agents/generic"
	"github.com/yhlooo/gosh/pkg/shellintegrations"
	"github.com/yhlooo/gosh/pkg/term"
)

// Options 运行选项
type Options struct {
	// 启动 shell 的命令
	Command string
	// 启动 shell 的参数
	Args []string
	// shell 额外环境变量
	Env []string

	// 会话数据存储目录
	SessionDir string
	// 内置脚本存放目录
	ScriptDir string
	// 跟踪输入输出
	TraceIO bool

	// Agent
	Agent generic.Agent

	// 输入提示
	Prompt string
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
		opts:  opts,
		agent: opts.Agent,
	}, nil
}

// Controller 控制器
type Controller struct {
	opts Options

	started atomic.Int32

	// TODO: 输入操作里有些逻辑会操作输出相关对象，输出操作里也有操作输入相关对象，分两个锁不一定能锁住
	inputLock  sync.Mutex
	outputLock sync.Mutex

	ctx          context.Context
	logger       logr.Logger
	agent        generic.Agent
	cmd          *exec.Cmd
	agentPtmx    *os.File
	agentTTY     *os.File
	shellPtmx    *os.File
	input        *os.File
	output       *os.File
	inputParser  *ansi.Parser
	outputParser *ansi.Parser

	ready            bool
	inAgent          bool
	inAgentOutput    bool
	inExec           bool
	wroteExtraPrompt bool

	inputInterceptor io.Writer
	inputBuff        *bytes.Buffer
	agentInputBox    *term.InputBox
	deferOutput      *bytes.Buffer

	commandCollector *term.CommandCollector
}

// State 状态
type State uint32

const (
	// Shell 输入由 shell 处理
	Shell State = iota
	// Exec 输入由正在执行的命令处理
	Exec
	// AgentInput 写 Agent 输入
	AgentInput
	// AgentOutput Agent 输出中
	AgentOutput
)

var _ agentsgeneric.ShellController = (*Controller)(nil)

const (
	CommandLogFile     = "commands.jsonl"
	ExecOutputRawFile  = "exec_output.raw"
	TraceInputRawFile  = "input.raw"
	TraceOutputRawFile = "output.raw"
)

// Run 运行直到 shell 运行结束或 pty 关闭
//
// NOTE: 只能执行一次
func (ctl *Controller) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("controller")
	ctl.logger = logger

	started := ctl.started.Add(1)
	if started > 1 {
		return fmt.Errorf("%w: controller already started", ErrAlreadyStarted)
	}

	ctl.ctx = ctx

	var err error
	ctl.commandCollector, err = term.NewCommandCollector(
		ctx,
		filepath.Join(ctl.opts.SessionDir, CommandLogFile),
		filepath.Join(ctl.opts.SessionDir, ExecOutputRawFile),
	)
	if err != nil {
		return fmt.Errorf("create command collector error: %w", err)
	}

	ctl.inputBuff = &bytes.Buffer{}
	ctl.deferOutput = &bytes.Buffer{}

	// 初始化 Agent
	if err = ctl.agent.Initialize(ctx, generic.Options{
		TerminalType:             os.Getenv("TERM"),
		ShellController:          ctl,
		ChatOutputStreamHandler:  ctl.agentOutputHandler(),
		PermissionRequestHandler: ctl.handlePermissionRequest,
	}); err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	// 构造执行 shell 命令
	ctl.cmd = exec.CommandContext(ctx, ctl.opts.Command, ctl.opts.Args...)
	if len(ctl.opts.Env) > 0 {
		ctl.cmd.Env = append(ctl.cmd.Env, os.Environ()...)
		ctl.cmd.Env = append(ctl.cmd.Env, ctl.opts.Env...)
	}

	// 设置 agent 输入输出流
	ctl.agentPtmx, ctl.agentTTY, err = pty.Open()
	if err != nil {
		return fmt.Errorf("open agent pty error: %w", err)
	}
	defer func() {
		_ = ctl.agentPtmx.Close()
		_ = ctl.agentTTY.Close()
	}()

	// 设置 shell 输入输出流并启动 shell
	ctl.shellPtmx, err = pty.Start(ctl.cmd)
	if err != nil {
		return fmt.Errorf("start %q error: %w", ctl.cmd.Args, err)
	}
	defer func() {
		_ = ctl.shellPtmx.Close()
		if ctl.cmd.Process != nil {
			_ = ctl.cmd.Process.Kill()
		}
	}()

	// 处理窗口大小变化信号
	winchCh := make(chan os.Signal, 1)
	signal.Notify(winchCh, syscall.SIGWINCH)
	go func() {
		for range winchCh {
			if err := pty.InheritSize(os.Stdin, ctl.shellPtmx); err != nil {
				logger.Error(err, "resize pty error")
			}
			if err := pty.InheritSize(os.Stdin, ctl.agentPtmx); err != nil {
				logger.Error(err, "resize pty error")
			}
			rows, cols, _ := pty.Getsize(os.Stdin)
			ctl.commandCollector.Resize(rows, cols)
		}
	}()
	winchCh <- syscall.SIGWINCH
	defer func() {
		signal.Stop(winchCh)
		close(winchCh)
	}()

	ctl.input = os.Stdin
	ctl.output = os.Stdout
	ctl.agentInputBox = term.NewInputBox(ctl.output)

	ctl.inputParser = ansi.NewParser()
	ctl.inputParser.SetHandler(ctl.InputHandler().ParseHandler())
	inW := io.Writer(ctl.InputHandler())

	ctl.outputParser = ansi.NewParser()
	ctl.outputParser.SetHandler(ctl.ShellOutputHandler().ParseHandler())
	shellOutW := io.Writer(ctl.ShellOutputHandler())

	if ctl.opts.TraceIO {
		inRawFilePath := filepath.Join(ctl.opts.SessionDir, TraceInputRawFile)
		inRawFile, err := os.OpenFile(inRawFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("open input trace file %q error: %w", inRawFilePath, err)
		}
		defer func() { _ = inRawFile.Close() }()
		inW = io.MultiWriter(inW, inRawFile)

		outRawFilePath := filepath.Join(ctl.opts.SessionDir, TraceOutputRawFile)
		outRawFile, err := os.OpenFile(outRawFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("open output trace file %q error: %w", inRawFilePath, err)
		}
		defer func() { _ = outRawFile.Close() }()
		shellOutW = io.MultiWriter(shellOutW, outRawFile)
	}

	// 设置输入流为 raw 格式
	oldState, err := goterm.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("set stdin raw error: %w", err)
	}
	defer func() { _ = goterm.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		time.Sleep(500 * time.Millisecond)
		if ctl.ready {
			// 已经开启 Shell Integration 了
			return
		}

		// 开启 Shell Integration
		initCmd, err := shellintegrations.InitShell(filepath.Base(ctl.opts.Command), filepath.Join(ctl.opts.ScriptDir))
		if err != nil {
			ctl.logger.Error(err, "init init shell error")
			_, _ = ctl.output.WriteString(fmt.Sprintf("\x1b[31mInit shell error: %s\x1b0m\r\n", err.Error()))
			_ = ctl.shellPtmx.Close()
			return
		}
		if initCmd != "" {
			if _, err := ctl.shellPtmx.Write([]byte(initCmd + "\r")); err != nil {
				ctl.logger.Error(err, "send init command error")
				_, _ = ctl.output.WriteString(fmt.Sprintf(
					"\x1b[31mInit shell error: send init command error: %s\x1b0m\r\n",
					err.Error(),
				))
				_ = ctl.shellPtmx.Close()
				return
			}
		}

		time.Sleep(1000 * time.Millisecond)
		if !ctl.ready {
			// 开启 Shell Integration
			ctl.logger.Info("enable shell integration failed")
			_, _ = ctl.output.WriteString(fmt.Sprintf(
				"\x1b[31mInit shell error: enable shell integration failed\x1b0m\r\n",
			))
			_ = ctl.shellPtmx.Close()
			return
		}
	}()

	// 转发输入
	go func() { _, _ = io.Copy(inW, ctl.input) }()

	// 转发 agent 输出
	go func() { _, _ = io.Copy(ctl.output, ctl.agentPtmx) }()

	// 转发 shell 输出
	_, _ = io.Copy(shellOutW, ctl.shellPtmx)

	return nil
}

// CommandCollector 返回当前命令收集器
func (ctl *Controller) CommandCollector() *term.CommandCollector {
	return ctl.commandCollector
}

// State 返回当前状态
func (ctl *Controller) State() State {
	if ctl.inExec {
		return Exec
	}
	if ctl.inAgentOutput {
		return AgentOutput
	}
	if ctl.inAgent {
		return AgentInput
	}
	return Shell
}
