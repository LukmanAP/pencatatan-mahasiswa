package admin

import (
    "context"
    "errors"
    "regexp"
    "strings"
    "time"

    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
)

type MahasiswaService struct {
    repo *repo.MahasiswaRepository
}

func NewMahasiswaService(r *repo.MahasiswaRepository) *MahasiswaService {
    return &MahasiswaService{repo: r}
}

var (
    // Reuse ErrInvalidInput and ErrConflict from this package (declared in fakultas_service.go)
    // Define unprocessable error for FK not found cases
    ErrUnprocessable = errors.New("unprocessable")

    nimPattern      = regexp.MustCompile(`^[A-Za-z0-9]{12}$`)
    jkSet           = map[string]struct{}{"L": {}, "P": {}}
    statusSet       = map[string]struct{}{"Aktif": {}, "Cuti": {}, "Lulus": {}, "Drop Out": {}, "Non-Aktif": {}}
    hpMhsPattern    = regexp.MustCompile(`^[0-9+\- ]{1,20}$`)
    nikPattern      = regexp.MustCompile(`^[0-9]{16}$`)
)

// validateCommon trims and validates common fields (except IDs and tahun_masuk)
func (s *MahasiswaService) validateCommon(m *model.Mahasiswa, isCreate bool, isPut bool) error {
    m.IDProdi = strings.TrimSpace(m.IDProdi)
    m.NamaLengkap = strings.TrimSpace(m.NamaLengkap)
    m.JenisKelamin = strings.TrimSpace(m.JenisKelamin)

    if m.NamaLengkap == "" || len(m.NamaLengkap) < 3 || len(m.NamaLengkap) > 120 {
        return ErrInvalidInput
    }
    if _, ok := jkSet[m.JenisKelamin]; !ok {
        return ErrInvalidInput
    }

    if m.NIK != nil {
        v := strings.TrimSpace(*m.NIK)
        if v == "" {
            m.NIK = nil
        } else {
            if !nikPattern.MatchString(v) {
                return ErrInvalidInput
            }
            m.NIK = &v
        }
    }
    if m.TempatLahir != nil {
        v := strings.TrimSpace(*m.TempatLahir)
        if v == "" {
            m.TempatLahir = nil
        } else if len(v) > 80 {
            return ErrInvalidInput
        } else {
            m.TempatLahir = &v
        }
    }
    if m.Alamat != nil {
        v := strings.TrimSpace(*m.Alamat)
        if v == "" {
            m.Alamat = nil
        } else {
            m.Alamat = &v
        }
    }
    if m.Email != nil {
        v := strings.TrimSpace(*m.Email)
        if v == "" {
            m.Email = nil
        } else {
            if len(v) > 120 || !emailPattern.MatchString(v) { // emailPattern from dosen_service.go
                return ErrInvalidInput
            }
            m.Email = &v
        }
    }
    if m.NoHP != nil {
        v := strings.TrimSpace(*m.NoHP)
        if v == "" {
            m.NoHP = nil
        } else {
            if !hpMhsPattern.MatchString(v) {
                return ErrInvalidInput
            }
            m.NoHP = &v
        }
    }

    if isPut {
        // PUT requires status present (handler sets it when provided). Ensure not empty
        if m.Status == "" {
            return ErrInvalidInput
        }
        if _, ok := statusSet[m.Status]; !ok {
            return ErrInvalidInput
        }
    } else if isCreate {
        // Default status if empty on create
        if m.Status == "" {
            m.Status = "Aktif"
        }
        if _, ok := statusSet[m.Status]; !ok {
            return ErrInvalidInput
        }
    }

    return nil
}

// List with filters and pagination
func (s *MahasiswaService) List(ctx context.Context, q string, idProdi *string, angkatan *int, status *string, limit, offset int, orderBy string) ([]model.Mahasiswa, error) {
    if limit < 0 || offset < 0 {
        return nil, ErrInvalidInput
    }
    if idProdi != nil {
        v := strings.TrimSpace(*idProdi)
        if v == "" {
            idProdi = nil
        } else {
            if !prodiIDPattern.MatchString(v) { // from prodi_service.go
                return nil, ErrInvalidInput
            }
            idProdi = &v
        }
    }
    if status != nil {
        v := strings.TrimSpace(*status)
        if v == "" {
            status = nil
        } else {
            if _, ok := statusSet[v]; !ok {
                return nil, ErrInvalidInput
            }
            status = &v
        }
    }
    return s.repo.List(ctx, strings.TrimSpace(q), idProdi, angkatan, status, limit, offset, orderBy)
}

func (s *MahasiswaService) Get(ctx context.Context, id string) (*model.Mahasiswa, error) {
    id = strings.TrimSpace(id)
    if !nimPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }
    return s.repo.GetByID(ctx, id)
}

func (s *MahasiswaService) Create(ctx context.Context, m *model.Mahasiswa) (*model.Mahasiswa, error) {
    // Validate ID (NIM) provided and unique
    m.IDMahasiswa = strings.TrimSpace(m.IDMahasiswa)
    if !nimPattern.MatchString(m.IDMahasiswa) {
        return nil, ErrInvalidInput
    }
    if exist, err := s.repo.ExistsID(ctx, m.IDMahasiswa); err != nil {
        return nil, err
    } else if exist {
        return nil, ErrConflict
    }

    // Validate mandatory and optional fields
    if err := s.validateCommon(m, true, false); err != nil {
        return nil, err
    }

    // tahun_masuk validation
    currentYear := time.Now().Year()
    if m.TahunMasuk < 2000 || m.TahunMasuk > currentYear+1 {
        return nil, ErrInvalidInput
    }

    // id_prodi validation and FK existence
    if m.IDProdi == "" || !prodiIDPattern.MatchString(m.IDProdi) {
        return nil, ErrInvalidInput
    }
    if ok, err := s.repo.ExistsProdi(ctx, m.IDProdi); err != nil {
        return nil, err
    } else if !ok {
        return nil, ErrUnprocessable
    }

    // Uniqueness checks for optional unique fields
    if m.Email != nil {
        if exist, err := s.repo.ExistsEmail(ctx, *m.Email, nil); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    if m.NIK != nil {
        if exist, err := s.repo.ExistsNIK(ctx, *m.NIK, nil); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    return s.repo.Create(ctx, m)
}

func (s *MahasiswaService) UpdatePut(ctx context.Context, id string, m *model.Mahasiswa) (*model.Mahasiswa, error) {
    id = strings.TrimSpace(id)
    if !nimPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }

    // Validate fields (PUT requires full data including status)
    if err := s.validateCommon(m, false, true); err != nil {
        return nil, err
    }

    // tahun_masuk validation
    currentYear := time.Now().Year()
    if m.TahunMasuk < 2000 || m.TahunMasuk > currentYear+1 {
        return nil, ErrInvalidInput
    }

    // id_prodi validation and existence
    if m.IDProdi == "" || !prodiIDPattern.MatchString(m.IDProdi) {
        return nil, ErrInvalidInput
    }
    if ok, err := s.repo.ExistsProdi(ctx, m.IDProdi); err != nil {
        return nil, err
    } else if !ok {
        return nil, ErrUnprocessable
    }

    // Uniqueness checks with exclude id
    if m.Email != nil {
        if exist, err := s.repo.ExistsEmail(ctx, *m.Email, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }
    if m.NIK != nil {
        if exist, err := s.repo.ExistsNIK(ctx, *m.NIK, &id); err != nil {
            return nil, err
        } else if exist {
            return nil, ErrConflict
        }
    }

    return s.repo.UpdatePut(ctx, id, m)
}

func (s *MahasiswaService) UpdatePatch(
    ctx context.Context,
    id string,
    idProdi, nik, namaLengkap, jenisKelamin, tempatLahir, alamat, email, noHP, status *string,
    tanggalLahir *time.Time,
    tahunMasuk *int,
) (*model.Mahasiswa, error) {
    id = strings.TrimSpace(id)
    if !nimPattern.MatchString(id) {
        return nil, ErrInvalidInput
    }

    // Validate and normalize each provided field
    if idProdi != nil {
        v := strings.TrimSpace(*idProdi)
        if v == "" {
            return nil, ErrInvalidInput
        }
        if !prodiIDPattern.MatchString(v) {
            return nil, ErrInvalidInput
        }
        if ok, err := s.repo.ExistsProdi(ctx, v); err != nil {
            return nil, err
        } else if !ok {
            return nil, ErrUnprocessable
        }
        idProdi = &v
    }
    if nik != nil {
        v := strings.TrimSpace(*nik)
        if v == "" {
            nik = nil
        } else {
            if !nikPattern.MatchString(v) {
                return nil, ErrInvalidInput
            }
            if exist, err := s.repo.ExistsNIK(ctx, v, &id); err != nil {
                return nil, err
            } else if exist {
                return nil, ErrConflict
            }
            nik = &v
        }
    }
    if namaLengkap != nil {
        v := strings.TrimSpace(*namaLengkap)
        if len(v) < 3 || len(v) > 120 {
            return nil, ErrInvalidInput
        }
        namaLengkap = &v
    }
    if jenisKelamin != nil {
        v := strings.TrimSpace(*jenisKelamin)
        if _, ok := jkSet[v]; !ok {
            return nil, ErrInvalidInput
        }
        jenisKelamin = &v
    }
    if tempatLahir != nil {
        v := strings.TrimSpace(*tempatLahir)
        if v == "" {
            tempatLahir = nil
        } else if len(v) > 80 {
            return nil, ErrInvalidInput
        } else {
            tempatLahir = &v
        }
    }
    if alamat != nil {
        v := strings.TrimSpace(*alamat)
        if v == "" {
            alamat = nil
        } else {
            alamat = &v
        }
    }
    if email != nil {
        v := strings.TrimSpace(*email)
        if v == "" {
            email = nil
        } else {
            if len(v) > 120 || !emailPattern.MatchString(v) {
                return nil, ErrInvalidInput
            }
            if exist, err := s.repo.ExistsEmail(ctx, v, &id); err != nil {
                return nil, err
            } else if exist {
                return nil, ErrConflict
            }
            email = &v
        }
    }
    if noHP != nil {
        v := strings.TrimSpace(*noHP)
        if v == "" {
            noHP = nil
        } else {
            if !hpMhsPattern.MatchString(v) {
                return nil, ErrInvalidInput
            }
            noHP = &v
        }
    }
    if status != nil {
        v := strings.TrimSpace(*status)
        if v == "" {
            status = nil
        } else {
            if _, ok := statusSet[v]; !ok {
                return nil, ErrInvalidInput
            }
            status = &v
        }
    }
    if tahunMasuk != nil {
        v := *tahunMasuk
        currentYear := time.Now().Year()
        if v < 2000 || v > currentYear+1 {
            return nil, ErrInvalidInput
        }
        tahunMasuk = &v
    }

    return s.repo.UpdatePatch(ctx, id, idProdi, nik, namaLengkap, jenisKelamin, tempatLahir, alamat, email, noHP, status, tanggalLahir, tahunMasuk)
}

func (s *MahasiswaService) Delete(ctx context.Context, id string) error {
    id = strings.TrimSpace(id)
    if !nimPattern.MatchString(id) {
        return ErrInvalidInput
    }
    if has, err := s.repo.HasKRSRelated(ctx, id); err != nil {
        return err
    } else if has {
        return ErrConflict
    }
    return s.repo.Delete(ctx, id)
}