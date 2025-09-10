package admin

import (
    "context"
    "crypto/rand"
    "errors"
    "fmt"
    "math/big"
    "regexp"
    "strings"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
)

type Service struct {
    repo *repo.FakultasRepository
}

func NewService(r *repo.FakultasRepository) *Service {
    return &Service{repo: r}
}

var (
    ErrInvalidInput = errors.New("invalid input")
    ErrConflict     = errors.New("conflict")
    idPattern       = regexp.MustCompile(`^[A-Za-z0-9]{8}$`)
)

// List with optional search and pagination
func (s *Service) List(ctx context.Context, search string, limit, offset int) ([]model.Fakultas, error) {
    search = strings.TrimSpace(search)
    if limit < 0 || offset < 0 {
        return nil, ErrInvalidInput
    }
    return s.repo.List(ctx, search, limit, offset)
}

// Get detail by id
func (s *Service) Get(ctx context.Context, id string) (*model.Fakultas, error) {
    id = strings.TrimSpace(id)
    if !idPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    return s.repo.GetByID(ctx, id)
}

// generateUniqueID membuat ID 8 karakter pattern "FAK" + 5 digit angka (contoh: FAK00001)
func (s *Service) generateUniqueID(ctx context.Context) (string, error) {
    for i := 0; i < 10; i++ { // maksimal 10 percobaan untuk menghindari collision
        nBig, err := rand.Int(rand.Reader, big.NewInt(100000))
        if err != nil {
            return "", err
        }
        id := fmt.Sprintf("FAK%05d", nBig.Int64())
        exists, err := s.repo.ExistsID(ctx, id)
        if err != nil {
            return "", err
        }
        if !exists {
            return id, nil
        }
    }
    return "", ErrConflict
}

// Create new fakultas with validations
func (s *Service) Create(ctx context.Context, f *model.Fakultas) (*model.Fakultas, error) {
    f.IDFakultas = strings.TrimSpace(f.IDFakultas)
    f.NamaFakultas = strings.TrimSpace(f.NamaFakultas)
    if f.Singkatan != nil {
        trimmed := strings.TrimSpace(*f.Singkatan)
        if trimmed == "" {
            f.Singkatan = nil
        } else {
            f.Singkatan = &trimmed
        }
    }

    // Validasi nama & singkatan
    if len(f.NamaFakultas) < 3 || len(f.NamaFakultas) > 100 {
        return nil, ErrInvalidInput
    }
    if f.Singkatan != nil && len(*f.Singkatan) > 20 {
        return nil, ErrInvalidInput
    }

    // Cek unik nama (case-insensitive)
    if exists, err := s.repo.ExistsNamaCI(ctx, f.NamaFakultas); err != nil {
        return nil, err
    } else if exists {
        return nil, ErrConflict
    }

    // Jika ID kosong, generate otomatis. Jika diisi, validasi pola dan unik
    if f.IDFakultas == "" {
        id, err := s.generateUniqueID(ctx)
        if err != nil {
            return nil, err
        }
        f.IDFakultas = id
    } else {
        if !idPattern.MatchString(f.IDFakultas) {
            return nil, ErrInvalidInput
        }
        if exists, err := s.repo.ExistsID(ctx, f.IDFakultas); err != nil {
            return nil, err
        } else if exists {
            return nil, ErrConflict
        }
    }

    return s.repo.Create(ctx, f)
}

// Update existing fakultas by id. Fields are optional.
func (s *Service) Update(ctx context.Context, id string, nama *string, singkatan *string) (*model.Fakultas, error) {
    id = strings.TrimSpace(id)
    if !idPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }

    var namaV *string
    if nama != nil {
        trimmed := strings.TrimSpace(*nama)
        if len(trimmed) == 0 || len(trimmed) < 3 || len(trimmed) > 100 {
            return nil, ErrInvalidInput
        }
        namaV = &trimmed
        // cek unik nama baru, case-insensitive
        if exists, err := s.repo.ExistsNamaCI(ctx, trimmed); err != nil {
            return nil, err
        } else if exists {
            return nil, ErrConflict
        }
    }

    var singV *string
    if singkatan != nil {
        trimmed := strings.TrimSpace(*singkatan)
        if trimmed == "" {
            singV = nil
        } else {
            if len(trimmed) > 20 {
                return nil, ErrInvalidInput
            }
            singV = &trimmed
        }
    }

    return s.repo.Update(ctx, id, namaV, singV)
}

// Delete a fakultas. Reject if prodi exists.
func (s *Service) Delete(ctx context.Context, id string) error {
    id = strings.TrimSpace(id)
    if !idPattern.MatchString(id) {
        return ErrInvalidInput
    }
    // Tolak jika masih ada prodi terkait
    if has, err := s.repo.HasProdiRelated(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    return s.repo.Delete(ctx, id)
}