# ANSI 序列

| 序列      | 名称                                         |
|---------|--------------------------------------------|
| `\x1bN` | SS2 – Single Shift Two                     |
| `\x1bO` | SS3 – Single Shift Three                   |
| `\x1bP` | DCS – 设备控制字符串（Device Control String）       |
| `\x1b[` | CSI - 控制序列导入器（Control Sequence Introducer） |
| `\x1b\` | ST – 字符串终止（String Terminator）              |
| `\x1b]` | OSC – 操作系统命令（Operating System Command）     |
| `\x1bX` | SOS – 字符串开始（Start of String）               |
| `\x1b^` | PM – 私有消息（Privacy Message）                 |
| `\x1b_` | APC – 应用程序命令（Application Program Command）  |
| `\x1bc` | RIS – 重置为初始状态（Reset to Initial State）      |

ST（`\x1b\`，即 `ESC \`）是 OSC、DCS、SOS、PM、APC 等字符串序列的终止符。在 7-bit 环境中可用 BEL（`\x07`）替代。

## CSI

| 序列            | 名称                                             | 描述                                                           |
|---------------|------------------------------------------------|--------------------------------------------------------------|
| `CSI<n>A`     | CUU – 光标上移（Cursor Up）                          | 光标向上移动 n 行                                                   |
| `CSI<n>B`     | CUD – 光标下移（Cursor Down）                        | 光标向下移动 n 行                                                   |
| `CSI<n>C`     | CUF – 光标前移（Cursor Forward）                     | 光标向右移动 n 列                                                   |
| `CSI<n>D`     | CUB – 光标后移（Cursor Back）                        | 光标向左移动 n 列                                                   |
| `CSI<n>E`     | CNL – 光标下一行（Cursor Next Line）                  | 光标移到下面第 n 行的行首                                               |
| `CSI<n>F`     | CPL – 光标上一行（Cursor Previous Line）              | 光标移到上面第 n 行的行首                                               |
| `CSI<n>G`     | CHA – 光标水平绝对位置（Cursor Horizontal Absolute）     | 光标移到当前行第 n 列                                                 |
| `CSI<n>;<m>H` | CUP – 光标位置（Cursor Position）                    | 光标移到第 n 行第 m 列（原点 1,1）                                       |
| `CSI<n>;<m>f` | HVP – 水平垂直位置（Horizontal Vertical Position）     | 同 CUP                                                        |
| `CSI<n>d`     | VPA – 垂直行绝对位置（Vertical Position Absolute）      | 光标移到第 n 行，列保持不变                                              |
| `CSI<n>e`     | VPR – 垂直行相对位置（Vertical Position Relative）      | 光标向下移动 n 行，列保持不变                                             |
| `CSI s`       | SCP / SCOSC – 保存光标位置（Save Cursor Position）     | 保存当前光标位置                                                     |
| `CSI u`       | RCP / SCORC – 恢复光标位置（Restore Cursor Position）  | 恢复保存的光标位置                                                    |
| `CSI<n>J`     | ED – 擦除显示（Erase in Display）                    | 0: 光标擦到屏尾; 1: 屏首擦到光标; 2: 全屏; 3: 清除滚动缓冲区                      |
| `CSI<n>K`     | EL – 擦除行（Erase in Line）                        | 0: 光标擦到行尾; 1: 行首擦到光标; 2: 整行                                  |
| `CSI<n>X`     | ECH – 擦除字符（Erase Character）                    | 将光标及之后 n 个字符替换为空格                                            |
| `CSI<n>S`     | SU – 上滚（Scroll Up）                             | 向上滚动 n 行                                                     |
| `CSI<n>T`     | SD – 下滚（Scroll Down）                           | 向下滚动 n 行                                                     |
| `CSI<n>;<m>r` | DECSTBM – 设置上下滚动边界（Set Top and Bottom Margins） | 设置滚动区域从第 n 行到第 m 行                                           |
| `CSI<n>@`     | ICH – 插入字符（Insert Character）                   | 在光标处插入 n 个空格                                                 |
| `CSI<n>P`     | DCH – 删除字符（Delete Character）                   | 删除光标处的 n 个字符                                                 |
| `CSI<n>L`     | IL – 插入行（Insert Line）                          | 在当前行之上插入 n 行                                                 |
| `CSI<n>M`     | DL – 删除行（Delete Line）                          | 删除当前行起的 n 行                                                  |
| `CSI<n>b`     | REP – 重复（Repeat）                               | 重复前一个打印字符 n 次                                                |
| `CSI<n>I`     | CHT – 光标水平向前制表（Cursor Horizontal Forward Tab）  | 光标向前移动 n 个制表位                                                |
| `CSI<n>Z`     | CBT – 光标向后制表（Cursor Backward Tab）              | 光标向后移动 n 个制表位                                                |
| `CSI<n>m`     | SGR – 选择图形渲染（Select Graphic Rendition）         | 设置文本属性（颜色、粗体、下划线等），多参用 `;` 分隔                                |
| `CSI<n>h`     | SM – 设置模式（Set Mode）                            | 设置 ANSI 模式                                                   |
| `CSI<n>l`     | RM – 重置模式（Reset Mode）                          | 重置 ANSI 模式                                                   |
| `CSI?<n>h`    | DECSET – DEC 私有模式设置                            | 设置 DEC 私有模式                                                  |
| `CSI?<n>l`    | DECRST – DEC 私有模式重置                            | 重置 DEC 私有模式                                                  |
| `CSI<n>c`     | DA – 主设备属性（Primary Device Attributes）          | 请求终端属性，终端应答 `CSI?…c`                                         |
| `CSI>n c`     | DA – 次设备属性（Secondary Device Attributes）        | 请求终端次要属性，终端应答 `CSI>…c`                                       |
| `CSI<5>n`     | DSR – 设备状态报告（Device Status Report）             | 请求设备状态，终端应答 `CSI0n`                                          |
| `CSI<6>n`     | DSR – 光标位置报告（Cursor Position Report）           | 请求光标位置，终端应答 `CSI<r>;<c>R`                                    |
| `CSI<n>;<m>R` | CPR – 光标位置报告（Cursor Position Report）           | 光标位置报告（终端主动发送）                                               |
| `CSI!p`       | DECSTR – 软终端重置（Soft Terminal Reset）            | 软重置终端到初始状态                                                   |
| `CSI"p`       | DECSCL – 设置兼容级别（Set Conformance Level）         | 设置终端兼容级别                                                     |
| `CSI<n>q`     | DECLL – 加载 LED（Load LEDs）                      | 设置键盘 LED：0–4                                                 |
| `CSI<n>SP q`  | DECSCUSR – 设置光标样式（Set Cursor Style）            | 0: 闪烁块; 1: 闪烁块; 2: 稳定块; 3: 闪烁下划线; 4: 稳定下划线; 5: 闪烁竖线; 6: 稳定竖线 |
| `CSI<n>"q`    | DECSCA – 选择字符保护属性（Select Character Protection） | 0/1/2                                                        |
| `CSI<n>SP G`  | TBC – 清除制表（Tab Clear）                          | 0: 当前位置; 3: 清除全部制表位                                          |
| `CSI<n>SP W`  | SET_TB – 设置制表位（Set Tab Stop）                   | 在指定位置设置制表位                                                   |
| `CSI H`       | HTS – 水平制表设置（Horizontal Tab Set）               | 在当前列设置制表位                                                    |
| `CSI<n>t`     | XTWINOPS – xterm 窗口操作                          | 窗口最小化、最大化、调整大小、报告尺寸等                                         |
| `CSI<n>SP ~`  | DECSCPP – 设置每页列数（Set Columns per Page）         | 设置终端宽度（列数）                                                   |
| `CSI<n>*\|`   | DECSNLS – 设置每页行数（Set Lines per Screen）         | 设置终端高度（行数）                                                   |
| `CSI#P`       | XTPUSHCOLORS – 压入颜色                            | 将当前调色板压入栈                                                    |
| `CSI#Q`       | XTPOPCOLORS – 弹出颜色                             | 从栈中弹出调色板                                                     |
| `CSI#R`       | XTREPORTCOLORS – 报告颜色                          | 报告当前调色板                                                      |
| `CSI>q`       | XTVERSION – 查询终端版本                             | 查询终端名称和版本                                                    |
| `CSI>4;<m>m`  | XTMODKEYS – 修改其他键                              | 设置其他键的修饰键协议                                                  |

## OSC

| 序列                                            | 名称                                     | 描述                             |
|-----------------------------------------------|----------------------------------------|--------------------------------|
| `OSC 133 ; A [; click_events=1] ST`           | 提示符开始 (Prompt Start)                   | 在Shell提示符开始打印前发送               |
| `OSC 133 ; B ST`                              | 命令开始 (Command Start)                   | 提示符结束，用户开始输入命令区域               |
| `OSC 133 ; C [; cmdline_url=<EncodedURL>] ST` | 命令已执行 (Command Executed)               | 用户按下回车，命令开始执行，可选携带命令行 URL      |
| `OSC 133 ; D [; <ExitCode>] ST`               | 命令结束 (Command Finished)                | 命令执行完毕，可选携带退出码                 |
| `OSC 1337 ; SetMark ST`                       | 设置标记 (Set Mark)                        | 在终端中设置一个标记                     |
| `OSC 1337 ; ClearScrollback ST`               | 清除滚动缓冲区 (Clear Scrollback)             | 清除终端滚动缓冲区                      |
| `OSC 1337 ; File=<name>;<opts>:<base64> ST`   | 文件传输 (File Transfer)                   | 内联文件传输，支持 base64 编码            |
| `OSC 1337 ; RequestAttention=<n> ST`          | 请求注意 (Request Attention)               | 弹跳 Dock 图标请求用户注意，n=0 关闭、n=1 开启 |
| `OSC 1337 ; CurrentDir=<url> ST`              | 当前目录 (Current Directory)               | 向终端报告当前工作目录                    |
| `OSC 1337 ; RemoteHost=<user>@<host> ST`      | 远程主机 (Remote Host)                     | 向终端报告当前连接的远程主机                 |
| `OSC 1337 ; ShellIntegrationVersion=<n> ST`   | Shell 集成版本 (Shell Integration Version) | 报告 Shell 集成协议版本号               |
| `OSC 1337 ; SetUserVar=<name>=<value> ST`     | 设置用户变量 (Set User Variable)             | 设置用户自定义变量                      |
| `OSC 1337 ; CopyToClipboard=<base64> ST`      | 复制到剪贴板 (Copy to Clipboard)             | 将内容复制到系统剪贴板                    |
| `OSC 1337 ; ReportCellSize ST`                | 报告单元格大小 (Report Cell Size)             | 查询终端单元格像素尺寸                    |
| `OSC 1337 ; HighlightCursorLine=<n> ST`       | 高亮光标行 (Highlight Cursor Line)          | 高亮光标所在行，n=0 关闭、n=1 开启          |
| `OSC 1337 ; StealFocus ST`                    | 获取焦点 (Steal Focus)                     | 使终端窗口获取焦点                      |