package shellintegrations

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

// 各 shell 开启 Shell Integration 的脚本
var (
	//go:embed shell_integration.bash
	bashScript []byte
	//go:embed shell_integration.zsh
	zshScript []byte
	//go:embed shell_integration.fish
	fishScript []byte
	//go:embed shell_integration.tcsh
	tcshScript []byte
	//go:embed shell_integration.xsh
	xonshScript []byte
)

// InitShell 初始化 shell 以开启 Shell Integration
func InitShell(shellName string, scriptDir string) (initCmd string, err error) {
	var scriptContent []byte
	scriptPath := ""
	switch shellName {
	case "bash":
		scriptContent = bashScript
		scriptPath = filepath.Join(scriptDir, "shell_integration.bash")
		initCmd = fmt.Sprintf("source %q", scriptPath)
	case "zsh":
		scriptContent = zshScript
		scriptPath = filepath.Join(scriptDir, "shell_integration.zsh")
		initCmd = fmt.Sprintf("source %q", scriptPath)
	case "fish":
		scriptContent = fishScript
		scriptPath = filepath.Join(scriptDir, "shell_integration.fish")
		initCmd = fmt.Sprintf("source %q", scriptPath)
	case "tcsh":
		scriptContent = tcshScript
		scriptPath = filepath.Join(scriptDir, "shell_integration.tcsh")
		initCmd = fmt.Sprintf("source %q", scriptPath)
	case "xonsh":
		scriptContent = xonshScript
		scriptPath = os.ExpandEnv("${HOME}/.config/xonsh/rc.d/gosh_shell_integration.xsh")
	default:
		// 不支持的 shell ，不做任何处理
		return "", nil
	}

	// 写初始化脚本
	if scriptPath != "" {
		if err := os.MkdirAll(filepath.Dir(scriptPath), 0755); err != nil {
			return "", fmt.Errorf("make dir %q error: %w", filepath.Dir(scriptPath), err)
		}
		if err := os.WriteFile(scriptPath, scriptContent, 0644); err != nil {
			return "", fmt.Errorf("write init script %q error: %w", scriptPath, err)
		}
	}

	// 返回初始化命令
	return initCmd, nil
}
