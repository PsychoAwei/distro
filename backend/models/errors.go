package models

import "errors"

// 业务错误定义
var (
	ErrNoSeats          = errors.New("座位不足")
	ErrNotFound         = errors.New("预订不存在")
	ErrAlreadyCancelled = errors.New("预订已取消")
	ErrNotOwner         = errors.New("无权操作此订单")
	ErrAlreadyPaid      = errors.New("订单已支付")
	ErrNotPaid          = errors.New("订单未支付")
	ErrInvalidRole      = errors.New("无效的用户角色")
	ErrAdminRequired     = errors.New("需要管理员权限")
)
