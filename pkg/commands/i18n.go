package commands

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	MsgCmdShortDesc = &i18n.Message{
		ID:    "commands.CmdShortDesc",
		Other: "Gosh is a powerful shell with LLM enhancement.",
	}
	MsgCmdLongDesc = &i18n.Message{
		ID: "commands.CmdLongDesc",
		Other: `Gosh is a powerful shell with LLM enhancement.

Actually it is not a shell, but an enhanced wrapper for shells like zsh / bash. You can use your familiar shell as usual, but it is better than before.`,
	}

	MsgGlobalOptsVerbosityDesc = &i18n.Message{
		ID:    "commands.GlobalOptsVerbosityDesc",
		Other: "Number for the log level verbosity (0, 1, or 2)",
	}
	MsgGlobalOptsDebugDesc = &i18n.Message{ID: "commands.GlobalOptsDebugDesc", Other: "Run in debug mode"}

	MsgOptsShellDesc = &i18n.Message{ID: "commands.OptsShellDesc", Other: "Shell"}

	MsgCmdShortDescVersion         = &i18n.Message{ID: "commands.CmdShortDescVersion", Other: "Print the version information"}
	MsgVersionOptsOutputFormatDesc = &i18n.Message{ID: "commands.VersionOptsOutputFormatDesc", Other: "Output format. One of (json, yaml)"}
)
