package admin

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
)

type DosenRepository struct {
    pool *pgxpool.Pool
}

func NewDosenRepository(pool *pgxpool.Pool) *DosenRepository {
    return &DosenRepository{pool: pool}
}

// List dosen dengan optional q (search nama/nidn/email), pagination dan orderBy sudah disanitasi di service/handler
func (r *DosenRepository) List(ctx context.Context, q string, limit, offset int, orderBy string) ([]model.Dosen, error) {
    sb := strings.Builder{}
    args := []any{}
    sb.WriteString("SELECT id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik, created_at, updated_at FROM dosen")

    if q != "" {
        // cari di nama_dosen, nidn, email (case-insensitive)
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        sb.WriteString(fmt.Sprintf(" WHERE (nama_dosen ILIKE $%d OR nidn ILIKE $%d OR email ILIKE $%d)", 1, 2, 3))
    }
    if orderBy == "" {
        orderBy = "nama_dosen ASC"
    }
    sb.WriteString(" ORDER BY ")
    sb.WriteString(orderBy)
    sb.WriteString(" LIMIT ")
    sb.WriteString(fmt.Sprintf("%d", limit))
    sb.WriteString(" OFFSET ")
    sb.WriteString(fmt.Sprintf("%d", offset))

    rows, err := r.pool.Query(ctx, sb.String(), args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := []model.Dosen{}
    for rows.Next() {
        var d model.Dosen
        if err := rows.Scan(&d.IDDosen, &d.NIDN, &d.NamaDosen, &d.Email, &d.NoHP, &d.JabatanAkademik, &d.CreatedAt, &d.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, d)
    }
    return out, rows.Err()
}

func (r *DosenRepository) GetByID(ctx context.Context, id string) (*model.Dosen, error) {
    const q = `SELECT id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik, created_at, updated_at FROM dosen WHERE id_dosen = $1`
    row := r.pool.QueryRow(ctx, q, id)
    var d model.Dosen
    if err := row.Scan(&d.IDDosen, &d.NIDN, &d.NamaDosen, &d.Email, &d.NoHP, &d.JabatanAkademik, &d.CreatedAt, &d.UpdatedAt); err != nil {
        return nil, err
    }
    return &d, nil
}

func (r *DosenRepository) ExistsID(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM dosen WHERE id_dosen = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, id).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *DosenRepository) ExistsNIDN(ctx context.Context, nidn string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM dosen WHERE nidn = $1 AND id_dosen <> $2 LIMIT 1`
        var dummy int
        err := r.pool.QueryRow(ctx, q, nidn, *excludeID).Scan(&dummy)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM dosen WHERE nidn = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, nidn).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *DosenRepository) ExistsEmail(ctx context.Context, email string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM dosen WHERE email = $1 AND id_dosen <> $2 LIMIT 1`
        var dummy int
        err := r.pool.QueryRow(ctx, q, email, *excludeID).Scan(&dummy)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM dosen WHERE email = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, email).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *DosenRepository) Create(ctx context.Context, d *model.Dosen) (*model.Dosen, error) {
    const q = `INSERT INTO dosen (id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik)
               VALUES ($1,$2,$3,$4,$5,$6)
               RETURNING id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, d.IDDosen, d.NIDN, d.NamaDosen, d.Email, d.NoHP, d.JabatanAkademik)
    var out model.Dosen
    if err := row.Scan(&out.IDDosen, &out.NIDN, &out.NamaDosen, &out.Email, &out.NoHP, &out.JabatanAkademik, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *DosenRepository) UpdatePut(ctx context.Context, id string, d *model.Dosen) (*model.Dosen, error) {
    const q = `UPDATE dosen
               SET nidn=$1, nama_dosen=$2, email=$3, no_hp=$4, jabatan_akademik=$5
               WHERE id_dosen=$6
               RETURNING id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, d.NIDN, d.NamaDosen, d.Email, d.NoHP, d.JabatanAkademik, id)
    var out model.Dosen
    if err := row.Scan(&out.IDDosen, &out.NIDN, &out.NamaDosen, &out.Email, &out.NoHP, &out.JabatanAkademik, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *DosenRepository) UpdatePatch(ctx context.Context, id string, nidn, nama, email, nohp, jabatan *string) (*model.Dosen, error) {
    sets := []string{}
    args := []any{}
    idx := 1
    if nidn != nil {
        sets = append(sets, fmt.Sprintf("nidn = $%d", idx))
        args = append(args, *nidn)
        idx++
    }
    if nama != nil {
        sets = append(sets, fmt.Sprintf("nama_dosen = $%d", idx))
        args = append(args, *nama)
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
    if jabatan != nil {
        sets = append(sets, fmt.Sprintf("jabatan_akademik = $%d", idx))
        args = append(args, *jabatan)
        idx++
    }

    if len(sets) == 0 {
        // tidak ada perubahan, kembalikan current row
        return r.GetByID(ctx, id)
    }

    q := fmt.Sprintf("UPDATE dosen SET %s WHERE id_dosen = $%d RETURNING id_dosen, nidn, nama_dosen, email, no_hp, jabatan_akademik, created_at, updated_at",
        strings.Join(sets, ", "), idx)
    args = append(args, id)

    row := r.pool.QueryRow(ctx, q, args...)
    var out model.Dosen
    if err := row.Scan(&out.IDDosen, &out.NIDN, &out.NamaDosen, &out.Email, &out.NoHP, &out.JabatanAkademik, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *DosenRepository) HasMataKuliahPenanggungJawab(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM mata_kuliah WHERE id_dosen_pj = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, id).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *DosenRepository) HasKelasKuliahPengampu(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM kelas_kuliah WHERE id_dosen_pengampu = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, id).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *DosenRepository) Delete(ctx context.Context, id string) error {
    const q = `DELETE FROM dosen WHERE id_dosen = $1`
    ct, err := r.pool.Exec(ctx, q, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}