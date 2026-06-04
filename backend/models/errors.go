package models

import "errors"

// 业务错误定义
var (
	ErrNoSeats          = errors.New("座位不足")
	ErrNotFound         = errors.New("预订不存在")
	ErrAlreadyCancelled = errors.New("预订已取消")
)
