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

	userID := c.GetInt64("user_id")

	// 查询航班获取票价
	flight, err := models.GetFlight(req.FlightID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "航班不存在"})
		return
	}

	// 生成预订号
	bookingNo := uuid.New().String()
	totalPrice := flight.Price * float64(req.SeatCount)

	booking, err := models.CreateBooking(bookingNo, userID, req.FlightID, req.PassengerName, req.PassengerPhone, req.SeatCount, totalPrice)
	if err != nil {
		if errors.Is(err, models.ErrNoSeats) {
			c.JSON(http.StatusConflict, gin.H{"error": "座位不足，无法预订"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "预订成功",
		"booking": booking,
		"flight":  flight,
	})
}

// ListMyBookings GET /api/bookings — 当前用户的订单列表
func ListMyBookings(c *gin.Context) {
	userID := c.GetInt64("user_id")

	bookings, err := models.ListBookingsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

// GetBooking GET /api/bookings/:booking_no
func GetBooking(c *gin.Context) {
	bookingNo := c.Param("booking_no")
	userID := c.GetInt64("user_id")

	booking, err := models.GetBooking(bookingNo, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		if errors.Is(err, models.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此订单"})
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
	userID := c.GetInt64("user_id")

	booking, err := models.CancelBooking(bookingNo, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		if errors.Is(err, models.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权取消此订单"})
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

// PayBooking POST /api/bookings/:booking_no/pay — 模拟支付
func PayBooking(c *gin.Context) {
	bookingNo := c.Param("booking_no")
	userID := c.GetInt64("user_id")

	// 1. 查预订，校验所有权
	booking, err := models.GetBooking(bookingNo, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		if errors.Is(err, models.ErrNotOwner) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此订单"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if booking.Status == "cancelled" {
		c.JSON(http.StatusConflict, gin.H{"error": "订单已取消，无法支付"})
		return
	}

	// 2. 检查是否已有支付记录
	existing, _ := models.GetPaymentByBooking(booking.ID)
	if existing != nil {
		if existing.Status == "paid" {
			c.JSON(http.StatusConflict, gin.H{"error": "订单已支付"})
			return
		}
		// 已有 pending 记录，直接执行支付
		payment, err := models.ExecutePayment(existing.ID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "支付成功",
			"payment": payment,
		})
		return
	}

	// 3. 创建支付记录并执行支付
	payment, err := models.CreatePayment(booking.ID, userID, booking.TotalPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payment, err = models.ExecutePayment(payment.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "支付成功",
		"payment": payment,
	})
}

// ========== 管理员订单管理 ==========

// AdminListBookings GET /api/admin/bookings
func AdminListBookings(c *gin.Context) {
	bookings, err := models.ListAllBookings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

// AdminCancelBooking DELETE /api/admin/bookings/:booking_no
func AdminCancelBooking(c *gin.Context) {
	bookingNo := c.Param("booking_no")

	// 管理员强制取消，userID=0 跳过所有权校验
	booking, err := models.CancelBooking(bookingNo, 0)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		if errors.Is(err, models.ErrAlreadyCancelled) {
			c.JSON(http.StatusConflict, gin.H{"error": "预订已取消"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已强制取消",
		"booking": booking,
	})
}

// AdminListPayments GET /api/admin/payments
func AdminListPayments(c *gin.Context) {
	payments, err := models.ListAllPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// AdminGetBookingDetail GET /api/admin/bookings/:booking_no
func AdminGetBookingDetail(c *gin.Context) {
	bookingNo := c.Param("booking_no")

	// 管理员查看，userID=0 跳过所有权校验
	booking, err := models.GetBooking(bookingNo, 0)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "预订不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 同时查支付信息
	payment, _ := models.GetPaymentByBooking(booking.ID)

	c.JSON(http.StatusOK, gin.H{
		"booking": booking,
		"payment": payment,
	})
}
