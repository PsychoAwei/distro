package handlers

import (
	"net/http"
	"strconv"

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
