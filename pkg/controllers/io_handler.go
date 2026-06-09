package controllers

import "io"

// WriterFn 实现了 io.Writer 的写函数
type WriterFn func(p []byte) (n int, err error)

var _ io.Writer = WriterFn(nil)

// Write 写
func (fn WriterFn) Write(p []byte) (n int, err error) {
	return fn(p)
}

// handleOutput 处理写 pty 的输出内容
func (ctl *Controller) handleOutput(p []byte) (n int, err error) {
	// TODO: 解析、 hook 输出
	return ctl.output.Write(p)
}

// handleInput 处理写到 pty 的输入内容
func (ctl *Controller) handleInput(p []byte) (n int, err error) {
	// TODO: hook 输入
	return ctl.ptmx.Write(p)
}
