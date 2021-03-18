-- CREATE TABLE IF NOT EXISTS `users` (
--   `id` int UNSIGNED AUTO_INCREMENT PRIMARY KEY,
--   `name` varchar(255) NOT NULL,
--   `email` varchar(255) NOT NULL,
--   `company` varchar(255) NOT NULL,
--   `role` int NOT NULL,
--   `password` text NOT NULL,
--   UNIQUE KEY `email` (`email`) USING BTREE
-- );


CREATE TABLE IF NOT EXISTS `users` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT "Пользователь",
  `email` varchar(255) NOT NULL,
  `company` varchar(255) NOT NULL DEFAULT "",
  `role` int NOT NULL DEFAULT 0,
  `password` text NOT NULL,
  UNIQUE (`email`)
);