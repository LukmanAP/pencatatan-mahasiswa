package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	model "pencatatan-data-mahasiswa/internal/todo/model/auth"
)

// Repository bertanggung jawab berinteraksi dengan database
// Hanya expose method yang dibutuhkan oleh service

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetByUsername mengambil user berdasarkan username, atau mengembalikan (nil, sql.ErrNoRows)
func (r *Repository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	const q = `
		SELECT id_user, username, password_hash, role, ref_id, created_at, updated_at
		FROM users
		WHERE username = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, q, username)

	var (
		u   model.User
		ref sql.NullString
	)
	if err := row.Scan(&u.IDUser, &u.Username, &u.PasswordHash, &u.Role, &ref, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	if ref.Valid {
		v := ref.String
		u.RefID = &v
	}
	return &u, nil
}

// UsernameExists mengembalikan true jika username sudah ada
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	const q = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
	var dummy int
	err := r.pool.QueryRow(ctx, q, username).Scan(&dummy)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ExistsMahasiswaByID validasi keberadaan id_mahasiswa
func (r *Repository) ExistsMahasiswaByID(ctx context.Context, id string) (bool, error) {
	const q = `SELECT 1 FROM mahasiswa WHERE id_mahasiswa = $1 LIMIT 1`
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

// ExistsDosenByID validasi keberadaan id_dosen
func (r *Repository) ExistsDosenByID(ctx context.Context, id string) (bool, error) {
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

// Create menyimpan user baru, mengembalikan user yang dibuat (termasuk id dan timestamp)
func (r *Repository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	const q = `
		INSERT INTO users (username, password_hash, role, ref_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id_user, username, password_hash, role, ref_id, created_at, updated_at
	`
	// Jangan kirim *string langsung; gunakan nilai atau NULL
	var refParam interface{}
	if u.RefID != nil && *u.RefID != "" {
		refParam = *u.RefID
	} else {
		refParam = nil
	}

	row := r.pool.QueryRow(ctx, q, u.Username, u.PasswordHash, u.Role, refParam)
	var out model.User
	var ref sql.NullString
	if err := row.Scan(&out.IDUser, &out.Username, &out.PasswordHash, &out.Role, &ref, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	if ref.Valid {
		v := ref.String
		out.RefID = &v
	}
	return &out, nil
}
