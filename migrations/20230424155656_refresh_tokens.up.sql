CREATE TABLE `refresh_tokens`
(
	`user_id` INT REFERENCES `users` (`id`) PRIMARY KEY,
	`token`   TEXT NOT NULL
);