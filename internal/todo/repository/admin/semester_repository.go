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

type SemesterRepository struct {
    pool *pgxpool.Pool
}

func NewSemesterRepository(pool *pgxpool.Pool) *SemesterRepository {
    return &SemesterRepository{pool: pool}
}

// List returns semesters with optional filters and pagination; orderBy must be sanitized beforehand
func (r *SemesterRepository) List(ctx context.Context, q string, tahunAjaran, term *string, limit, offset int, orderBy string) ([]model.Semester, error) {
    sb := strings.Builder{}
    args := []any{}
    sb.WriteString("SELECT id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai, created_at, updated_at FROM semester")

    where := []string{}
    if q != "" {
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        args = append(args, "%"+q+"%")
        where = append(where, fmt.Sprintf("(id_semester ILIKE $%d OR tahun_ajaran ILIKE $%d OR term ILIKE $%d)", len(args)-2, len(args)-1, len(args)))
    }
    if tahunAjaran != nil && *tahunAjaran != "" {
        args = append(args, *tahunAjaran)
        where = append(where, fmt.Sprintf("tahun_ajaran = $%d", len(args)))
    }
    if term != nil && *term != "" {
        args = append(args, *term)
        where = append(where, fmt.Sprintf("term = $%d", len(args)))
    }

    if len(where) > 0 {
        sb.WriteString(" WHERE ")
        sb.WriteString(strings.Join(where, " AND "))
    }

    if orderBy == "" {
        orderBy = "id_semester DESC"
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

    var out []model.Semester
    for rows.Next() {
        var s model.Semester
        if err := rows.Scan(&s.IDSemester, &s.TahunAjaran, &s.Term, &s.TanggalMulai, &s.TanggalSelesai, &s.CreatedAt, &s.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, s)
    }
    return out, rows.Err()
}

func (r *SemesterRepository) GetByID(ctx context.Context, id string) (*model.Semester, error) {
    const q = `SELECT id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai, created_at, updated_at FROM semester WHERE id_semester = $1`
    row := r.pool.QueryRow(ctx, q, id)
    var s model.Semester
    if err := row.Scan(&s.IDSemester, &s.TahunAjaran, &s.Term, &s.TanggalMulai, &s.TanggalSelesai, &s.CreatedAt, &s.UpdatedAt); err != nil {
        return nil, err
    }
    return &s, nil
}

func (r *SemesterRepository) ExistsID(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM semester WHERE id_semester = $1 LIMIT 1`
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

func (r *SemesterRepository) Create(ctx context.Context, s *model.Semester) (*model.Semester, error) {
    const q = `INSERT INTO semester (id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai)
              VALUES ($1,$2,$3,$4,$5)
              RETURNING id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, s.IDSemester, s.TahunAjaran, s.Term, s.TanggalMulai, s.TanggalSelesai)
    var out model.Semester
    if err := row.Scan(&out.IDSemester, &out.TahunAjaran, &out.Term, &out.TanggalMulai, &out.TanggalSelesai, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *SemesterRepository) UpdatePut(ctx context.Context, id string, s *model.Semester) (*model.Semester, error) {
    const q = `UPDATE semester SET tahun_ajaran=$1, term=$2, tanggal_mulai=$3, tanggal_selesai=$4 WHERE id_semester=$5
              RETURNING id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai, created_at, updated_at`
    row := r.pool.QueryRow(ctx, q, s.TahunAjaran, s.Term, s.TanggalMulai, s.TanggalSelesai, id)
    var out model.Semester
    if err := row.Scan(&out.IDSemester, &out.TahunAjaran, &out.Term, &out.TanggalMulai, &out.TanggalSelesai, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *SemesterRepository) UpdatePatch(ctx context.Context, id string, tahunAjaran, term *string, tglMulai, tglSelesai *time.Time) (*model.Semester, error) {
    sets := []string{}
    args := []any{}
    idx := 1
    if tahunAjaran != nil {
        sets = append(sets, fmt.Sprintf("tahun_ajaran = $%d", idx))
        args = append(args, *tahunAjaran)
        idx++
    }
    if term != nil {
        sets = append(sets, fmt.Sprintf("term = $%d", idx))
        args = append(args, *term)
        idx++
    }
    if tglMulai != nil {
        sets = append(sets, fmt.Sprintf("tanggal_mulai = $%d", idx))
        args = append(args, *tglMulai)
        idx++
    }
    if tglSelesai != nil {
        sets = append(sets, fmt.Sprintf("tanggal_selesai = $%d", idx))
        args = append(args, *tglSelesai)
        idx++
    }

    if len(sets) == 0 {
        return r.GetByID(ctx, id)
    }

    args = append(args, id)
    q := fmt.Sprintf("UPDATE semester SET %s WHERE id_semester = $%d RETURNING id_semester, tahun_ajaran, term, tanggal_mulai, tanggal_selesai, created_at, updated_at", strings.Join(sets, ", "), idx)
    row := r.pool.QueryRow(ctx, q, args...)
    var out model.Semester
    if err := row.Scan(&out.IDSemester, &out.TahunAjaran, &out.Term, &out.TanggalMulai, &out.TanggalSelesai, &out.CreatedAt, &out.UpdatedAt); err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *SemesterRepository) HasKelasRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM kelas_kuliah WHERE id_semester = $1 LIMIT 1`
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

func (r *SemesterRepository) HasKRSRelated(ctx context.Context, id string) (bool, error) {
    const q = `SELECT 1 FROM krs WHERE id_semester = $1 LIMIT 1`
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

func (r *SemesterRepository) Delete(ctx context.Context, id string) error {
    const q = `DELETE FROM semester WHERE id_semester = $1`
    ct, err := r.pool.Exec(ctx, q, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return pgx.ErrNoRows
    }
    return nil
}