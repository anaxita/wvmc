CREATE TABLE IF NOT EXISTS `servers` (
  `id` varchar(255) PRIMARY KEY,
  `title` varchar(255) NOT NULL,
  `ip4` varchar(255) NOT NULL,
  `hv` varchar(255) NOT NULL,
  `hostname` varchar(255) NOT NULL,
  `user_name` varchar(255) NOT NULL,
  `user_password` varchar(255) NOT NULL
);