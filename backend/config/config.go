package config

import "os"

// DBConfig 数据库连接配置
type DBConfig struct {
	DSN       string // Data Source Name，MySQL 连接串
	JWTSecret string // JWT 签名密钥
}

// Load 从环境变量加载配置，提供默认值
func Load() *DBConfig {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// 默认连接本地 TiDB，无密码，数据库 flight_booking
		dsn = "root:@tcp(127.0.0.1:4000)/?charset=utf8mb4&parseTime=True&loc=Local"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "flight-booking-secret-key-2026"
	}

	return &DBConfig{DSN: dsn, JWTSecret: jwtSecret}
}
