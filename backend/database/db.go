package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB

// Init 连接数据库，创建库和表
func Init(dsn string) error {
	// DSN 格式: user:pass@tcp(host:port)/dbname?params
	// 1. 先连接到 TiDB（不指定数据库），创建目标库
	baseDSN := dsn
	if idx := strings.LastIndex(dsn, "/"); idx != -1 {
		// 去掉 /dbname 部分，只保留 user:pass@tcp(host:port)/
		baseDSN = dsn[:idx+1]
	}

	initDB, err := sql.Open("mysql", baseDSN)
	if err != nil {
		return fmt.Errorf("打开初始连接失败: %w", err)
	}

	if err = initDB.Ping(); err != nil {
		initDB.Close()
		return fmt.Errorf("连接 TiDB 失败: %w", err)
	}

	if _, err = initDB.Exec("CREATE DATABASE IF NOT EXISTS flight_booking"); err != nil {
		initDB.Close()
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	initDB.Close()

	// 2. 用带数据库名的 DSN 重连
	dbDSN := dsn
	if idx := strings.LastIndex(dsn, "/"); idx != -1 {
		afterSlash := dsn[idx+1:]
		if strings.HasPrefix(afterSlash, "?") {
			// 原来没有数据库名: /?params → /flight_booking?params
			dbDSN = dsn[:idx+1] + "flight_booking" + afterSlash
		}
	}

	DB, err = sql.Open("mysql", dbDSN)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("✓ 数据库连接成功")

	// 3. 执行建表 DDL
	if err = execSchema(); err != nil {
		return fmt.Errorf("建表失败: %w", err)
	}

	// 4. 运行数据库迁移（增量添加缺失的列）
	if err = runMigrations(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// execSchema 执行内嵌的 schema.sql
func execSchema() error {
	statements := splitSQL(schemaSQL)
	for i, stmt := range statements {
		if stmt == "" {
			continue
		}
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("执行第 %d 条 SQL 失败: %w\nSQL: %s", i+1, err, stmt)
		}
	}

	log.Println("✓ 数据库表初始化完成")
	return nil
}

// runMigrations 增量迁移：为已有表添加缺失的列
func runMigrations() error {
	migrations := []struct {
		table  string
		column string
		sql    string
	}{
		{"users", "role", "ALTER TABLE users ADD COLUMN role ENUM('user','admin') NOT NULL DEFAULT 'user' COMMENT '用户角色'"},
		{"bookings", "user_id", "ALTER TABLE bookings ADD COLUMN user_id BIGINT NOT NULL DEFAULT 0 COMMENT '预订用户ID'"},
	}

	for _, m := range migrations {
		var count int
		err := DB.QueryRow(
			"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = 'flight_booking' AND TABLE_NAME = ? AND COLUMN_NAME = ?",
			m.table, m.column,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("检查列 %s.%s 是否存在失败: %w", m.table, m.column, err)
		}
		if count == 0 {
			log.Printf("→ 迁移: 添加 %s.%s 列", m.table, m.column)
			if _, err := DB.Exec(m.sql); err != nil {
				return fmt.Errorf("添加列 %s.%s 失败: %w", m.table, m.column, err)
			}
		}
	}

	log.Println("✓ 数据库迁移检查完成")
	return nil
}

// splitSQL 按分号分割 SQL 语句
func splitSQL(sql string) []string {
	var statements []string
	current := ""
	for _, ch := range sql {
		if ch == ';' {
			trimmed := strings.TrimSpace(current)
			if trimmed != "" {
				statements = append(statements, trimmed)
			}
			current = ""
		} else {
			current += string(ch)
		}
	}
	trimmed := strings.TrimSpace(current)
	if trimmed != "" {
		statements = append(statements, trimmed)
	}
	return statements
}
