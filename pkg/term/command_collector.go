package term

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	headlessterm "github.com/danielgatis/go-headless-term"
)

// NewCommandCollector 创建命令收集器
func NewCommandCollector(logPath, execOutPath string) (*CommandCollector, error) {
	cc := &CommandCollector{
		parser:     ansi.NewParser(),
		buff:       &bytes.Buffer{},
		promptBuff: &bytes.Buffer{},
		cmdBuff: headlessterm.New(
			headlessterm.WithRecording(headlessterm.NewMemoryRecording()),
		),
	}

	// 设置 ANSI 解析处理器
	cc.parser.SetHandler(cc.parseHandler())

	// 打开命令记录日志文件
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file %q error: %w", logPath, err)
	}
	cc.cmdLogFile = logFile

	// 打开命令执行输出文件
	execOutFile, err := os.OpenFile(execOutPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		_ = cc.Close()
		return nil, fmt.Errorf("open exec output file %q error: %w", execOutPath, err)
	}
	cc.execOutFile = execOutFile

	return cc, nil
}

// OutputState 输出状态
type OutputState uint32

const (
	// OutputOthers 输出其它（无需关心的）内容
	OutputOthers OutputState = iota
	// OutputPrompt 输出提示符
	OutputPrompt
	// OutputCommand 输出命令内容
	OutputCommand
	// OutputCommandExec 输出命令执行输出
	OutputCommandExec
)

// CommandCollector 命令收集器
type CommandCollector struct {
	lock sync.RWMutex

	closed bool
	state  OutputState

	parser      *ansi.Parser
	buff        *bytes.Buffer
	promptBuff  *bytes.Buffer
	cmdBuff     *headlessterm.Terminal
	execOutFile *os.File
	cmdLogFile  *os.File
}

var _ io.Writer = (*CommandCollector)(nil)

// Write 写入 shell 输出
func (cc *CommandCollector) Write(p []byte) (n int, err error) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if cc.closed {
		return 0, fs.ErrClosed
	}

	if _, err := cc.execOutFile.Seek(0, io.SeekEnd); err != nil {
		return 0, fmt.Errorf("seek exec output file to end error: %w", err)
	}
	for i, c := range p {
		cc.parser.Advance(c)
		cc.buff.WriteByte(c)

		if cc.parser.State() != parser.GroundState {
			// 解析到特殊序列，暂不处理
			continue
		}

		switch cc.state {
		case OutputOthers:
		case OutputPrompt:
			_, _ = cc.promptBuff.Write(cc.buff.Bytes())
		case OutputCommand:
			_, _ = cc.cmdBuff.Write(cc.buff.Bytes())
		case OutputCommandExec:
			if n, err = cc.execOutFile.Write(cc.buff.Bytes()); err != nil {
				return i - cc.buff.Len() + n + 1, fmt.Errorf("write exec output file error: %w", err)
			}
		}

		cc.buff.Reset()
	}

	return len(p), nil
}

// Close 关闭
func (cc *CommandCollector) Close() error {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if cc.closed {
		return fs.ErrClosed
	}

	var errs []error
	if cc.execOutFile != nil {
		if err := cc.execOutFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if cc.cmdLogFile != nil {
		if err := cc.cmdLogFile.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	cc.closed = true

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Resize 设置终端大小
func (cc *CommandCollector) Resize(rows, cols int) {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	if cc.closed {
		return
	}
	cc.cmdBuff.Resize(rows, cols)
}

// CurrentPromptAndCommand 当前输入提示和输入的命令
func (cc *CommandCollector) CurrentPromptAndCommand() []byte {
	cc.lock.RLock()
	defer cc.lock.RUnlock()
	return append(bytes.Clone(cc.promptBuff.Bytes()), cc.cmdBuff.RecordedData()...)
}

// parseHandler 返回 ANSI 解析处理器
func (cc *CommandCollector) parseHandler() ansi.Handler {
	return ansi.Handler{
		HandleOsc: cc.handleOSC,
	}
}

// handleOSC 处理 OSC 序列
func (cc *CommandCollector) handleOSC(cmd int, data []byte) {
	switch cmd {
	case 133:
		if len(data) < 5 {
			return
		}
		switch string(data[:5]) {
		case "133;A":
			// 提示符开始
			cc.state = OutputPrompt
			cc.promptBuff.Reset()
		case "133;B":
			// 命令开始
			cc.state = OutputCommand
			cc.cmdBuff.ResetState()
			cc.cmdBuff.ClearRecording()
		case "133;C":
			// 命令开始执行
			cc.state = OutputCommandExec
			_, _ = cc.execOutFile.WriteString("\n---------------- Command Start ----------------\n")
			_, _ = cc.execOutFile.WriteString(cc.cmdBuff.String())
			_, _ = cc.execOutFile.WriteString("\n---------------- Command Executed ----------------\n")
		case "133;D":
			// 命令执行结束
			cc.state = OutputOthers
			exitCode := 0
			dataDivided := strings.Split(string(data), ";")
			if len(dataDivided) >= 3 {
				exitCode, _ = strconv.Atoi(dataDivided[2])
			}
			_, _ = cc.execOutFile.WriteString(fmt.Sprintf(
				"\n---------------- Command Finished (%d) ----------------\n",
				exitCode,
			))
		}
	}
}
