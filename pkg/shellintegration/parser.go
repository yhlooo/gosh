package shellintegration

import (
	"strconv"
	"strings"
	"sync"
)

// State 解析状态
type State int32

const (
	// StateGround 基础状态
	StateGround State = iota
	// StateEsc 刚接收到一个 ESC
	StateEsc
	// StateOSCPs 接收到 ESC ] 后，正在解析 Ps 数字
	StateOSCPs
	// StateOSCPt 接收到 ; 后，正在解析 Pt 文本
	StateOSCPt
	// StateOSCPtEsc Pt 中接收到 ESC，等待 ST 终止符的 \
	StateOSCPtEsc
)

const maxBuffSize = 64 << 10 // 64KB

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
	case StateOSCPs:
		return p.addOnOSCPs(c)
	case StateOSCPt:
		return p.addOnOSCPt(c)
	case StateOSCPtEsc:
		return p.addOnOSCPtEsc(c)
	default:
		panic("not implemented")
	}
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
		p.buff = []byte{c}
	}
	return nil
}

// addOnEsc 处理 StateEsc 下 Add
func (p *Parser) addOnEsc(c byte) Message {
	switch c {
	case ']':
		p.state = StateOSCPs
		p.buff = append(p.buff, c)
	default:
		// 非 OSC 序列（如 CSI），忽略并重置
		p.state = StateGround
		p.buff = nil
	}
	return nil
}

// addOnOSCPs 处理 StateOSCPs 下 Add
func (p *Parser) addOnOSCPs(c byte) Message {
	switch {
	case c >= '0' && c <= '9':
		p.buff = append(p.buff, c)
	case c == ';':
		p.state = StateOSCPt
		p.buff = append(p.buff, c)
	default:
		// 非法 Ps 字符，丢弃当前序列
		p.state = StateGround
		p.buff = nil
	}
	return nil
}

// addOnOSCPt 处理 StateOSCPt 下 Add
func (p *Parser) addOnOSCPt(c byte) Message {
	if len(p.buff) >= maxBuffSize {
		// 缓冲区溢出，丢弃当前序列
		p.state = StateGround
		p.buff = nil
		return nil
	}
	switch c {
	case '\x07': // BEL
		msg := p.parseMessage()
		p.state = StateGround
		p.buff = nil
		return msg
	case '\x1b': // ESC — 可能是 ST 终止符的前半部分
		p.state = StateOSCPtEsc
		p.buff = append(p.buff, c)
	default:
		p.buff = append(p.buff, c)
	}
	return nil
}

// addOnOSCPtEsc 处理 StateOSCPtEsc 下 Add
func (p *Parser) addOnOSCPtEsc(c byte) Message {
	switch c {
	case '\\': // ST 终止符 (\)
		// 去掉 buffer 末尾的 ESC（不是 Pt 内容）后再解析
		p.buff = p.buff[:len(p.buff)-1]
		msg := p.parseMessage()
		p.state = StateGround
		p.buff = nil
		return msg
	default:
		// 假警报，ESC 属于 Pt 内容，回退继续解析 Pt
		p.state = StateOSCPt
		p.buff = append(p.buff, c)
	}
	return nil
}

// parseMessage 从 buff 解析 OSC 序列并生成对应 Message
//
// buff 格式: ESC ] Ps ; Pt
func (p *Parser) parseMessage() Message {
	// 跳过 ESC ]
	data := p.buff[2:]

	// 定位分隔符 ;
	semiIdx := -1
	for i, b := range data {
		if b == ';' {
			semiIdx = i
			break
		}
	}
	if semiIdx == -1 {
		return nil
	}

	ps := string(data[:semiIdx])
	pt := string(data[semiIdx+1:])

	switch ps {
	case "0":
		return OSC0Message{Title: pt}
	case "1":
		return OSC1Message{Title: pt}
	case "2":
		return OSC2Message{Title: pt}
	case "133":
		return p.parseOSC133(pt)
	case "1337":
		return p.parseOSC1337(pt)
	default:
		return nil
	}
}

// parseOSC133 解析 OSC 133 的 Pt 部分
func (p *Parser) parseOSC133(pt string) Message {
	// Pt 格式: A | B | C | D | D;exitcode
	switch {
	case pt == "A":
		return OSC133AMessage{}
	case pt == "B":
		return OSC133BMessage{}
	case pt == "C":
		return OSC133CMessage{}
	case strings.HasPrefix(pt, "D"):
		exitCode := int32(0)
		if len(pt) > 1 && pt[1] == ';' {
			if v, err := strconv.ParseInt(pt[2:], 10, 32); err == nil {
				exitCode = int32(v)
			}
		}
		return OSC133DMessage{ExitCode: exitCode}
	default:
		return nil
	}
}

// parseOSC1337 解析 OSC 1337 的 Pt 部分
func (p *Parser) parseOSC1337(pt string) Message {
	key, value := pt, ""
	if idx := strings.IndexByte(pt, '='); idx >= 0 {
		key = pt[:idx]
		value = pt[idx+1:]
	}
	return OSC1337Message{Key: key, Value: value}
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
