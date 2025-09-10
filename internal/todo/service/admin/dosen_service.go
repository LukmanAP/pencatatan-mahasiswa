package admin

import (
    "context"
    "crypto/rand"
    "math/big"
    "regexp"
    "strings"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
)

type DosenService struct {
    repo *repo.DosenRepository
}

func NewDosenService(r *repo.DosenRepository) *DosenService {
    return &DosenService{repo: r}
}

var (
    dosenIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{10}$`)
    hpPattern      = regexp.MustCompile(`^[0-9+]{1,20}$`)
    emailPattern   = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
    numericPattern = regexp.MustCompile(`^[0-9]+$`)
)

func (s *DosenService) validateCommon(d *model.Dosen) error {
    d.NamaDosen = strings.TrimSpace(d.NamaDosen)
    if d.NamaDosen == "" || len(d.NamaDosen) < 3 || len(d.NamaDosen) > 120 {
        return ErrInvalidInput
    }

    if d.NIDN != nil {
        v := strings.TrimSpace(*d.NIDN)
        if v == "" {
            d.NIDN = nil
        } else {
            if len(v) > 16 || !numericPattern.MatchString(v) {
                return ErrInvalidInput
            }
            d.NIDN = &v
        }
    }
    if d.Email != nil {
        v := strings.TrimSpace(*d.Email)
        if v == "" {
            d.Email = nil
        } else {
            if len(v) > 120 || !emailPattern.MatchString(v) {
                return ErrInvalidInput
            }
            d.Email = &v
        }
    }
    if d.NoHP != nil {
        v := strings.TrimSpace(*d.NoHP)
        if v == "" {
            d.NoHP = nil
        } else {
            if len(v) > 20 || !hpPattern.MatchString(v) {
                return ErrInvalidInput
            }
            d.NoHP = &v
        }
    }
    if d.JabatanAkademik != nil {
        v := strings.TrimSpace(*d.JabatanAkademik)
        if v == "" {
            d.JabatanAkademik = nil
        } else if len(v) > 60 {
            return ErrInvalidInput
        } else {
            d.JabatanAkademik = &v
        }
    }
    return nil
}

// generateUniqueID membuat ID acak 10 karakter alfanumerik dan memastikan unik
func (s *DosenService) generateUniqueID(ctx context.Context) (string, error) {
    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
    for i := 0; i < 10; i++ { // maksimal 10 percobaan untuk menghindari collision
        b := make([]byte, 10)
        for j := 0; j < 10; j++ {
            nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
            if err != nil {
                return "", err
            }
            b[j] = letters[nBig.Int64()]
        }
        id := string(b)
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

// List dosen dengan pencarian q pada nama/nidn/email
func (s *DosenService) List(ctx context.Context, q string, limit, offset int, orderBy string) ([]model.Dosen, error) {
    if limit < 0 || offset < 0 {
        return nil, ErrInvalidInput
    }
    q = strings.TrimSpace(q)
    return s.repo.List(ctx, q, limit, offset, orderBy)
}

func (s *DosenService) Get(ctx context.Context, id string) (*model.Dosen, error) {
    id = strings.TrimSpace(id)
    if !dosenIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    return s.repo.GetByID(ctx, id)
}

func (s *DosenService) Create(ctx context.Context, d *model.Dosen) (*model.Dosen, error) {
    d.IDDosen = strings.TrimSpace(d.IDDosen)
    // Jika ID kosong, generate otomatis; jika diisi, validasi pola dan unik
    if d.IDDosen == "" {
        id, err := s.generateUniqueID(ctx)
        if err != nil {
            return nil, err
        }
        d.IDDosen = id
    } else {
        if !dosenIDPattern.MatchString(d.IDDosen) {
            return nil, ErrInvalidInput
        }
        if exist, err := s.repo.ExistsID(ctx, d.IDDosen); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    if err := s.validateCommon(d); err != nil {
        return nil, err
    }
    // Unik NIDN bila ada
    if d.NIDN != nil {
        if exist, err := s.repo.ExistsNIDN(ctx, *d.NIDN, nil); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    // Unik email bila ada
    if d.Email != nil {
        if exist, err := s.repo.ExistsEmail(ctx, *d.Email, nil); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    return s.repo.Create(ctx, d)
}

func (s *DosenService) UpdatePut(ctx context.Context, id string, d *model.Dosen) (*model.Dosen, error) {
    id = strings.TrimSpace(id)
    if !dosenIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    if err := s.validateCommon(d); err != nil {
        return nil, err
    }
    if d.NIDN != nil {
        if exist, err := s.repo.ExistsNIDN(ctx, *d.NIDN, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    if d.Email != nil {
        if exist, err := s.repo.ExistsEmail(ctx, *d.Email, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    return s.repo.UpdatePut(ctx, id, d)
}

func (s *DosenService) UpdatePatch(ctx context.Context, id string, nidn, nama, email, nohp, jabatan *string) (*model.Dosen, error) {
    id = strings.TrimSpace(id)
    if !dosenIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }

    if nidn != nil {
        *nidn = strings.TrimSpace(*nidn)
        if *nidn == "" {
            nidn = nil
        } else {
            if len(*nidn) > 16 || !numericPattern.MatchString(*nidn) {
                return nil, ErrInvalidInput
            }
            if exist, err := s.repo.ExistsNIDN(ctx, *nidn, &id); err != nil {
                return nil, err
            } else if exist {
                return nil, ErrConflict
            }
        }
    }
    if nama != nil {
        *nama = strings.TrimSpace(*nama)
        if len(*nama) < 3 || len(*nama) > 120 {
            return nil, ErrInvalidInput
        }
    }
    if email != nil {
        *email = strings.TrimSpace(*email)
        if *email == "" {
            email = nil
        } else {
            if len(*email) > 120 || !emailPattern.MatchString(*email) {
                return nil, ErrInvalidInput
            }
            if exist, err := s.repo.ExistsEmail(ctx, *email, &id); err != nil {
                return nil, err
            } else if exist {
                return nil, ErrConflict
            }
        }
    }
    if nohp != nil {
        *nohp = strings.TrimSpace(*nohp)
        if *nohp == "" {
            nohp = nil
        } else {
            if len(*nohp) > 20 || !hpPattern.MatchString(*nohp) {
                return nil, ErrInvalidInput
            }
        }
    }
    if jabatan != nil {
        *jabatan = strings.TrimSpace(*jabatan)
        if *jabatan == "" {
            jabatan = nil
        } else if len(*jabatan) > 60 {
            return nil, ErrInvalidInput
        }
    }

    return s.repo.UpdatePatch(ctx, id, nidn, nama, email, nohp, jabatan)
}

func (s *DosenService) Delete(ctx context.Context, id string) error {
    id = strings.TrimSpace(id)
    if !dosenIDPattern.MatchString(id) {
        return ErrInvalidInput
    }
    if has, err := s.repo.HasMataKuliahPenanggungJawab(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    if has, err := s.repo.HasKelasKuliahPengampu(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    return s.repo.Delete(ctx, id)
}