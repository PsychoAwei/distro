package seed

import (
	"fmt"
	"log"

	"flight-booking/database"

	"golang.org/x/crypto/bcrypt"
)

// Run 插入示例航班数据和默认管理员
func Run() error {
	// 1. 检查是否已有默认管理员
	var adminCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&adminCount); err != nil {
		return fmt.Errorf("检查管理员账号失败: %w", err)
	}
	if adminCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = database.DB.Exec(
			"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
			"admin", string(hash), "admin",
		)
		if err != nil {
			return err
		}
		log.Println("✓ 已创建默认管理员账号 (admin / admin123)")
	}

	// 2. 检查是否已存在航班数据
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM flights").Scan(&count); err != nil {
		return fmt.Errorf("检查航班数据失败: %w", err)
	}
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
		{"CA1234", "北京", "上海", "2026-06-10 08:00:00", "2026-06-10 10:30:00", 180, 1280.00},
		{"CA1235", "北京", "上海", "2026-06-10 14:00:00", "2026-06-10 16:30:00", 180, 1380.00},
		{"MU5678", "上海", "广州", "2026-06-10 09:00:00", "2026-06-10 11:30:00", 200, 980.00},
		{"MU5679", "上海", "广州", "2026-06-10 18:00:00", "2026-06-10 20:30:00", 200, 1080.00},
		{"CZ9012", "广州", "北京", "2026-06-10 07:30:00", "2026-06-10 10:00:00", 160, 1560.00},
		{"CZ9013", "广州", "北京", "2026-06-10 13:00:00", "2026-06-10 15:30:00", 160, 1680.00},
		{"HU3456", "北京", "成都", "2026-06-11 10:00:00", "2026-06-11 13:00:00", 150, 1120.00},
		{"HU3457", "成都", "北京", "2026-06-11 15:00:00", "2026-06-11 18:00:00", 150, 1180.00},
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
