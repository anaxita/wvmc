CREATE TABLE `users_servers`
(
	`user_id`   INT REFERENCES `users` (`id`) PRIMARY KEY,
	`server_id` TEXT
);