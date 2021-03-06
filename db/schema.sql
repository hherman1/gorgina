CREATE TABLE IF NOT EXISTS CATALOG
(
	id NCHAR(36) NOT NULL PRIMARY KEY,
	category NCHAR(64),
	brand NCHAR(64),
	color NCHAR(64),
	pattern NCHAR(64),
	title text,
	description text,
	price real,
	last_activity timestamp,
	last_note text,
	hidden boolean NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS ACTIVITY
(
	id NCHAR(36) NOT NULL PRIMARY KEY,
	c_id NCHAR(36) references catalog(id) NOT NULL,
	ts timestamp NOT NULL,
	note text
);