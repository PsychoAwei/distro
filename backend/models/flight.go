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

// ListFlights 查询航班列表，支持出发地、目的地、日期过滤
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
