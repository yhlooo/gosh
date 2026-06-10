package shellintegration

import "sync"

// State 解析状态
type State int32

const (
	// StateGround 基础状态
	StateGround State = iota
	// StateEsc 刚接收到一个 ESC
	StateEsc
	// TODO: 补全状态机状态
)

// Parser Shell 集成控制序列解析器
type Parser struct {
	lock sync.Mutex

	// 状态机当前状态
	state State
	// 当前缓冲的内容
	buff []byte
}

// Add 往解析状态机中添加一个字符
//
// 若完成一个序列解析则返回非 nil Message ，否则返回 nil
func (p *Parser) Add(c byte) Message {
	p.lock.Lock()
	defer p.lock.Unlock()

	switch p.state {
	case StateGround:
		return p.addOnGround(c)
	case StateEsc:
		return p.addOnEsc(c)
	}

	// TODO: 补全状态机实现
	panic("not implemented")
}

// State 返回解析状态机当前状态
func (p *Parser) State() State {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.state
}

// Buff 返回当前缓冲的内容
func (p *Parser) Buff() string {
	p.lock.Lock()
	defer p.lock.Unlock()
	return string(p.buff)
}

// addOnGround 处理 StateGround 下 Add
func (p *Parser) addOnGround(c byte) Message {
	switch c {
	case '\x1b':
		p.state = StateEsc
	}
	return nil
}

// addOnEsc 处理 StateEsc 下 Add
func (p *Parser) addOnEsc(c byte) Message {
	// TODO: ...
	panic("not implemented")
}

// Message 消息
type Message interface {
	// Type 返回消息类型
	Type() Type
}

// Type 控制序列类型
type Type string

const (
	// OSC0 同时设置窗口标题和标签页标题序列
	OSC0 Type = "OSC0"
	// OSC1 设置标签页标题序列
	OSC1 Type = "OSC1"
	// OSC2 设置窗口标题序列
	OSC2 Type = "OSC2"

	// OSC133A 提示符开始标记
	OSC133A Type = "OSC133A"
	// OSC133B 提示符结束标记
	OSC133B Type = "OSC133B"
	// OSC133C 命令执行前标记
	OSC133C Type = "OSC133C"
	// OSC133D 命令结束、返回码
	OSC133D Type = "OSC133D"

	// OSC1337 iTerm2 Shell Integration 扩展序列
	OSC1337 Type = "OSC1337"
)

// OSC0Message 同时设置窗口标题和标签页标题序列
//
// `ESC ] 0 ; title BEL`
type OSC0Message struct {
	Title string
}

// Type 返回消息类型
func (OSC0Message) Type() Type {
	return OSC0
}

// OSC1Message 设置标签页标题序列
//
// `ESC ] 1 ; title BEL`
type OSC1Message struct {
	Title string
}

// Type 返回消息类型
func (OSC1Message) Type() Type {
	return OSC1
}

// OSC2Message 设置窗口标题序列
//
// `ESC ] 2 ; title BEL`
type OSC2Message struct {
	Title string
}

// Type 返回消息类型
func (OSC2Message) Type() Type {
	return OSC2
}

// OSC133AMessage 提示符开始标记
//
// `ESC ] 133 ; A BEL`
type OSC133AMessage struct{}

// Type 返回消息类型
func (OSC133AMessage) Type() Type {
	return OSC133A
}

// OSC133BMessage 提示符结束标记
//
// `ESC ] 133 ; B BEL`
type OSC133BMessage struct{}

// Type 返回消息类型
func (OSC133BMessage) Type() Type {
	return OSC133B
}

// OSC133CMessage 命令执行前标记
//
// `ESC ] 133 ; C BEL`
type OSC133CMessage struct{}

// Type 返回消息类型
func (OSC133CMessage) Type() Type {
	return OSC133C
}

// OSC133DMessage 命令结束、返回码
//
// `ESC ] 133 ; D ; ExitCode BEL`
type OSC133DMessage struct {
	ExitCode int32
}

// Type 返回消息类型
func (OSC133DMessage) Type() Type {
	return OSC133D
}

// OSC1337Message iTerm2 Shell Integration 扩展序列
//
// `ESC ] 1337 ; Key = Value BEL`
type OSC1337Message struct {
	Key   string
	Value string
}

// Type 返回消息类型
func (OSC1337Message) Type() Type {
	return OSC1337
}
