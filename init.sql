CREATE TABLE IF NOT EXIST `users` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `username` VARCHAR(50) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL,
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入一条测试数据 (明文 123456，实际生产请用 bcrypt 加密存储)
-- 尝试往 users 表里插入一条记录，username='admin'，password='123456'
INSERT INTO `users` (`username`, `password`) VALUES ('admin', '123456') 
-- 如果插入时触发了唯一键冲突表里已经有 'admin' 执行 UPDATE `username`=`username`无操作更新 避免插入失败报错
ON DUPLICATE KEY UPDATE `username`=`username`;