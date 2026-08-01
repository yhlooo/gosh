**[简体中文](README_CN.md)** | [English](README.md)

---

![GitHub License](https://img.shields.io/github/license/yhlooo/gosh)
[![GitHub Release](https://img.shields.io/github/v/release/yhlooo/gosh)](https://github.com/yhlooo/gosh/releases/latest)
[![release](https://github.com/yhlooo/gosh/actions/workflows/release.yaml/badge.svg)](https://github.com/yhlooo/gosh/actions/workflows/release.yaml)

# Gosh - LLM 增强的 Shell

> **🏗️ 该项目还处于较早期阶段。**

Gosh 不是类似 Zsh / Bash 的真正的 Shell ，而是它们的增强包装。你可以像往常一样使用你熟悉的 Shell ，但是它比过去更好了。

## 快速开始

1. **安装 Gosh**

   **脚本安装：**

   ```bash
   curl -L https://raw.githubusercontent.com/yhlooo/gosh/refs/heads/main/scripts/install.sh | bash
   ```

   脚本将 `gosh` 安装到 `~/.local/bin` 目录。若该目录不在 `PATH` 变量中，需按照脚本提示添加。

   **手动安装：**

   通过 [Releases](https://github.com/yhlooo/gosh/releases) 页面下载可执行二进制，解压并将其中 `gosh` 文件放置到任意 `$PATH` 目录下。

2. **使用**

   运行 `gosh` 启动， Gosh 首次运行时会引导配置 LLM API 和密钥，按提示进行操作。

   随后 Gosh 将显示 Shell 界面，与你平时使用的 Shell 几乎一样，就像往常一样使用。

   在使用过程中输入 `shift+tab` 可在 Agent 模式和 Shell 模式切换（两种模式下输入提示会发生变化），在 Agent 模式下你可以和 LLM 驱动的 Agent 进行对话，它能解决你在 Shell 中遇到的各种问题。
