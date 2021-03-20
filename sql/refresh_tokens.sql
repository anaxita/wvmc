-- CREATE TABLE IF NOT EXISTS `refresh_tokens` (
--   `user_id` int PRIMARY KEY,
--   `token` text NOT NULL,
--   UNIQUE KEY `user_id` (`user_id`)
-- );


CREATE TABLE IF NOT EXISTS `refresh_tokens` (
  `user_id` int PRIMARY KEY,
  `token` text NOT NULL
);