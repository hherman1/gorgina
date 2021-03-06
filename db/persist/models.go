// Code generated by sqlc. DO NOT EDIT.

package persist

import (
	"database/sql"
	"time"
)

type Activity struct {
	ID   string
	CID  string
	Ts   time.Time
	Note sql.NullString
}

type Catalog struct {
	ID           string
	Category     sql.NullString
	Brand        sql.NullString
	Color        sql.NullString
	Pattern      sql.NullString
	Title        sql.NullString
	Description  sql.NullString
	Price        sql.NullFloat64
	LastActivity sql.NullTime
	LastNote     sql.NullString
	Hidden       bool
}
