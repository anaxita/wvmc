-- CREATE TABLE IF NOT EXISTS `users_servers` (
--     `user_id` int PRIMARY KEY,
--     `server_id` varchar(255) NOT NULL,
--     KEY `server_id` (`server_id`) USING BTREE
-- ) ENGINE = InnoDB DEFAULT CHARSET = utf8;


CREATE TABLE IF NOT EXISTS `users_servers` (
    `user_id` int PRIMARY KEY,
    `server_id` varchar(255),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
);