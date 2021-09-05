CREATE TABLE IF NOT EXISTS `refresh_tokens` (
  `user_id` int PRIMARY KEY,
  `token` text NOT NULL
);