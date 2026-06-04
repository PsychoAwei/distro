package models

import (
	"database/sql"
	"time"

	"flight-booking/database"
)

// Booking 预订模型
type Booking struct {
	ID             int64     `json:"id"`
	BookingNo      string    `json:"booking_no"`
	FlightID       int64     `json:"flight_id"`
	PassengerName  string    `json:"passenger_name"`
	PassengerPhone string    `json:"passenger_phone"`
	SeatCount      int       `json:"seat_count"`
	TotalPrice     float64   `json:"total_price"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateBooking 创建预订（在事务中扣减座位）
func CreateBooking(bookingNo string, flightID int64, passengerName, passengerPhone string, seatCount int, totalPrice float64) (*Booking, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 查询当前剩余座位（加行锁）
	var availableSeats int
	err = tx.QueryRow("SELECT available_seats FROM flights WHERE id = ? FOR UPDATE", flightID).Scan(&availableSeats)
	if err != nil {
		return nil, err
	}

	if availableSeats < seatCount {
		return nil, ErrNoSeats
	}

	// 2. 扣减座位
	_, err = tx.Exec("UPDATE flights SET available_seats = available_seats - ? WHERE id = ?", seatCount, flightID)
	if err != nil {
		return nil, err
	}

	// 3. 创建预订记录
	result, err := tx.Exec(
		"INSERT INTO bookings (booking_no, flight_id, passenger_name, passenger_phone, seat_count, total_price) VALUES (?, ?, ?, ?, ?, ?)",
		bookingNo, flightID, passengerName, passengerPhone, seatCount, totalPrice,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &Booking{
		ID:             id,
		BookingNo:      bookingNo,
		FlightID:       flightID,
		PassengerName:  passengerName,
		PassengerPhone: passengerPhone,
		SeatCount:      seatCount,
		TotalPrice:     totalPrice,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
	}, nil
}

// CancelBooking 取消预订（在事务中释放座位）
func CancelBooking(bookingNo string) (*Booking, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 查询预订（加行锁）
	b := &Booking{}
	err = tx.QueryRow(
		"SELECT id, booking_no, flight_id, passenger_name, passenger_phone, seat_count, total_price, status, created_at FROM bookings WHERE booking_no = ? FOR UPDATE",
		bookingNo,
	).Scan(&b.ID, &b.BookingNo, &b.FlightID, &b.PassengerName, &b.PassengerPhone, &b.SeatCount, &b.TotalPrice, &b.Status, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if b.Status == "cancelled" {
		return nil, ErrAlreadyCancelled
	}

	// 2. 更新预订状态
	_, err = tx.Exec("UPDATE bookings SET status = 'cancelled' WHERE booking_no = ?", bookingNo)
	if err != nil {
		return nil, err
	}

	// 3. 释放座位
	_, err = tx.Exec("UPDATE flights SET available_seats = available_seats + ? WHERE id = ?", b.SeatCount, b.FlightID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	b.Status = "cancelled"
	return b, nil
}

// GetBooking 根据预订号查询预订
func GetBooking(bookingNo string) (*Booking, error) {
	b := &Booking{}
	err := database.DB.QueryRow(
		"SELECT id, booking_no, flight_id, passenger_name, passenger_phone, seat_count, total_price, status, created_at FROM bookings WHERE booking_no = ?",
		bookingNo,
	).Scan(&b.ID, &b.BookingNo, &b.FlightID, &b.PassengerName, &b.PassengerPhone, &b.SeatCount, &b.TotalPrice, &b.Status, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}
