package shellintegration

import (
	"testing"
)

// feed 将字符串中的每个字节依次喂给 parser，返回最后一条非 nil 消息
func feed(p *Parser, s string) Message {
	var last Message
	for i := 0; i < len(s); i++ {
		if msg := p.Add(s[i]); msg != nil {
			last = msg
		}
	}
	return last
}

// feedAll 将字符串中的每个字节依次喂给 parser，返回所有非 nil 消息
func feedAll(p *Parser, s string) []Message {
	var msgs []Message
	for i := 0; i < len(s); i++ {
		if msg := p.Add(s[i]); msg != nil {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func TestOSC0_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]0;my title\x07")

	osc0, ok := msg.(OSC0Message)
	if !ok {
		t.Fatalf("expected OSC0Message, got %T", msg)
	}
	if osc0.Title != "my title" {
		t.Errorf("Title = %q, want %q", osc0.Title, "my title")
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestOSC1_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]1;tab title\x07")

	osc1, ok := msg.(OSC1Message)
	if !ok {
		t.Fatalf("expected OSC1Message, got %T", msg)
	}
	if osc1.Title != "tab title" {
		t.Errorf("Title = %q, want %q", osc1.Title, "tab title")
	}
}

func TestOSC2_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]2;hostname:pwd\x07")

	osc2, ok := msg.(OSC2Message)
	if !ok {
		t.Fatalf("expected OSC2Message, got %T", msg)
	}
	if osc2.Title != "hostname:pwd" {
		t.Errorf("Title = %q, want %q", osc2.Title, "hostname:pwd")
	}
}

func TestOSC133A_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;A\x07")

	if _, ok := msg.(OSC133AMessage); !ok {
		t.Fatalf("expected OSC133AMessage, got %T", msg)
	}
}

func TestOSC133B_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;B\x07")

	if _, ok := msg.(OSC133BMessage); !ok {
		t.Fatalf("expected OSC133BMessage, got %T", msg)
	}
}

func TestOSC133C_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;C\x07")

	if _, ok := msg.(OSC133CMessage); !ok {
		t.Fatalf("expected OSC133CMessage, got %T", msg)
	}
}

func TestOSC133D_WithExitCode(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;D;42\x07")

	d, ok := msg.(OSC133DMessage)
	if !ok {
		t.Fatalf("expected OSC133DMessage, got %T", msg)
	}
	if d.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", d.ExitCode)
	}
}

func TestOSC133D_ZeroExitCode(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;D;0\x07")

	d, ok := msg.(OSC133DMessage)
	if !ok {
		t.Fatalf("expected OSC133DMessage, got %T", msg)
	}
	if d.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", d.ExitCode)
	}
}

func TestOSC133D_NoExitCode(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;D\x07")

	d, ok := msg.(OSC133DMessage)
	if !ok {
		t.Fatalf("expected OSC133DMessage, got %T", msg)
	}
	if d.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", d.ExitCode)
	}
}

func TestOSC1337_BEL(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]1337;CurrentDir=/home\x07")

	m, ok := msg.(OSC1337Message)
	if !ok {
		t.Fatalf("expected OSC1337Message, got %T", msg)
	}
	if m.Key != "CurrentDir" {
		t.Errorf("Key = %q, want %q", m.Key, "CurrentDir")
	}
	if m.Value != "/home" {
		t.Errorf("Value = %q, want %q", m.Value, "/home")
	}
}

func TestOSC1337_NoValue(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]1337;RemoteHost\x07")

	m, ok := msg.(OSC1337Message)
	if !ok {
		t.Fatalf("expected OSC1337Message, got %T", msg)
	}
	if m.Key != "RemoteHost" {
		t.Errorf("Key = %q, want %q", m.Key, "RemoteHost")
	}
	if m.Value != "" {
		t.Errorf("Value = %q, want empty", m.Value)
	}
}

func TestST_Terminator(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]133;C\x1b\\")

	if _, ok := msg.(OSC133CMessage); !ok {
		t.Fatalf("expected OSC133CMessage via ST, got %T", msg)
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestST_Terminator_WithTitle(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]0;hello world\x1b\\")

	osc0, ok := msg.(OSC0Message)
	if !ok {
		t.Fatalf("expected OSC0Message via ST, got %T", msg)
	}
	if osc0.Title != "hello world" {
		t.Errorf("Title = %q, want %q", osc0.Title, "hello world")
	}
}

func TestCSI_Ignored(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b[31m")

	if msg != nil {
		t.Errorf("expected nil for CSI sequence, got %T", msg)
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestCSI_Ignored_ResetsState(t *testing.T) {
	// 先进入 Esc 状态，发送非 OSC 字符后应回到 Ground
	p := &Parser{}
	p.Add('\x1b')
	if p.State() != StateEsc {
		t.Fatalf("expected StateEsc, got %v", p.State())
	}
	p.Add('[')
	if p.State() != StateGround {
		t.Errorf("expected StateGround after CSI start, got %v", p.State())
	}
}

func TestNormalText_Ignored(t *testing.T) {
	p := &Parser{}
	msgs := feedAll(p, "hello\n")

	if len(msgs) != 0 {
		t.Errorf("expected 0 messages for normal text, got %d", len(msgs))
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestNormalText_PreservesGroundState(t *testing.T) {
	p := &Parser{}
	for _, c := range []byte("echo hello world") {
		msg := p.Add(c)
		if msg != nil {
			t.Fatalf("unexpected message for byte %q: %T", c, msg)
		}
		if p.State() != StateGround {
			t.Errorf("state changed to %v after byte %q", p.State(), c)
		}
	}
}

func TestEmptyPs_Discarded(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b];text\x07")

	if msg != nil {
		t.Errorf("expected nil for empty Ps, got %T", msg)
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestNonNumericPs_Discarded(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]abc;text\x07")

	if msg != nil {
		t.Errorf("expected nil for non-numeric Ps, got %T", msg)
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestUnknownPs_Discarded(t *testing.T) {
	p := &Parser{}
	msg := feed(p, "\x1b]999;text\x07")

	if msg != nil {
		t.Errorf("expected nil for unknown Ps, got %T", msg)
	}
	if p.State() != StateGround {
		t.Errorf("state = %v, want StateGround", p.State())
	}
}

func TestESC_InPt_NotST(t *testing.T) {
	// ESC 后跟非 \ 字符：ESC 属于 Pt 内容
	p := &Parser{}
	msg := feed(p, "\x1b]0;hello\x1bworld\x07")

	osc0, ok := msg.(OSC0Message)
	if !ok {
		t.Fatalf("expected OSC0Message, got %T", msg)
	}
	if osc0.Title != "hello\x1bworld" {
		t.Errorf("Title = %q, want %q", osc0.Title, "hello\x1bworld")
	}
}

func TestPartialESC_ThenReset(t *testing.T) {
	// 收到 ESC 进入 StateEsc，再收到普通字符回 Ground
	p := &Parser{}
	p.Add('\x1b')
	if p.State() != StateEsc {
		t.Fatalf("expected StateEsc, got %v", p.State())
	}
	p.Add('x')
	if p.State() != StateGround {
		t.Errorf("expected StateGround after non-OSC char, got %v", p.State())
	}
}

func TestMultipleSequences(t *testing.T) {
	p := &Parser{}
	input := "\x1b]133;A\x07hello\x1b]133;C\x07world\x1b]133;D;1\x07"
	msgs := feedAll(p, input)

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if _, ok := msgs[0].(OSC133AMessage); !ok {
		t.Errorf("msg[0] expected OSC133AMessage, got %T", msgs[0])
	}
	if _, ok := msgs[1].(OSC133CMessage); !ok {
		t.Errorf("msg[1] expected OSC133CMessage, got %T", msgs[1])
	}
	d, ok := msgs[2].(OSC133DMessage)
	if !ok {
		t.Fatalf("msg[2] expected OSC133DMessage, got %T", msgs[2])
	}
	if d.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", d.ExitCode)
	}
	if p.State() != StateGround {
		t.Errorf("final state = %v, want StateGround", p.State())
	}
}

func TestBufferResetOnComplete(t *testing.T) {
	p := &Parser{}
	feed(p, "\x1b]0;title\x07")

	if p.Buff() != "" {
		t.Errorf("Buff = %q, want empty after complete sequence", p.Buff())
	}
}

func TestBufferResetOnInvalid(t *testing.T) {
	p := &Parser{}
	feed(p, "\x1b[invalid")

	if p.Buff() != "" {
		t.Errorf("Buff = %q, want empty after invalid sequence", p.Buff())
	}
}

func TestMessageType(t *testing.T) {
	tests := []struct {
		msg  Message
		want Type
	}{
		{OSC0Message{Title: "t"}, OSC0},
		{OSC1Message{Title: "t"}, OSC1},
		{OSC2Message{Title: "t"}, OSC2},
		{OSC133AMessage{}, OSC133A},
		{OSC133BMessage{}, OSC133B},
		{OSC133CMessage{}, OSC133C},
		{OSC133DMessage{ExitCode: 1}, OSC133D},
		{OSC1337Message{Key: "k", Value: "v"}, OSC1337},
	}

	for _, tt := range tests {
		if got := tt.msg.Type(); got != tt.want {
			t.Errorf("%T.Type() = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestConcurrentAdd(t *testing.T) {
	p := &Parser{}
	done := make(chan struct{})

	// 多个 goroutine 并发喂不同的 OSC 序列
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				feed(p, "\x1b]133;A\x07")
				feed(p, "\x1b]133;C\x07")
				feed(p, "\x1b]0;t\x07")
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	// 最终状态应为 Ground
	if p.State() != StateGround {
		t.Errorf("final state = %v, want StateGround", p.State())
	}
}
