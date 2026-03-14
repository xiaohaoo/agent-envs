package config

import "errors"

// 配置相关错误定义
var (
	// ErrConfigNotFound 配置文件不存在
	ErrConfigNotFound = errors.New("配置文件不存在")

	// ErrInvalidFormat 配置格式无效
	ErrInvalidFormat = errors.New("配置格式无效")

	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("权限不足")

	// ErrActiveProfileNotFound 激活的 profile 不存在
	ErrActiveProfileNotFound = errors.New("激活的 profile 不存在")
)
