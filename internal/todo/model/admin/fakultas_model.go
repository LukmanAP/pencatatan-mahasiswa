package admin

import "time"

// Fakultas merepresentasikan baris pada tabel fakultas
// singkatan bersifat opsional
// created_at dan updated_at dikelola oleh database/trigger
type Fakultas struct {
    IDFakultas   string     `db:"id_fakultas" json:"id_fakultas"`
    NamaFakultas string     `db:"nama_fakultas" json:"nama_fakultas"`
    Singkatan    *string    `db:"singkatan" json:"singkatan"`
    CreatedAt    time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}