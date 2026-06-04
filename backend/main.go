package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"flight-booking/config"
	"flight-booking/database"
	"flight-booking/handlers"
	"flight-booking/middleware"
	"flight-booking/models"
	"flight-booking/seed"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// nullLogger 丢弃 MySQL 驱动日志（重试期间的 [mysql] 报错不展示）
type nullLogger struct{}

func (nullLogger) Print(v ...interface{}) {}

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 设置 JWT 密钥
	models.SetJWTSecret(cfg.JWTSecret)

	// 3. 连接数据库（带重试，适应 TiDB 初始化延迟）
	//    静默 MySQL 驱动日志，避免重试期间刷屏 "unexpected EOF"
	mysql.SetLogger(nullLogger{})
	var err error
	for i := 0; i < 30; i++ {
		if err = database.Init(cfg.DSN); err == nil {
			break
		}
		log.Printf("等待数据库就绪 (%d/30)...", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 4. 插入示例数据
	if err := seed.Run(); err != nil {
		log.Fatalf("种子数据插入失败: %v", err)
	}

	// 5. 前端目录（支持 Docker 环境变量覆盖）
	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = "../frontend"
	}

	// 6. 设置 Gin 路由
	r := gin.Default()

	r.Static("/static", frontendDir)
	r.GET("/", func(c *gin.Context) {
		c.File(frontendDir + "/index.html")
	})
	r.GET("/login", func(c *gin.Context) {
		c.File(frontendDir + "/login.html")
	})
	r.GET("/register", func(c *gin.Context) {
		c.File(frontendDir + "/register.html")
	})

	api := r.Group("/api")
	{
		api.GET("/ping", handlers.Ping)

		// 认证
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)

		// 航班（公开）
		api.GET("/flights", handlers.ListFlights)
		api.GET("/flights/:id", handlers.GetFlight)

		// 预订（需认证）
		auth := api.Group("", middleware.AuthRequired())
		{
			auth.GET("/profile", handlers.GetProfile)
			auth.POST("/bookings", handlers.CreateBooking)
			auth.GET("/bookings/:booking_no", handlers.GetBooking)
			auth.DELETE("/bookings/:booking_no", handlers.CancelBooking)
		}
	}

	// 7. 启动服务
	addr := ":8080"
	fmt.Printf("\n✈️  飞机订票系统已启动，监听 %s\n", addr)
	fmt.Println("📋 API 端点:")
	fmt.Println("   GET    /api/ping")
	fmt.Println("   POST   /api/register")
	fmt.Println("   POST   /api/login")
	fmt.Println("   GET    /api/profile        [需认证]")
	fmt.Println("   GET    /api/flights?origin=&destination=&date=")
	fmt.Println("   GET    /api/flights/:id")
	fmt.Println("   POST   /api/bookings        [需认证]")
	fmt.Println("   GET    /api/bookings/:booking_no [需认证]")
	fmt.Println("   DELETE /api/bookings/:booking_no [需认证]")
	fmt.Println()
	fmt.Println("🌐 前端页面:")
	fmt.Println("   http://localhost:8080/         （航班主页）")
	fmt.Println("   http://localhost:8080/login    （登录）")
	fmt.Println("   http://localhost:8080/register （注册）")

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
