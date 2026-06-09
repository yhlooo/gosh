package osc133

// Parser OSC-133 序列解析器
type Parser interface {
	// Next 读取到下一个开始标记
	Next() bool
	// Read 读取标记内容直到结束标记，然后返回 io.EOF
	Read(p []byte) (n int, err error)
}
