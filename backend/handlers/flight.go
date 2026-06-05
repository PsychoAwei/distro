package handlers

import (
	"net/http"
	"strconv"
	"time"

	"flight-booking/database"
	"flight-booking/models"

	"github.com/gin-gonic/gin"
)

// Ping 健康检查
func Ping(c *gin.Context) {
	err := database.DB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "message": "TiDB connected"})
}

// ListFlights GET /api/flights?origin=&destination=&date=
func ListFlights(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")
	date := c.Query("date")

	flights, err := models.ListFlights(origin, destination, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if flights == nil {
		flights = []models.Flight{} // 返回空数组而不是 null
	}
	c.JSON(http.StatusOK, gin.H{"flights": flights})
}

// GetFlight GET /api/flights/:id
func GetFlight(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的航班ID"})
		return
	}

	flight, err := models.GetFlight(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "航班不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"flight": flight})
}

// ========== 管理员航班管理 ==========

// CreateFlightRequest 新增航班请求体
type CreateFlightRequest struct {
	FlightNo      string  `json:"flight_no" binding:"required"`
	Origin        string  `json:"origin" binding:"required"`
	Destination   string  `json:"destination" binding:"required"`
	DepartureTime string  `json:"departure_time" binding:"required"`
	ArrivalTime   string  `json:"arrival_time" binding:"required"`
	TotalSeats    int     `json:"total_seats" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
}

// AdminListFlights GET /api/admin/flights
func AdminListFlights(c *gin.Context) {
	flights, err := models.ListAllFlights()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"flights": flights})
}

// AdminCreateFlight POST /api/admin/flights
func AdminCreateFlight(c *gin.Context) {
	var req CreateFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效：" + err.Error()})
		return
	}

	depTime, err := time.Parse("2006-01-02 15:04:05", req.DepartureTime)
	if err != nil {
		depTime, err = time.Parse(time.RFC3339, req.DepartureTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "出发时间格式错误，请使用 YYYY-MM-DD HH:MM:SS"})
			return
		}
	}

	arrTime, err := time.Parse("2006-01-02 15:04:05", req.ArrivalTime)
	if err != nil {
		arrTime, err = time.Parse(time.RFC3339, req.ArrivalTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到达时间格式错误"})
			return
		}
	}

	flight, err := models.CreateFlight(req.FlightNo, req.Origin, req.Destination, depTime, arrTime, req.TotalSeats, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "航班创建成功",
		"flight":  flight,
	})
}

// AdminUpdateFlight PUT /api/admin/flights/:id
func AdminUpdateFlight(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的航班ID"})
		return
	}

	var req CreateFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效：" + err.Error()})
		return
	}

	depTime, err := time.Parse("2006-01-02 15:04:05", req.DepartureTime)
	if err != nil {
		depTime, err = time.Parse(time.RFC3339, req.DepartureTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "出发时间格式错误"})
			return
		}
	}

	arrTime, err := time.Parse("2006-01-02 15:04:05", req.ArrivalTime)
	if err != nil {
		arrTime, err = time.Parse(time.RFC3339, req.ArrivalTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到达时间格式错误"})
			return
		}
	}

	flight, err := models.UpdateFlight(id, req.FlightNo, req.Origin, req.Destination, depTime, arrTime, req.TotalSeats, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "航班更新成功",
		"flight":  flight,
	})
}

// AdminDeleteFlight DELETE /api/admin/flights/:id
func AdminDeleteFlight(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的航班ID"})
		return
	}

	if err := models.DeleteFlight(id); err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "航班不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "航班已删除"})
}
