package admin

import "time"

// Mahasiswa merepresentasikan baris pada tabel mahasiswa
// Beberapa kolom bersifat opsional (pointer). Kolom angkatan dikelola trigger = tahun_masuk
// id_mahasiswa adalah NIM 12 karakter alfanumerik
// jenis_kelamin hanya {L,P}; status salah satu {Aktif,Cuti,Lulus,Drop Out,Non-Aktif}
type Mahasiswa struct {
	IDMahasiswa  string     `db:"id_mahasiswa" json:"id_mahasiswa"`
	IDProdi      string     `db:"id_prodi" json:"id_prodi"`
	NIK          *string    `db:"nik" json:"nik,omitempty"`
	NamaLengkap  string     `db:"nama_lengkap" json:"nama_lengkap"`
	JenisKelamin string     `db:"jenis_kelamin" json:"jenis_kelamin"`
	TempatLahir  *string    `db:"tempat_lahir" json:"tempat_lahir,omitempty"`
	TanggalLahir *time.Time `db:"tanggal_lahir" json:"tanggal_lahir,omitempty"`
	Alamat       *string    `db:"alamat" json:"alamat,omitempty"`
	Email        *string    `db:"email" json:"email,omitempty"`
	NoHP         *string    `db:"no_hp" json:"no_hp,omitempty"`
	TahunMasuk   int        `db:"tahun_masuk" json:"tahun_masuk"`
	Status       string     `db:"status" json:"status"`
	Angkatan     int        `db:"angkatan" json:"angkatan"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
