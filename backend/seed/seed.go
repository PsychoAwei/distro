package seed

import (
	"log"

	"flight-booking/database"
)

// Run 插入示例航班数据
func Run() error {
	// 检查是否已存在数据
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM flights").Scan(&count)
	if count > 0 {
		log.Printf("✓ 已有 %d 条航班数据，跳过种子数据插入\n", count)
		return nil
	}

	flights := []struct {
		FlightNo    string
		Origin      string
		Destination string
		Departure   string
		Arrival     string
		Seats       int
		Price       float64
	}{
		{"CA1234", "北京", "上海", "2026-06-02 08:00:00", "2026-06-02 10:30:00", 180, 1280.00},
		{"CA1235", "北京", "上海", "2026-06-02 14:00:00", "2026-06-02 16:30:00", 180, 1380.00},
		{"MU5678", "上海", "广州", "2026-06-02 09:00:00", "2026-06-02 11:30:00", 200, 980.00},
		{"MU5679", "上海", "广州", "2026-06-02 18:00:00", "2026-06-02 20:30:00", 200, 1080.00},
		{"CZ9012", "广州", "北京", "2026-06-02 07:30:00", "2026-06-02 10:00:00", 160, 1560.00},
		{"CZ9013", "广州", "北京", "2026-06-02 13:00:00", "2026-06-02 15:30:00", 160, 1680.00},
		{"HU3456", "北京", "成都", "2026-06-03 10:00:00", "2026-06-03 13:00:00", 150, 1120.00},
		{"HU3457", "成都", "北京", "2026-06-03 15:00:00", "2026-06-03 18:00:00", 150, 1180.00},
	}

	for _, f := range flights {
		_, err := database.DB.Exec(
			"INSERT INTO flights (flight_no, origin, destination, departure_time, arrival_time, total_seats, available_seats, price) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			f.FlightNo, f.Origin, f.Destination, f.Departure, f.Arrival, f.Seats, f.Seats, f.Price,
		)
		if err != nil {
			return err
		}
	}

	log.Printf("✓ 已插入 %d 条示例航班数据\n", len(flights))
	return nil
}
