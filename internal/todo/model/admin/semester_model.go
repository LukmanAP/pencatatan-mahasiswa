package admin

import "time"

// Semester merepresentasikan data semester akademik
// id_semester mengikuti format YYYY{1|2|3} => 1:Ganjil, 2:Genap, 3:Antara
// tahun_ajaran mengikuti format "YYYY/YYYY" dan konsisten dengan id_semester
// tanggal_mulai/tanggal_selesai opsional
// created_at/updated_at dikelola oleh database (trigger untuk updated_at)
type Semester struct {
    IDSemester     string     `json:"id_semester" db:"id_semester"`
    TahunAjaran    string     `json:"tahun_ajaran" db:"tahun_ajaran"`
    Term           string     `json:"term" db:"term"`
    TanggalMulai   *time.Time `json:"tanggal_mulai,omitempty" db:"tanggal_mulai"`
    TanggalSelesai *time.Time `json:"tanggal_selesai,omitempty" db:"tanggal_selesai"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}