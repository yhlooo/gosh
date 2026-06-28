package commands

import "github.com/nicksnyder/go-i18n/v2/i18n"

var (
	MsgGlobalOptsVerbosityDesc = &i18n.Message{
		ID:    "commands.GlobalOptsVerbosityDesc",
		Other: "Number for the log level verbosity (0, 1, or 2)",
	}
	MsgGlobalOptsDebugDesc = &i18n.Message{ID: "commands.GlobalOptsDebugDesc", Other: "Run in debug mode"}

	MsgCmdShortDesc = &i18n.Message{
		ID:    "commands.CmdShortDesc",
		Other: "Gosh is a powerful shell with LLM enhancement.",
	}
	MsgCmdLongDesc = &i18n.Message{
		ID: "commands.CmdLongDesc",
		Other: `Gosh is a powerful shell with LLM enhancement.

Actually it is not a shell, but an enhanced wrapper for shells like zsh / bash. You can use your familiar shell as usual, but it is better than before.`,
	}
	MsgOptsShellDesc = &i18n.Message{ID: "commands.OptsShellDesc", Other: "Shell"}
	MsgOptsModelDesc = &i18n.Message{
		ID:    "commands.OptsModelDesc",
		Other: "Primary model for the current session",
	}
	MsgOptsVisionModelDesc = &i18n.Message{
		ID:    "commands.OptsVisionModelDesc",
		Other: "Vision model for the current session",
	}
	MsgOptsReasoningLevelDesc = &i18n.Message{
		ID:    "commands.OptsReasoningLevelDesc",
		Other: "Reasoning level (0, 1 or 2)",
	}

	MsgCmdShortDescVersion         = &i18n.Message{ID: "commands.CmdShortDescVersion", Other: "Print the version information"}
	MsgVersionOptsOutputFormatDesc = &i18n.Message{ID: "commands.VersionOptsOutputFormatDesc", Other: "Output format. One of (json, yaml)"}

	MsgCmdShortDescDebug          = &i18n.Message{ID: "commands.CmdShortDescDebug", Other: "Debug tools (internal)"}
	MsgCmdShortDescDebugParseANSI = &i18n.Message{ID: "commands.CmdShortDescDebugParseANSI", Other: "Parse ANSI from STDIN"}
)
