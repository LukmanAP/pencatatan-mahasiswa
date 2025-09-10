package auth

import "time"

// User merepresentasikan baris pada tabel users
// Kolom password_hash tidak akan diekspose keluar handler
// RefID bersifat opsional (nullable)
type User struct {
	IDUser       int64      `db:"id_user" json:"id_user"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role" json:"role"`
	RefID        *string    `db:"ref_id" json:"ref_id"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}