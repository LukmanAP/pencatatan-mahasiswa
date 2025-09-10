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

type FakultasRepository struct {
    pool *pgxpool.Pool
}

func NewFakultasRepository(pool *pgxpool.Pool) *FakultasRepository {
    return &FakultasRepository{pool: pool}
}

// List mengembalikan daftar fakultas dengan filter pencarian nama (ILIKE) dan pagination
func (r *FakultasRepository) List(ctx context.Context, search string, limit, offset int) ([]model.Fakultas, error) {
    sb := strings.Builder{}
    args := []any{}
    sb.WriteString("SELECT id_fakultas, nama_fakultas, singkatan, created_at, updated_at FROM fakultas")
    if search != "" {
        args = append(args, "%"+search+"%")
        sb.WriteString(fmt.Sprintf(" WHERE nama_fakultas ILIKE $%d", len(args)))
    }
    // default ordering by nama_fakultas asc untuk konsistensi
    sb.WriteString(" ORDER BY nama_fakultas ASC")

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

    var out []model.Fakultas
    for rows.Next() {
        var f model.Fakultas
        if err := rows.Scan(&f.IDFakultas, &f.NamaFakultas, &f.Singkatan, &f.CreatedAt, &f.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, f)
    }
    return out, rows.Err()
}

// GetByID mengambil satu fakultas berdasarkan id
func (r *FakultasRepository) GetByID(ctx context.Context, id string) (*model.Fakultas, error) {
    const q = `SELECT id_fakultas, nama_fakultas, singkatan, created_at, updated_at FROM fakultas WHERE id_fakultas = $1`
    row := r.pool.QueryRow(ctx, q, id)
    var f model.Fakultas
    if err := row.Scan(&f.IDFakultas, &f.NamaFakultas, &f.Singkatan, &f.CreatedAt, &f.UpdatedAt); err != nil {
        return nil, err
    }
    return &f, nil
}

// ExistsID mengembalikan true jika id_fakultas sudah ada
func (r *FakultasRepository) ExistsID(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM fakultas WHERE id_fakultas = $1 LIMIT 1`
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

// ExistsNamaCI mengembalikan true jika nama_fakultas sudah ada (case-insensitive)
func (r *FakultasRepository) ExistsNamaCI(ctx context.Context, nama string) (bool, error) {
    const q = `SELECT 1 FROM fakultas WHERE LOWER(nama_fakultas) = LOWER($1) LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, nama).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

// Create menambahkan fakultas baru
func (r *FakultasRepository) Create(ctx context.Context, f *model.Fakultas) (*model.Fakultas, error) {
    const q = `INSERT INTO fakultas (id_fakultas, nama_fakultas, singkatan) VALUES ($1, $2, $3)
               RETURNING id_fakultas, nama_fakultas, singkatan, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, f.IDFakultas, f.NamaFakultas, f.Singkatan)
    var out model.Fakultas
    if err := row.Scan(&out.IDFakultas, &out.NamaFakultas, &out.Singkatan, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

// Update memperbarui nama_fakultas dan/atau singkatan
func (r *FakultasRepository) Update(ctx context.Context, id string, nama *string, singkatan *string) (*model.Fakultas, error) {
    // Bangun SET dinamis
    sets := []string{}
    args := []any{}
    idx := 1
    if nama != nil {
        sets = append(sets, fmt.Sprintf("nama_fakultas = $%d", idx))
        args = append(args, *nama)
        idx++
    }
    if singkatan != nil {
        sets = append(sets, fmt.Sprintf("singkatan = $%d", idx))
        args = append(args, *singkatan)
        idx++
    }
    if len(sets) == 0 {
        return r.GetByID(ctx, id) // tidak ada perubahan, kembalikan data lama
    }
    args = append(args, id)
    q := fmt.Sprintf("UPDATE fakultas SET %s WHERE id_fakultas = $%d RETURNING id_fakultas, nama_fakultas, singkatan, created_at, updated_at", strings.Join(sets, ", "), idx)

    row := r.pool.QueryRow(ctx, q, args...)
    var out model.Fakultas
    if err := row.Scan(&out.IDFakultas, &out.NamaFakultas, &out.Singkatan, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

// HasProdiRelated mengecek apakah masih ada prodi terkait fakultas
func (r *FakultasRepository) HasProdiRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM prodi WHERE id_fakultas = $1 LIMIT 1`
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

// Delete menghapus baris fakultas
func (r *FakultasRepository) Delete(ctx context.Context, id string) error {
    const q = `DELETE FROM fakultas WHERE id_fakultas = $1`
    ct, err := r.pool.Exec(ctx, q, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}