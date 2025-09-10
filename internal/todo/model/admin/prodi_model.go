package admin

import "time"

// Prodi merepresentasikan baris pada tabel prodi
// created_at dan updated_at dikelola oleh database/trigger
// akreditasi bersifat opsional
// jenjang harus salah satu dari {D3, D4, S1, S2, S3}
type Prodi struct {
    IDProdi    string     `db:"id_prodi" json:"id_prodi"`
    IDFakultas string     `db:"id_fakultas" json:"id_fakultas"`
    NamaProdi  string     `db:"nama_prodi" json:"nama_prodi"`
    Jenjang    string     `db:"jenjang" json:"jenjang"`
    KodeProdi  string     `db:"kode_prodi" json:"kode_prodi"`
    Akreditasi *string    `db:"akreditasi" json:"akreditasi"`
    CreatedAt  time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}