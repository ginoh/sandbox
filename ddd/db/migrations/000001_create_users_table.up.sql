CREATE TABLE IF NOT EXISTS `users` (
    `id` int(11) UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'User ID',
    `name` varchar(32) NOT NULL COMMENT 'User Name',
    `created_at` datetime(6) NOT NULL COMMENT 'Created Time',
    `updated_at` datetime(6) NOT NULL COMMENT 'Updated Time'
) COMMENT='User';