CREATE TABLE `users_servers` (
    `user_id` int PRIMARY KEY,
    `server_id` varchar(255) NOT NULL,
    KEY `server_id` (`server_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8;