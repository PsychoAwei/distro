package models

import (
	"database/sql"
	"errors"
	"time"

	"flight-booking/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists    = errors.New("用户名已存在")
	ErrWrongPassword = errors.New("密码错误")
	ErrUserNotFound  = errors.New("用户不存在")
)

var jwtSecret []byte

// SetJWTSecret 设置 JWT 签名密钥
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// User 用户模型
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Claims JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Register 注册新用户
func Register(username, password, role string) (*User, error) {
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "admin" {
		return nil, ErrInvalidRole
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if count > 0 {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, string(hash), role,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &User{ID: id, Username: username, Role: role, CreatedAt: time.Now()}, nil
}

// Login 验证用户并返回 JWT token 和角色
func Login(username, password string) (string, string, error) {
	var u User
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, role FROM users WHERE username = ?", username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	if err == sql.ErrNoRows {
		return "", "", ErrUserNotFound
	}
	if err != nil {
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrWrongPassword
	}

	claims := &Claims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	return tokenStr, u.Role, nil
}

// ValidateToken 验证 JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("无效的token")
	}
	return claims, nil
}

// GetUserByID 根据 ID 查询用户
func GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ListUsers 查询所有用户（管理员用）
func ListUsers() ([]User, error) {
	rows, err := database.DB.Query(
		"SELECT id, username, password_hash, role, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []User{}
	}
	return users, nil
}

// UpdateUserRole 修改用户角色
func UpdateUserRole(id int64, role string) error {
	if role != "user" && role != "admin" {
		return ErrInvalidRole
	}
	result, err := database.DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser 删除用户
func DeleteUser(id int64) error {
	result, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// GetUserStats 返回用户统计
func GetUserStats() (total int, err error) {
	err = database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	return
}
