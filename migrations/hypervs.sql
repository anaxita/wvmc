CREATE TABLE IF NOT EXISTS `hypervs` (
  `name` varchar(255) PRIMARY KEY,
  `ip4` varchar(255) NOT NULL DEFAULT '0.0.0.0'
);