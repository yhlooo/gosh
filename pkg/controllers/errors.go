package controllers

import "errors"

var (
	// ErrInvalidArguments 无效参数
	ErrInvalidArguments = errors.New("InvalidArguments")
	// ErrAlreadyStarted 已经启动了
	ErrAlreadyStarted = errors.New("AlreadyStarted")
)
