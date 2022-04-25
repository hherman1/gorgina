CREATE TABLE IF NOT EXISTS CATALOG
(
	id NCHAR(64) NOT NULL PRIMARY KEY,
	img BLOB,
	title text,
	description text
);

CREATE TABLE IF NOT EXISTS ACTIVITY
(
	id NCHAR(34) NOT NULL PRIMARY KEY,
	c_id NCHAR(64) references catalog(id),
	ts timestamp
);