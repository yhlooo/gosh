package term

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	headlessterm "github.com/danielgatis/go-headless-term"
	"github.com/go-logr/logr"
)

// NewCommandCollector 创建命令收集器
func NewCommandCollector(ctx context.Context, logPath, execOutPath string) (*CommandCollector, error) {
	cc := &CommandCollector{
		logger:     logr.FromContextOrDiscard(ctx).WithName("command-collector"),
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
	cc.cmdLogEncoder = json.NewEncoder(logFile)

	// 打开命令执行输出文件
	execOutFile, err := os.OpenFile(execOutPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		_ = cc.Close()
		return nil, fmt.Errorf("open exec output file %q error: %w", execOutPath, err)
	}
	cc.execOutFilePath = execOutPath
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
	lock   sync.RWMutex
	logger logr.Logger

	closed bool
	state  OutputState

	cmdIndex  int
	histories []CommandRecord
	curCmd    *CommandRecord

	parser     *ansi.Parser
	buff       *bytes.Buffer
	promptBuff *bytes.Buffer
	cmdBuff    *headlessterm.Terminal

	execOutFilePath string
	execOutFile     *os.File

	cmdLogEncoder *json.Encoder
	cmdLogFile    *os.File

	waitNextCommandChannels []chan CommandRecord
}

var _ io.Writer = (*CommandCollector)(nil)

// CommandRecord 命令执行记录
type CommandRecord struct {
	// 命令序号
	Index int `json:"index"`
	// 命令行
	CommandLine string `json:"commandLine"`
	// 开始时间
	StartTime time.Time `json:"startTime"`
	// 结束时间
	EndTime *time.Time `json:"endTime,omitempty"`
	// 命令执行退出码
	ExitCode *int `json:"exitCode,omitempty"`
	// 执行输出开始在日志中的字节偏移（含）
	ExecOutputStart int64 `json:"execOutputStart"`
	// 执行输出结束在日志中的字节偏移（不含）
	ExecOutputEnd *int64 `json:"execOutputEnd,omitempty"`
}

// Write 写入 shell 输出
func (cc *CommandCollector) Write(p []byte) (n int, err error) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if cc.closed {
		return 0, fs.ErrClosed
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

// CurrentPromptAndCommand 当前输入提示和未提交的命令
func (cc *CommandCollector) CurrentPromptAndCommand() []byte {
	cc.lock.RLock()
	defer cc.lock.RUnlock()
	return append(bytes.Clone(cc.promptBuff.Bytes()), cc.cmdBuff.RecordedData()...)
}

// CurrentCommand 返回当前未提交的命令
func (cc *CommandCollector) CurrentCommand() string {
	return cc.cmdBuff.String()
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
			cc.handleExecOutputStart()

		case "133;D":
			// 命令执行结束
			cc.state = OutputOthers
			cc.handleExecOutputEnd(data)
		}
	}
}

// handleExecOutputStart 处理命令执行输出开始
func (cc *CommandCollector) handleExecOutputStart() {
	_, _ = cc.execOutFile.WriteString("-------- CommandLine --------\n")
	_, _ = cc.execOutFile.WriteString(cc.cmdBuff.String())
	_, _ = cc.execOutFile.WriteString("\n-------- ExecOutput --------\n")

	stat, err := cc.execOutFile.Stat()
	if err != nil {
		cc.logger.Error(err, "stat exec output file error")
		cc.curCmd = nil
		return
	}
	cc.curCmd = &CommandRecord{
		CommandLine:     cc.cmdBuff.String(),
		StartTime:       time.Now(),
		ExecOutputStart: stat.Size(),
	}
}

// handleExecOutputEnd 处理命令执行输出结束
func (cc *CommandCollector) handleExecOutputEnd(data []byte) {
	// 提取退出码
	var exitCode *int
	dataDivided := strings.Split(string(data), ";") // 133;D;<exitCode>
	if len(dataDivided) >= 3 {
		exitCodeInt, err := strconv.Atoi(dataDivided[2])
		if err != nil {
			cc.logger.Error(err, "parse exit code error")
		} else {
			exitCode = &exitCodeInt
		}
	}

	if cc.curCmd == nil {
		// 没有记录命令开始，忽略
		return
	}

	// 记录命令执行历史
	curCmd := cc.curCmd
	cc.curCmd = nil
	stat, err := cc.execOutFile.Stat()
	if err != nil {
		cc.logger.Error(err, "stat exec output file error")
		return
	}
	curCmd.Index = cc.cmdIndex
	cc.cmdIndex++
	curCmd.ExitCode = exitCode
	curCmd.ExecOutputEnd = new(stat.Size())
	curCmd.EndTime = new(time.Now())
	_ = cc.cmdLogEncoder.Encode(curCmd)
	cc.histories = append(cc.histories, *curCmd)

	// 输出命令结束标志
	showingExitCode := ""
	if exitCode != nil {
		showingExitCode = fmt.Sprintf("(%d)", *exitCode)
	}
	_, _ = cc.execOutFile.WriteString(fmt.Sprintf("\n-------- Exit%s --------\n", showingExitCode))

	// 通知订阅方
	for _, ch := range cc.waitNextCommandChannels {
		select {
		case ch <- *curCmd:
		default:
			cc.logger.Info("WARN wait next command channel busy, skip")
		}
		close(ch)
	}
	cc.waitNextCommandChannels = nil
}

// WaitNextCommand 等待下一个命令执行完成
//
// 下一个命令执行完成后通过 ch 发送命令执行记录，随后关闭 ch
func (cc *CommandCollector) WaitNextCommand(ch chan CommandRecord) {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	cc.waitNextCommandChannels = append(cc.waitNextCommandChannels, ch)
}

// ListLastNCommands 列出最近 n 个命令
//
// n 最大值为 100 ，记录的命令数少于 n 时返回命令数可以小于 n ，结果按命令开始时间升序排列
func (cc *CommandCollector) ListLastNCommands(n int) []CommandRecord {
	cc.lock.RLock()
	defer cc.lock.RUnlock()

	if n > 100 {
		n = 100
	}
	if cc.curCmd != nil {
		n--
	}
	if n < 0 {
		n = 0
	}
	if n > len(cc.histories) {
		n = len(cc.histories)
	}

	var ret []CommandRecord

	if n > 0 {
		ret = make([]CommandRecord, n)
		copy(ret, cc.histories[len(cc.histories)-n:])
	}

	if cc.curCmd != nil {
		last := *cc.curCmd
		last.Index = cc.cmdIndex // 设置临时序号
		ret = append(ret, last)
	}

	return ret
}

// ReadExecOutput 读命令执行输出
//
// 读取指定命令的输出，从第 offset+1 个字节起读最多 limit 个字节
// limit 最大值为 1Mi ，在剩余内容不足 limit 时返回内容大小可以小于 limit
func (cc *CommandCollector) ReadExecOutput(index int, offset, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, nil
	}
	if limit > 1<<20 {
		limit = 1 << 20
	}
	if offset < 0 {
		offset = 0
	}

	cc.lock.RLock()
	defer cc.lock.RUnlock()

	var cmd *CommandRecord
	if index > cc.cmdIndex || index < cc.cmdIndex-len(cc.histories) {
		return nil, fmt.Errorf("no command record: %d", index)
	}
	if index == cc.cmdIndex {
		if cc.curCmd == nil {
			return nil, fmt.Errorf("no command record: %d", index)
		}
		cmd = cc.curCmd
	} else {
		cmd = &cc.histories[index-cc.cmdIndex+len(cc.histories)]
	}

	if cmd.ExecOutputEnd != nil {
		size := *cmd.ExecOutputEnd - cmd.ExecOutputStart
		if offset >= size {
			return nil, nil
		}
		if limit > size-offset {
			limit = size - offset
		}
	}

	ret := make([]byte, limit)
	n, err := cc.execOutFile.ReadAt(ret, cmd.ExecOutputStart+offset)
	return ret[:n], err
}

// LastCommandIndex 最后一个开始执行的命令序号
func (cc *CommandCollector) LastCommandIndex() int {
	cc.lock.RLock()
	defer cc.lock.RUnlock()
	if cc.curCmd == nil {
		return cc.cmdIndex - 1
	}
	return cc.cmdIndex
}

// State 返回当前输出状态
func (cc *CommandCollector) State() OutputState {
	cc.lock.RLock()
	defer cc.lock.RUnlock()
	return cc.state
}

// CursorPos 获取当前光标位置
func (cc *CommandCollector) CursorPos() (row, col int) {
	return cc.cmdBuff.CursorPos()
}

// IsAtLineEnd 判断光标是否正在行尾
func (cc *CommandCollector) IsAtLineEnd() bool {
	row, col := cc.cmdBuff.CursorPos()
	cols := cc.cmdBuff.Cols()
	for i := col; i < cols; i++ {
		cell := cc.cmdBuff.Cell(row, i)
		if cell != nil && cell.Char != 0 && cell.Char != ' ' {
			return false
		}
	}
	return true
}
