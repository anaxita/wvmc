CREATE TABLE `users` (
	`id`       VARCHAR(36) PRIMARY KEY, -- UUID
	`name`     TEXT NOT NULL,
	`email`    TEXT NOT NULL,
	`company`  TEXT NOT NULL,
	`role`     INT  NOT NULL,
	`password` TEXT NOT NULL,
	UNIQUE (`email`)
);