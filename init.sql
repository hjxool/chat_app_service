CREATE DATABASE IF NOT EXISTS user_db DEFAULT CHARACTER SET utf8mb4;
-- 切换当前操作的数据库
USE user_db;

CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `username` VARCHAR(64) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL,
  -- 分布式系统用 TIMESTAMP 会有时区转换风险 如不同容器所在地区根据时区查看到的时间不同
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  -- ON UPDATE 关键字更新时自动更新字段值
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入一条测试用户数据，密码为明文 "123456" 的 bcrypt 哈希值
-- 尝试往 users 表里插入一条记录，username='admin'，password='123456'
-- INSERT INTO `users` (`username`, `password`) VALUES ('admin', '$2a$10$7R40kG0yJ7x6LgN3E6M24e4U6T3Yf0.1J/4X5O2j5mG3X5y6z7A8i') 
-- ON 是异常处理 DUPLICATE 是异常触发器名称
-- 如果插入时触发主键/唯一键冲突 就会执行 空更新 避免插入失败报错
-- ON DUPLICATE KEY UPDATE `id`=`id`;