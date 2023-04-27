CREATE TABLE `servers`
(
	`id`     TEXT PRIMARY KEY,
	`title`  TEXT NOT NULL,
	`hv`     TEXT NOT NULL,
	`state`  TEXT NOT NULL,
	`status` TEXT NOT NULL
);