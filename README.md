**[简体中文](README_CN.md)** | [English](README.md)

---

![GitHub License](https://img.shields.io/github/license/yhlooo/gosh)
[![GitHub Release](https://img.shields.io/github/v/release/yhlooo/gosh)](https://github.com/yhlooo/gosh/releases/latest)
[![release](https://github.com/yhlooo/gosh/actions/workflows/release.yaml/badge.svg)](https://github.com/yhlooo/gosh/actions/workflows/release.yaml)

# Gosh - LLM-Enhanced Shell

> **🏗️ This project is still in an early stage.**

Gosh is not a true shell like Zsh/Bash, but an enhanced wrapper for them. You can use your familiar shell as usual, but better than before.

## Quick Start

1. **Install Gosh**

   **Install via script:**

   ```bash
   curl -L https://raw.githubusercontent.com/yhlooo/gosh/refs/heads/master/scripts/install.sh | bash
   ```

   The script installs `gosh` to the `~/.local/bin` directory. If this directory is not in your `PATH`, follow the script's prompts to add it.

   **Manual installation:**

   Download the executable binary from the [Releases](https://github.com/yhlooo/gosh/releases) page, extract it, and place the `gosh` file into any directory in your `$PATH`.

2. **Usage**

   Run `gosh` to start. On first run, Gosh will guide you through configuring the LLM API and key — follow the prompts.

   Gosh will then display a shell interface almost identical to the shell you normally use. Just use it as usual.

   Press `shift+tab` to toggle between Agent mode and Shell mode (the input prompt changes in each mode). In Agent mode, you can converse with an LLM-powered agent that can help you solve various problems you encounter in the shell.