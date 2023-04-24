CREATE TABLE `users` (
	`id`       INTEGER PRIMARY KEY AUTOINCREMENT,
	`name`     TEXT NOT NULL,
	`email`    TEXT NOT NULL,
	`company`  TEXT NOT NULL,
	`role`     INT  NOT NULL,
	`password` TEXT NOT NULL,
	UNIQUE (`email`)
);