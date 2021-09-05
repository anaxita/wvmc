CREATE TABLE IF NOT EXISTS `users` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT "Пользователь",
  `email` varchar(255) NOT NULL,
  `company` varchar(255) NOT NULL DEFAULT "Компания",
  `role` int NOT NULL DEFAULT 0,
  `password` text NOT NULL,
  UNIQUE (`email`)
);