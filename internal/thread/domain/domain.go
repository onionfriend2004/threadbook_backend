package domain

import "time"

type Thread struct {
	ID        int       `db:"id"`
	CreatorID int       `db:"creator_id"`
	SpoolID   int       `db:"spool_id"`
	Title     string    `db:"title"`
	Type      string    `db:"type"`
	IsClosed  bool      `db:"is_closed"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
