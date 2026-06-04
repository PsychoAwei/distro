package handlers

import (
	"errors"
	"net/http"

	"flight-booking/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateBookingRequest 创建预订的请求体
type CreateBookingRequest struct {
	FlightID       int64  `json:"flight_id" binding:"required"`
	PassengerName  string `json:"passenger_name" binding:"required"`
	PassengerPhone string `json:"passenger_phone" binding:"required"`
	SeatCount      int    `json:"seat_count" binding:"required"`
}

// CreateBooking POST /api/bookings
func CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	if req.SeatCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订票数量必须大于0"})
		return
	}

	// 查询航班获取票价
	flight, err := models.GetFlight(req.FlightID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "航班不存在"})
		return
	}

	// 生成预订号
	bookingNo := uuid.New().String()
	totalPrice := flight.Price * float64(req.SeatCount)

	booking, err := models.CreateBooking(bookingNo, req.FlightID, req.PassengerName, req.PassengerPhone, req.SeatCount, totalPrice)
	if err != nil {
		if errors.Is(err, models.ErrNoSeats) {
			c.JSON(http.StatusConflict, gin.H{"error": "座位不足，无法预订"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "预订成功",
		"booking":  booking,
		"flight":   flight,
	})
}

// GetBooking GET /api/bookings/:booking_no
func GetBooking(c *gin.Context) {
	bookingNo := c.Param("booking_no")

	booking, err := models.GetBooking(bookingNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": booking})
}

// CancelBooking DELETE /api/bookings/:booking_no
func CancelBooking(c *gin.Context) {
	bookingNo := c.Param("booking_no")

	booking, err := models.CancelBooking(bookingNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		if errors.Is(err, models.ErrAlreadyCancelled) {
			c.JSON(http.StatusConflict, gin.H{"error": "预订已取消，无需重复操作"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "取消成功",
		"booking": booking,
	})
}
