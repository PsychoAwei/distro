package models

import (
	"time"

	"flight-booking/database"
)

// Flight 航班模型
type Flight struct {
	ID             int64     `json:"id"`
	FlightNo       string    `json:"flight_no"`
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	Price          float64   `json:"price"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListFlights 查询航班列表，支持出发地、目的地、日期过滤（仅显示有余票的航班）
func ListFlights(origin, destination, date string) ([]Flight, error) {
	query := "SELECT id, flight_no, origin, destination, departure_time, arrival_time, total_seats, available_seats, price, created_at FROM flights WHERE available_seats > 0"
	args := []interface{}{}

	if origin != "" {
		query += " AND origin = ?"
		args = append(args, origin)
	}
	if destination != "" {
		query += " AND destination = ?"
		args = append(args, destination)
	}
	if date != "" {
		query += " AND DATE(departure_time) = ?"
		args = append(args, date)
	}

	query += " ORDER BY departure_time ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var f Flight
		err := rows.Scan(&f.ID, &f.FlightNo, &f.Origin, &f.Destination,
			&f.DepartureTime, &f.ArrivalTime, &f.TotalSeats, &f.AvailableSeats, &f.Price, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		flights = append(flights, f)
	}
	if flights == nil {
		flights = []Flight{}
	}
	return flights, nil
}

// ListAllFlights 查询所有航班（管理员用，不过滤余票）
func ListAllFlights() ([]Flight, error) {
	rows, err := database.DB.Query(
		"SELECT id, flight_no, origin, destination, departure_time, arrival_time, total_seats, available_seats, price, created_at FROM flights ORDER BY departure_time ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var f Flight
		if err := rows.Scan(&f.ID, &f.FlightNo, &f.Origin, &f.Destination,
			&f.DepartureTime, &f.ArrivalTime, &f.TotalSeats, &f.AvailableSeats, &f.Price, &f.CreatedAt); err != nil {
			return nil, err
		}
		flights = append(flights, f)
	}
	if flights == nil {
		flights = []Flight{}
	}
	return flights, nil
}

// GetFlight 根据 ID 查询单个航班
func GetFlight(id int64) (*Flight, error) {
	f := &Flight{}
	err := database.DB.QueryRow(
		"SELECT id, flight_no, origin, destination, departure_time, arrival_time, total_seats, available_seats, price, created_at FROM flights WHERE id = ?",
		id,
	).Scan(&f.ID, &f.FlightNo, &f.Origin, &f.Destination,
		&f.DepartureTime, &f.ArrivalTime, &f.TotalSeats, &f.AvailableSeats, &f.Price, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// CreateFlight 新增航班（管理员用）
func CreateFlight(flightNo, origin, destination string, departureTime, arrivalTime time.Time, totalSeats int, price float64) (*Flight, error) {
	result, err := database.DB.Exec(
		"INSERT INTO flights (flight_no, origin, destination, departure_time, arrival_time, total_seats, available_seats, price) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		flightNo, origin, destination, departureTime, arrivalTime, totalSeats, totalSeats, price,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Flight{
		ID:             id,
		FlightNo:       flightNo,
		Origin:         origin,
		Destination:    destination,
		DepartureTime:  departureTime,
		ArrivalTime:    arrivalTime,
		TotalSeats:     totalSeats,
		AvailableSeats: totalSeats,
		Price:          price,
		CreatedAt:      time.Now(),
	}, nil
}

// UpdateFlight 修改航班（管理员用）
func UpdateFlight(id int64, flightNo, origin, destination string, departureTime, arrivalTime time.Time, totalSeats int, price float64) (*Flight, error) {
	// 先查当前可用座位和总座位差
	var currentTotal, currentAvailable int
	err := database.DB.QueryRow(
		"SELECT total_seats, available_seats FROM flights WHERE id = ?", id,
	).Scan(&currentTotal, &currentAvailable)
	if err != nil {
		return nil, err
	}

	// 调整 available_seats：新总座位 - (旧总座位 - 旧可用座位)
	bookedSeats := currentTotal - currentAvailable
	newAvailable := totalSeats - bookedSeats
	if newAvailable < 0 {
		newAvailable = 0
	}

	_, err = database.DB.Exec(
		"UPDATE flights SET flight_no=?, origin=?, destination=?, departure_time=?, arrival_time=?, total_seats=?, available_seats=?, price=? WHERE id=?",
		flightNo, origin, destination, departureTime, arrivalTime, totalSeats, newAvailable, price, id,
	)
	if err != nil {
		return nil, err
	}

	return GetFlight(id)
}

// DeleteFlight 删除航班（管理员用）
func DeleteFlight(id int64) error {
	result, err := database.DB.Exec("DELETE FROM flights WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
