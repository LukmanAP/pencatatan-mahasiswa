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

type ProdiRepository struct {
    pool *pgxpool.Pool
}

func NewProdiRepository(pool *pgxpool.Pool) *ProdiRepository {
    return &ProdiRepository{pool: pool}
}

// List returns prodi with optional filters and pagination and orderBy (pre-sanitized)
func (r *ProdiRepository) List(ctx context.Context, q string, idFakultas, jenjang, akreditasi *string, limit, offset int, orderBy string) ([]model.Prodi, error) {
    sb := strings.Builder{}
    args := []any{}
    sb.WriteString("SELECT id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi, created_at, updated_at FROM prodi")

    where := []string{}
    if q != "" {
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        where = append(where, fmt.Sprintf("(nama_prodi ILIKE $%d OR kode_prodi ILIKE $%d)", len(args)-1, len(args)))
    }
    if idFakultas != nil && *idFakultas != "" {
        args = append(args, *idFakultas)
        where = append(where, fmt.Sprintf("id_fakultas = $%d", len(args)))
    }
    if jenjang != nil && *jenjang != "" {
        args = append(args, *jenjang)
        where = append(where, fmt.Sprintf("jenjang = $%d", len(args)))
    }
    if akreditasi != nil && *akreditasi != "" {
        args = append(args, *akreditasi)
        where = append(where, fmt.Sprintf("akreditasi = $%d", len(args)))
    }
    if len(where) > 0 {
        sb.WriteString(" WHERE ")
        sb.WriteString(strings.Join(where, " AND "))
    }

    // orderBy passed in as safe string
    if orderBy == "" {
        orderBy = "nama_prodi ASC"
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

    var out []model.Prodi
    for rows.Next() {
        var p model.Prodi
        if err := rows.Scan(&p.IDProdi, &p.IDFakultas, &p.NamaProdi, &p.Jenjang, &p.KodeProdi, &p.Akreditasi, &p.CreatedAt, &p.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, p)
    }
    return out, rows.Err()
}

func (r *ProdiRepository) GetByID(ctx context.Context, id string) (*model.Prodi, error) {
    const q = `SELECT id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi, created_at, updated_at FROM prodi WHERE id_prodi = $1`
    row := r.pool.QueryRow(ctx, q, id)
    var p model.Prodi
    if err := row.Scan(&p.IDProdi, &p.IDFakultas, &p.NamaProdi, &p.Jenjang, &p.KodeProdi, &p.Akreditasi, &p.CreatedAt, &p.UpdatedAt); err != nil {
        return nil, err
    }
    return &p, nil
}

func (r *ProdiRepository) ExistsID(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM prodi WHERE id_prodi = $1 LIMIT 1`
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

func (r *ProdiRepository) ExistsFakultas(ctx context.Context, idFak string) (bool, error) {
    const q = `SELECT 1 FROM fakultas WHERE id_fakultas = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, idFak).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *ProdiRepository) ExistsKode(ctx context.Context, kode string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM prodi WHERE kode_prodi = $1 AND id_prodi <> $2 LIMIT 1`
        var dummy int
        err := r.pool.QueryRow(ctx, q, kode, *excludeID).Scan(&dummy)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM prodi WHERE kode_prodi = $1 LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, kode).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *ProdiRepository) ExistsNamaPerFakultasJenjangCI(ctx context.Context, idFakultas, jenjang, nama string, excludeID *string) (bool, error) {
    if excludeID != nil {
        const q = `SELECT 1 FROM prodi WHERE id_fakultas = $1 AND jenjang = $2 AND LOWER(nama_prodi) = LOWER($3) AND id_prodi <> $4 LIMIT 1`
        var dummy int
        err := r.pool.QueryRow(ctx, q, idFakultas, jenjang, nama, *excludeID).Scan(&dummy)
        if errors.Is(err, pgx.ErrNoRows) {
            return false, nil
        }
        if err != nil {
            return false, err
        }
        return true, nil
    }
    const q = `SELECT 1 FROM prodi WHERE id_fakultas = $1 AND jenjang = $2 AND LOWER(nama_prodi) = LOWER($3) LIMIT 1`
    var dummy int
    err := r.pool.QueryRow(ctx, q, idFakultas, jenjang, nama).Scan(&dummy)
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return true, nil
}

func (r *ProdiRepository) Create(ctx context.Context, p *model.Prodi) (*model.Prodi, error) {
    const q = `INSERT INTO prodi (id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi) VALUES ($1,$2,$3,$4,$5,$6)
               RETURNING id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, p.IDProdi, p.IDFakultas, p.NamaProdi, p.Jenjang, p.KodeProdi, p.Akreditasi)
    var out model.Prodi
    if err := row.Scan(&out.IDProdi, &out.IDFakultas, &out.NamaProdi, &out.Jenjang, &out.KodeProdi, &out.Akreditasi, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

// UpdatePut updates all mutable fields (id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi)
func (r *ProdiRepository) UpdatePut(ctx context.Context, id string, p *model.Prodi) (*model.Prodi, error) {
    const q = `UPDATE prodi
               SET id_fakultas=$1, nama_prodi=$2, jenjang=$3, kode_prodi=$4, akreditasi=$5
               WHERE id_prodi=$6
               RETURNING id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, p.IDFakultas, p.NamaProdi, p.Jenjang, p.KodeProdi, p.Akreditasi, id)
    var out model.Prodi
    if err := row.Scan(&out.IDProdi, &out.IDFakultas, &out.NamaProdi, &out.Jenjang, &out.KodeProdi, &out.Akreditasi, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

// UpdatePatch updates only provided fields
func (r *ProdiRepository) UpdatePatch(ctx context.Context, id string, idFakultas, nama, jenjang, kode *string, akreditasi *string) (*model.Prodi, error) {
    sets := []string{}
    args := []any{}
    idx := 1
    if idFakultas != nil {
        sets = append(sets, fmt.Sprintf("id_fakultas = $%d", idx))
        args = append(args, *idFakultas)
        idx++
    }
    if nama != nil {
        sets = append(sets, fmt.Sprintf("nama_prodi = $%d", idx))
        args = append(args, *nama)
        idx++
    }
    if jenjang != nil {
        sets = append(sets, fmt.Sprintf("jenjang = $%d", idx))
        args = append(args, *jenjang)
        idx++
    }
    if kode != nil {
        sets = append(sets, fmt.Sprintf("kode_prodi = $%d", idx))
        args = append(args, *kode)
        idx++
    }
    if akreditasi != nil {
        sets = append(sets, fmt.Sprintf("akreditasi = $%d", idx))
        args = append(args, *akreditasi)
        idx++
    }
    if len(sets) == 0 {
        return r.GetByID(ctx, id)
    }
    args = append(args, id)
    q := fmt.Sprintf("UPDATE prodi SET %s WHERE id_prodi = $%d RETURNING id_prodi, id_fakultas, nama_prodi, jenjang, kode_prodi, akreditasi, created_at, updated_at", strings.Join(sets, ", "), idx)
    row := r.pool.QueryRow(ctx, q, args...)
    var out model.Prodi
    if err := row.Scan(&out.IDProdi, &out.IDFakultas, &out.NamaProdi, &out.Jenjang, &out.KodeProdi, &out.Akreditasi, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *ProdiRepository) HasMahasiswaRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM mahasiswa WHERE id_prodi = $1 LIMIT 1`
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

func (r *ProdiRepository) HasMataKuliahRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM mata_kuliah WHERE id_prodi = $1 LIMIT 1`
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

func (r *ProdiRepository) Delete(ctx context.Context, id string) error {
    const q = `DELETE FROM prodi WHERE id_prodi = $1`
    ct, err := r.pool.Exec(ctx, q, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}