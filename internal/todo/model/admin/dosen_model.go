package admin

import "time"

// Dosen merepresentasikan baris pada tabel dosen
// created_at dan updated_at dikelola oleh database/trigger
// Field opsional: nidn, email, no_hp, jabatan_akademik
// Aturan unik: nidn unik (jika ada), email unik (jika ada)
type Dosen struct {
	IDDosen         string    `db:"id_dosen" json:"id_dosen"`
	NIDN            *string   `db:"nidn" json:"nidn"`
	NamaDosen       string    `db:"nama_dosen" json:"nama_dosen"`
	Email           *string   `db:"email" json:"email"`
	NoHP            *string   `db:"no_hp" json:"no_hp"`
	JabatanAkademik *string   `db:"jabatan_akademik" json:"jabatan_akademik"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
