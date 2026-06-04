-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间'
);

-- 航班表
CREATE TABLE IF NOT EXISTS flights (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    flight_no VARCHAR(20) NOT NULL UNIQUE COMMENT '航班号，如 CA1234',
    origin VARCHAR(100) NOT NULL COMMENT '出发城市',
    destination VARCHAR(100) NOT NULL COMMENT '到达城市',
    departure_time DATETIME NOT NULL COMMENT '出发时间',
    arrival_time DATETIME NOT NULL COMMENT '到达时间',
    total_seats INT NOT NULL DEFAULT 180 COMMENT '总座位数',
    available_seats INT NOT NULL DEFAULT 180 COMMENT '剩余座位数',
    price DECIMAL(10,2) NOT NULL COMMENT '票价',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

-- 预订表
CREATE TABLE IF NOT EXISTS bookings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    booking_no VARCHAR(36) NOT NULL UNIQUE COMMENT '预订号（UUID）',
    flight_id BIGINT NOT NULL COMMENT '航班ID',
    passenger_name VARCHAR(100) NOT NULL COMMENT '乘客姓名',
    passenger_phone VARCHAR(20) NOT NULL COMMENT '乘客电话',
    seat_count INT NOT NULL DEFAULT 1 COMMENT '订票数量',
    total_price DECIMAL(10,2) NOT NULL COMMENT '总价',
    status ENUM('confirmed','cancelled') DEFAULT 'confirmed' COMMENT '预订状态',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (flight_id) REFERENCES flights(id)
);
