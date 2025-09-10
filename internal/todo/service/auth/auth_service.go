package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	model "pencatatan-data-mahasiswa/internal/todo/model/auth"
	repo "pencatatan-data-mahasiswa/internal/todo/repository/auth"

	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repo      *repo.Repository
	jwtSecret string
}

func NewService(r *repo.Repository, jwtSecret string) *Service {
	return &Service{repo: r, jwtSecret: jwtSecret}
}

var (
	errInvalidCredential = errors.New("invalid username or password")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUsernameTaken     = errors.New("username already taken")
)

var allowedRoles = map[string]struct{}{
	"admin":     {},
	"dosen":     {},
	"mahasiswa": {},
	"operator":  {},
}

// Register membuat user baru setelah validasi
func (s *Service) Register(ctx context.Context, username, password, role string, refID *string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(password) < 8 {
		return nil, ErrInvalidInput
	}
	if len(username) > 50 {
		return nil, ErrInvalidInput
	}
	if _, ok := allowedRoles[role]; !ok {
		return nil, ErrInvalidInput
	}
	// normalisasi refID dan validasi panjang
	if refID != nil {
		trimmed := strings.TrimSpace(*refID)
		refID = &trimmed
		if *refID == "" {
			refID = nil
		} else if len(*refID) > 20 {
			return nil, ErrInvalidInput
		}
	}
	// cek username unik
	exists, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}
	// validasi ref_id wajib dan harus valid bila role mahasiswa/dosen
	if role == "mahasiswa" || role == "dosen" {
		if refID == nil || *refID == "" {
			return nil, ErrInvalidInput
		}
		if role == "mahasiswa" {
			ok, err := s.repo.ExistsMahasiswaByID(ctx, *refID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, ErrInvalidInput
			}
		}
		if role == "dosen" {
			ok, err := s.repo.ExistsDosenByID(ctx, *refID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, ErrInvalidInput
			}
		}
	} else if role == "operator" && refID != nil && *refID != "" { // operator tidak boleh punya ref_id; admin boleh opsional
		return nil, ErrInvalidInput
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Username: username, PasswordHash: string(hash), Role: role, RefID: refID}
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return nil, ErrUsernameTaken
			}
		}
		return nil, err
	}
	return created, nil
}

// Login memvalidasi kredensial, dan mengembalikan token JWT HS256 dan masa berlaku
func (s *Service) Login(ctx context.Context, username, password string) (token string, expiresIn int64, user *model.User, err error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", 0, nil, errInvalidCredential
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", 0, nil, errInvalidCredential
	}

	// generate JWT 24 jam
	expiresIn = 24 * 60 * 60
	claims := jwt.MapClaims{
		"user_id":  u.IDUser,
		"username": u.Username,
		"role":     u.Role,
		"ref_id":   u.RefID,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Second * time.Duration(expiresIn)).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, nil, err
	}

	// nolkan hash sebelum dikembalikan
	u.PasswordHash = ""
	return signed, expiresIn, u, nil
}
