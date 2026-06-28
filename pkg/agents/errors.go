package agents

import "errors"

var (
	ErrUserCancelled      = errors.New("UserCancelled")
	ErrContextWindowLimit = errors.New("ContextWindowLimit")
)
