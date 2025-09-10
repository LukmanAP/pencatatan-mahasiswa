package admin

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "regexp"
    "strings"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
)

type ProdiService struct {
    repo *repo.ProdiRepository
}

func NewProdiService(r *repo.ProdiRepository) *ProdiService {
    return &ProdiService{repo: r}
}

var (
    // gunakan ErrInvalidInput & ErrConflict dari package yang sama (sudah didefinisikan di service fakultas)
    prodiIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{8}$`)
    kodePattern    = regexp.MustCompile(`^[A-Za-z0-9_-]{1,16}$`)
    jenjangSet     = map[string]struct{}{"D3":{}, "D4":{}, "S1":{}, "S2":{}, "S3":{}}
    akreditasiSet  = map[string]struct{}{"A":{}, "B":{}, "C":{}, "Baik":{}, "Baik Sekali":{}, "Unggul":{}}
)

// util: autogenerate ID untuk Prodi dengan prefix PRD + 5 digit
func (s *ProdiService) generateUniqueID(ctx context.Context) (string, error) {
    for i := 0; i < 10; i++ {
        nBig, err := rand.Int(rand.Reader, big.NewInt(100000))
        if err != nil {
            return "", err
        }
        id := fmt.Sprintf("PRD%05d", nBig.Int64())
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

func (s *ProdiService) validateCommon(p *model.Prodi) error {
    p.IDProdi = strings.TrimSpace(p.IDProdi)
    p.IDFakultas = strings.TrimSpace(p.IDFakultas)
    p.NamaProdi = strings.TrimSpace(p.NamaProdi)
    p.Jenjang = strings.TrimSpace(p.Jenjang)
    p.KodeProdi = strings.TrimSpace(p.KodeProdi)
    if p.Akreditasi != nil {
        trimmed := strings.TrimSpace(*p.Akreditasi)
        if trimmed == "" {
            p.Akreditasi = nil
        } else {
            p.Akreditasi = &trimmed
        }
    }

    if p.NamaProdi == "" || len(p.NamaProdi) < 3 || len(p.NamaProdi) > 120 {
        return ErrInvalidInput
    }
    if _, ok := jenjangSet[p.Jenjang]; !ok {
        return ErrInvalidInput
    }
    if !kodePattern.MatchString(p.KodeProdi) {
        return ErrInvalidInput
    }
    if p.Akreditasi != nil {
        if _, ok := akreditasiSet[*p.Akreditasi]; !ok {
            return ErrInvalidInput
        }
    }
    return nil
}

// List Prodi dengan filter dan pagination
func (s *ProdiService) List(ctx context.Context, q string, idFakultas, jenjang, akreditasi *string, limit, offset int, orderBy string) ([]model.Prodi, error) {
    if limit < 0 || offset < 0 {
        return nil, ErrInvalidInput
    }
    return s.repo.List(ctx, q, idFakultas, jenjang, akreditasi, limit, offset, orderBy)
}

// Get detail prodi by id_prodi
func (s *ProdiService) Get(ctx context.Context, id string) (*model.Prodi, error) {
    id = strings.TrimSpace(id)
    if !prodiIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    return s.repo.GetByID(ctx, id)
}

// Create Prodi: ID auto-generate jika kosong; validasi unik kode dan nama per fakultas+jenjang
func (s *ProdiService) Create(ctx context.Context, p *model.Prodi) (*model.Prodi, error) {
    if err := s.validateCommon(p); err != nil {
        return nil, err
    }
    // id_fakultas harus ada
    if p.IDFakultas == "" {
        return nil, ErrInvalidInput
    }
    if ok, err := s.repo.ExistsFakultas(ctx, p.IDFakultas); err != nil {
        return nil, err
    } else if !ok {
        return nil, ErrInvalidInput
    }

    // kode unik global
    if exist, err := s.repo.ExistsKode(ctx, p.KodeProdi, nil); err != nil {
        return nil, err
    } else if exist {
        return nil, ErrConflict
    }

    // nama unik per fakultas + jenjang (case-insensitive)
    if exist, err := s.repo.ExistsNamaPerFakultasJenjangCI(ctx, p.IDFakultas, p.Jenjang, p.NamaProdi, nil); err != nil {
        return nil, err
    } else if exist {
        return nil, ErrConflict
    }

    // handle ID
    if p.IDProdi == "" {
        id, err := s.generateUniqueID(ctx)
        if err != nil {
            return nil, err
        }
        p.IDProdi = id
    } else {
        if !prodiIDPattern.MatchString(p.IDProdi) {
            return nil, ErrInvalidInput
        }
        if exist, err := s.repo.ExistsID(ctx, p.IDProdi); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    return s.repo.Create(ctx, p)
}

// UpdatePut: full update kecuali id_prodi
func (s *ProdiService) UpdatePut(ctx context.Context, id string, p *model.Prodi) (*model.Prodi, error) {
    id = strings.TrimSpace(id)
    if !prodiIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    // Validasi field umum
    if err := s.validateCommon(p); err != nil {
        return nil, err
    }
    // id_fakultas harus valid
    if p.IDFakultas == "" {
        return nil, ErrInvalidInput
    }
    if ok, err := s.repo.ExistsFakultas(ctx, p.IDFakultas); err != nil {
        return nil, err
    } else if !ok {
        return nil, ErrInvalidInput
    }
    // Kode unik exclude id
    if exist, err := s.repo.ExistsKode(ctx, p.KodeProdi, &id); err != nil {
        return nil, err
    } else if exist {
        return nil, ErrConflict
    }
    // Nama unik per fakultas+jenjang exclude id
    if exist, err := s.repo.ExistsNamaPerFakultasJenjangCI(ctx, p.IDFakultas, p.Jenjang, p.NamaProdi, &id); err != nil {
        return nil, err
    } else if exist {
        return nil, ErrConflict
    }

    return s.repo.UpdatePut(ctx, id, p)
}

// UpdatePatch: partial update
func (s *ProdiService) UpdatePatch(ctx context.Context, id string, idFakultas, nama, jenjang, kode, akreditasi *string) (*model.Prodi, error) {
    id = strings.TrimSpace(id)
    if !prodiIDPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }

    // Validasi field jika disediakan
    if idFakultas != nil {
        *idFakultas = strings.TrimSpace(*idFakultas)
        if *idFakultas == "" {
            return nil, ErrInvalidInput
        }
        if ok, err := s.repo.ExistsFakultas(ctx, *idFakultas); err != nil {
            return nil, err
        } else if !ok {
            return nil, ErrInvalidInput
        }
    }
    if nama != nil {
        *nama = strings.TrimSpace(*nama)
        if len(*nama) < 3 || len(*nama) > 120 {
            return nil, ErrInvalidInput
        }
    }
    if jenjang != nil {
        *jenjang = strings.TrimSpace(*jenjang)
        if _, ok := jenjangSet[*jenjang]; !ok {
            return nil, ErrInvalidInput
        }
    }
    if kode != nil {
        *kode = strings.TrimSpace(*kode)
        if !kodePattern.MatchString(*kode) {
            return nil, ErrInvalidInput
        }
        // Kode unik exclude id
        if exist, err := s.repo.ExistsKode(ctx, *kode, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    if akreditasi != nil {
        *akreditasi = strings.TrimSpace(*akreditasi)
        if *akreditasi == "" {
            akreditasi = nil
        } else {
            if _, ok := akreditasiSet[*akreditasi]; !ok {
                return nil, ErrInvalidInput
            }
        }
    }

    // Jika ada kombinasi (id_fakultas || jenjang || nama) berubah, cek unik kombinasi tersebut
    if idFakultas != nil || jenjang != nil || nama != nil {
        // Ambil existing untuk mendapatkan nilai default saat pointer nil
        cur, err := s.repo.GetByID(ctx, id)
        if err != nil {
            return nil, err
        }
        finalIDF := cur.IDFakultas
        finalJen := cur.Jenjang
        finalNama := cur.NamaProdi
        if idFakultas != nil {
            finalIDF = *idFakultas
        }
        if jenjang != nil {
            finalJen = *jenjang
        }
        if nama != nil {
            finalNama = *nama
        }
        if exist, err := s.repo.ExistsNamaPerFakultasJenjangCI(ctx, finalIDF, finalJen, finalNama, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    return s.repo.UpdatePatch(ctx, id, idFakultas, nama, jenjang, kode, akreditasi)
}

func (s *ProdiService) Delete(ctx context.Context, id string) error {
    id = strings.TrimSpace(id)
    if !prodiIDPattern.MatchString(id) {
        return ErrInvalidInput
    }
    if has, err := s.repo.HasMahasiswaRelated(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    if has, err := s.repo.HasMataKuliahRelated(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    return s.repo.Delete(ctx, id)
}