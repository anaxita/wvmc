CREATE TABLE `servers`
(
	`id`            INTEGER PRIMARY KEY AUTOINCREMENT,
	`vmid`          TEXT NOT NULL,
	`title`         TEXT NOT NULL,
	`ip4`           TEXT NOT NULL,
	`hv`            TEXT NOT NULL,
	`out_addr`      TEXT NOT NULL,
	`hostname`      TEXT NOT NULL,
	`description`   TEXT NOT NULL,
	`company`       TEXT NOT NULL,
	`user_name`     TEXT NOT NULL,
	`user_password` TEXT NOT NULL,
	UNIQUE (title, hv)
);