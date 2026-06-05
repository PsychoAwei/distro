package models

import (
	"database/sql"
	"time"

	"flight-booking/database"
)

// Payment 支付模型
type Payment struct {
	ID        int64      `json:"id"`
	BookingID int64      `json:"booking_id"`
	UserID    int64      `json:"user_id"`
	Amount    float64    `json:"amount"`
	Status    string     `json:"status"`
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreatePayment 创建支付记录（状态为 pending）
func CreatePayment(bookingID, userID int64, amount float64) (*Payment, error) {
	result, err := database.DB.Exec(
		"INSERT INTO payments (booking_id, user_id, amount) VALUES (?, ?, ?)",
		bookingID, userID, amount,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Payment{
		ID:        id,
		BookingID: bookingID,
		UserID:    userID,
		Amount:    amount,
		Status:    "pending",
		CreatedAt: time.Now(),
	}, nil
}

// ExecutePayment 模拟支付（将 pending → paid）
func ExecutePayment(paymentID int64, userID int64) (*Payment, error) {
	// 1. 查询支付记录
	p := &Payment{}
	var paidAt sql.NullTime
	err := database.DB.QueryRow(
		"SELECT id, booking_id, user_id, amount, status, paid_at, created_at FROM payments WHERE id = ?",
		paymentID,
	).Scan(&p.ID, &p.BookingID, &p.UserID, &p.Amount, &p.Status, &paidAt, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if p.UserID != userID {
		return nil, ErrNotOwner
	}

	if p.Status == "paid" {
		return nil, ErrAlreadyPaid
	}

	if p.Status == "refunded" {
		return nil, ErrAlreadyCancelled
	}

	// 2. 更新支付状态
	now := time.Now()
	_, err = database.DB.Exec(
		"UPDATE payments SET status = 'paid', paid_at = ? WHERE id = ?",
		now, paymentID,
	)
	if err != nil {
		return nil, err
	}

	p.Status = "paid"
	p.PaidAt = &now
	return p, nil
}

// GetPaymentByBooking 按预订 ID 查支付
func GetPaymentByBooking(bookingID int64) (*Payment, error) {
	p := &Payment{}
	var paidAt sql.NullTime
	err := database.DB.QueryRow(
		"SELECT id, booking_id, user_id, amount, status, paid_at, created_at FROM payments WHERE booking_id = ?",
		bookingID,
	).Scan(&p.ID, &p.BookingID, &p.UserID, &p.Amount, &p.Status, &paidAt, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if paidAt.Valid {
		p.PaidAt = &paidAt.Time
	}
	return p, nil
}

// ListPaymentsByUser 用户自己的支付记录
func ListPaymentsByUser(userID int64) ([]Payment, error) {
	rows, err := database.DB.Query(
		"SELECT id, booking_id, user_id, amount, status, paid_at, created_at FROM payments WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		var paidAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.BookingID, &p.UserID, &p.Amount, &p.Status, &paidAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			p.PaidAt = &paidAt.Time
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []Payment{}
	}
	return payments, nil
}

// ListAllPayments 管理员查所有支付
func ListAllPayments() ([]Payment, error) {
	rows, err := database.DB.Query(
		"SELECT id, booking_id, user_id, amount, status, paid_at, created_at FROM payments ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		var paidAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.BookingID, &p.UserID, &p.Amount, &p.Status, &paidAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		if paidAt.Valid {
			p.PaidAt = &paidAt.Time
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []Payment{}
	}
	return payments, nil
}
