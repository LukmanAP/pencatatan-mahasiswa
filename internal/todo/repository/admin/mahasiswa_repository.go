package admin

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
)

type MahasiswaRepository struct {
    pool *pgxpool.Pool
}

func NewMahasiswaRepository(pool *pgxpool.Pool) *MahasiswaRepository {
    return &MahasiswaRepository{pool: pool}
}

// List returns mahasiswa with optional filters and pagination; orderBy must be sanitized beforehand
func (r *MahasiswaRepository) List(ctx context.Context, q string, idProdi *string, angkatan *int, status *string, limit, offset int, orderBy string) ([]model.Mahasiswa, error) {
    sb := strings.Builder{}
    args := []any{}
    sb.WriteString("SELECT id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status, angkatan, created_at, updated_at FROM mahasiswa")

    where := []string{}
    if q != "" {
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        where = append(where, fmt.Sprintf("(nama_lengkap ILIKE $%d OR email ILIKE $%d OR id_mahasiswa ILIKE $%d)", len(args)-2, len(args)-1, len(args)))
    }
    if idProdi != nil && *idProdi != "" {
        args = append(args, *idProdi)
        where = append(where, fmt.Sprintf("id_prodi = $%d", len(args)))
    }
    if angkatan != nil && *angkatan > 0 {
        args = append(args, *angkatan)
        where = append(where, fmt.Sprintf("tahun_masuk = $%d", len(args)))
    }
    if status != nil && *status != "" {
        args = append(args, *status)
        where = append(where, fmt.Sprintf("status = $%d", len(args)))
    }

    if len(where) > 0 {
        sb.WriteString(" WHERE ")
        sb.WriteString(strings.Join(where, " AND "))
    }

    if orderBy == "" {
        orderBy = "nama_lengkap ASC"
    }
    sb.WriteString(" ORDER BY ")
    sb.WriteString(orderBy)

    if limit > 0 {
        args = append(args, limit)
        sb.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
    }
    if offset > 0 {
        args = append(args, offset)
        sb.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
    }

    rows, err := r.pool.Query(ctx, sb.String(), args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.Mahasiswa
    for rows.Next() {
        var m model.Mahasiswa
        if err := rows.Scan(
            &m.IDMahasiswa,
            &m.IDProdi,
            &m.NIK,
            &m.NamaLengkap,
            &m.JenisKelamin,
            &m.TempatLahir,
            &m.TanggalLahir,
            &m.Alamat,
            &m.Email,
            &m.NoHP,
            &m.TahunMasuk,
            &m.Status,
            &m.Angkatan,
            &m.CreatedAt,
            &m.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        out = append(out, m)
    }
    return out, rows.Err()
}

func (r *MahasiswaRepository) GetByID(ctx context.Context, id string) (*model.Mahasiswa, error) {
    const q = `SELECT id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status, angkatan, created_at, updated_at FROM mahasiswa WHERE id_mahasiswa = $1`
    row := r.pool.QueryRow(ctx, q, id)
    var m model.Mahasiswa
    if err := row.Scan(
        &m.IDMahasiswa,
        &m.IDProdi,
        &m.NIK,
        &m.NamaLengkap,
        &m.JenisKelamin,
        &m.TempatLahir,
        &m.TanggalLahir,
        &m.Alamat,
        &m.Email,
        &m.NoHP,
        &m.TahunMasuk,
        &m.Status,
        &m.Angkatan,
        &m.CreatedAt,
        &m.UpdatedAt,
    ); err != nil {
        return nil, err
    }
    return &m, nil
}

func (r *MahasiswaRepository) ExistsID(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM mahasiswa WHERE id_mahasiswa = $1 LIMIT 1`
    var x int
    err := r.pool.QueryRow(ctx, q, id).Scan(&x)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *MahasiswaRepository) ExistsEmail(ctx context.Context, email string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM mahasiswa WHERE email = $1 AND id_mahasiswa <> $2 LIMIT 1`
        var x int
        err := r.pool.QueryRow(ctx, q, email, *excludeID).Scan(&x)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM mahasiswa WHERE email = $1 LIMIT 1`
    var x int
    err := r.pool.QueryRow(ctx, q, email).Scan(&x)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *MahasiswaRepository) ExistsNIK(ctx context.Context, nik string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM mahasiswa WHERE nik = $1 AND id_mahasiswa <> $2 LIMIT 1`
        var x int
        err := r.pool.QueryRow(ctx, q, nik, *excludeID).Scan(&x)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM mahasiswa WHERE nik = $1 LIMIT 1`
    var x int
    err := r.pool.QueryRow(ctx, q, nik).Scan(&x)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *MahasiswaRepository) ExistsProdi(ctx context.Context, idProdi string) (bool, error) {
    const q = `SELECT 1 FROM prodi WHERE id_prodi = $1 LIMIT 1`
    var x int
    err := r.pool.QueryRow(ctx, q, idProdi).Scan(&x)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *MahasiswaRepository) Create(ctx context.Context, m *model.Mahasiswa) (*model.Mahasiswa, error) {
    const q = `INSERT INTO mahasiswa (id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status)
              VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
              RETURNING id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status, angkatan, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, m.IDMahasiswa, m.IDProdi, m.NIK, m.NamaLengkap, m.JenisKelamin, m.TempatLahir, m.TanggalLahir, m.Alamat, m.Email, m.NoHP, m.TahunMasuk, m.Status)
    var out model.Mahasiswa
    if err := row.Scan(&out.IDMahasiswa, &out.IDProdi, &out.NIK, &out.NamaLengkap, &out.JenisKelamin, &out.TempatLahir, &out.TanggalLahir, &out.Alamat, &out.Email, &out.NoHP, &out.TahunMasuk, &out.Status, &out.Angkatan, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *MahasiswaRepository) UpdatePut(ctx context.Context, id string, m *model.Mahasiswa) (*model.Mahasiswa, error) {
    const q = `UPDATE mahasiswa SET id_prodi=$1, nik=$2, nama_lengkap=$3, jenis_kelamin=$4, tempat_lahir=$5, tanggal_lahir=$6, alamat=$7, email=$8, no_hp=$9, tahun_masuk=$10, status=$11 WHERE id_mahasiswa=$12
              RETURNING id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status, angkatan, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, m.IDProdi, m.NIK, m.NamaLengkap, m.JenisKelamin, m.TempatLahir, m.TanggalLahir, m.Alamat, m.Email, m.NoHP, m.TahunMasuk, m.Status, id)
    var out model.Mahasiswa
    if err := row.Scan(&out.IDMahasiswa, &out.IDProdi, &out.NIK, &out.NamaLengkap, &out.JenisKelamin, &out.TempatLahir, &out.TanggalLahir, &out.Alamat, &out.Email, &out.NoHP, &out.TahunMasuk, &out.Status, &out.Angkatan, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *MahasiswaRepository) UpdatePatch(ctx context.Context, id string, idProdi, nik, nama, jk, tempat, alamat, email, nohp, status *string, tgl *time.Time, tahunMasuk *int) (*model.Mahasiswa, error) {
    sets := []string{}
    args := []any{}
    idx := 1
    if idProdi != nil {
        sets = append(sets, fmt.Sprintf("id_prodi = $%d", idx))
        args = append(args, *idProdi)
        idx++
    }
    if nik != nil {
        sets = append(sets, fmt.Sprintf("nik = $%d", idx))
        args = append(args, *nik)
        idx++
    }
    if nama != nil {
        sets = append(sets, fmt.Sprintf("nama_lengkap = $%d", idx))
        args = append(args, *nama)
        idx++
    }
    if jk != nil {
        sets = append(sets, fmt.Sprintf("jenis_kelamin = $%d", idx))
        args = append(args, *jk)
        idx++
    }
    if tempat != nil {
        sets = append(sets, fmt.Sprintf("tempat_lahir = $%d", idx))
        args = append(args, *tempat)
        idx++
    }
    if tgl != nil {
        sets = append(sets, fmt.Sprintf("tanggal_lahir = $%d", idx))
        args = append(args, *tgl)
        idx++
    }
    if alamat != nil {
        sets = append(sets, fmt.Sprintf("alamat = $%d", idx))
        args = append(args, *alamat)
        idx++
    }
    if email != nil {
        sets = append(sets, fmt.Sprintf("email = $%d", idx))
        args = append(args, *email)
        idx++
    }
    if nohp != nil {
        sets = append(sets, fmt.Sprintf("no_hp = $%d", idx))
        args = append(args, *nohp)
        idx++
    }
    if tahunMasuk != nil {
        sets = append(sets, fmt.Sprintf("tahun_masuk = $%d", idx))
        args = append(args, *tahunMasuk)
        idx++
    }
    if status != nil {
        sets = append(sets, fmt.Sprintf("status = $%d", idx))
        args = append(args, *status)
        idx++
    }

    if len(sets) == 0 {
        return r.GetByID(ctx, id)
    }

    args = append(args, id)
    q := fmt.Sprintf("UPDATE mahasiswa SET %s WHERE id_mahasiswa = $%d RETURNING id_mahasiswa, id_prodi, nik, nama_lengkap, jenis_kelamin, tempat_lahir, tanggal_lahir, alamat, email, no_hp, tahun_masuk, status, angkatan, created_at, updated_at", strings.Join(sets, ", "), idx)
    row := r.pool.QueryRow(ctx, q, args...)
    var out model.Mahasiswa
    if err := row.Scan(&out.IDMahasiswa, &out.IDProdi, &out.NIK, &out.NamaLengkap, &out.JenisKelamin, &out.TempatLahir, &out.TanggalLahir, &out.Alamat, &out.Email, &out.NoHP, &out.TahunMasuk, &out.Status, &out.Angkatan, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *MahasiswaRepository) HasKRSRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM krs WHERE id_mahasiswa = $1 LIMIT 1`
    var x int
    err := r.pool.QueryRow(ctx, q, id).Scan(&x)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *MahasiswaRepository) Delete(ctx context.Context, id string) error {
    const q = `DELETE FROM mahasiswa WHERE id_mahasiswa = $1`
    ct, err := r.pool.Exec(ctx, q, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}