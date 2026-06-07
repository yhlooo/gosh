package iotrace

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

// NewFileTracer 创建输入输出跟踪器
func NewFileTracer(logPath, rawPath string) (*Tracer, error) {
	t := &Tracer{}

	rawWriter := io.Discard
	if rawPath != "" {
		rawFile, err := os.OpenFile(rawPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("open raw file %q error: %w", rawPath, err)
		}
		rawWriter = rawFile
	}
	t.raw = rawWriter

	logWriter := io.Discard
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			_ = t.Close()
			return nil, fmt.Errorf("open log file %q error: %w", logPath, err)
		}
		logWriter = logFile
	}
	t.logWriter = logWriter
	t.logger = slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{}))

	return t, nil
}

// Tracer 输入输出跟踪器
type Tracer struct {
	lock sync.Mutex

	raw       io.Writer
	logWriter io.Writer
	logger    *slog.Logger

	lineBuff []byte
}

var _ io.WriteCloser = (*Tracer)(nil)

// TraceWriter 跟踪写入指定 io.Writer 的内容
func (t *Tracer) TraceWriter(w io.Writer) io.Writer {
	return io.MultiWriter(w, t)
}

// Write 写
func (t *Tracer) Write(p []byte) (n int, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	_, _ = t.raw.Write(p)
	for _, c := range p {
		t.recordByte(c)
	}

	return len(p), nil
}

// recordByte 记录字节
func (t *Tracer) recordByte(c byte) {
	switch c {
	case '\n':
		// 换行刷一下缓冲区
		t.logger.Info(string(append(t.lineBuff, c)))
		t.lineBuff = nil
	default:
		t.lineBuff = append(t.lineBuff, c)
		if len(t.lineBuff) >= 4<<10 {
			// 缓冲区内容太多了，刷一下缓冲区
			t.logger.Info(string(t.lineBuff))
			t.lineBuff = nil
		}
	}
}

// Close 关闭
func (t *Tracer) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var errs []error

	if t.raw != nil {
		if closer, ok := t.raw.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if t.logWriter != nil {
		if closer, ok := t.logWriter.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
