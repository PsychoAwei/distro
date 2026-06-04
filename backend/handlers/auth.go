package handlers

import (
	"net/http"

	"flight-booking/models"

	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register POST /api/register
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效：" + err.Error()})
		return
	}

	user, err := models.Register(req.Username, req.Password)
	if err == models.ErrUserExists {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user":    user,
	})
}

// Login POST /api/login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效：" + err.Error()})
		return
	}

	token, err := models.Login(req.Username, req.Password)
	if err == models.ErrUserNotFound || err == models.ErrWrongPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "登录成功",
		"token":    token,
		"username": req.Username,
	})
}

// GetProfile GET /api/profile （需认证）
func GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
	})
}
