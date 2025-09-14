package admin

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	model "pencatatan-data-mahasiswa/internal/todo/model/admin"
	repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
)

// Reuse ErrInvalidInput and ErrConflict from this package (declared in fakultas_service.go)

type SemesterService struct {
	repo *repo.SemesterRepository
}

func NewSemesterService(r *repo.SemesterRepository) *SemesterService {
	return &SemesterService{repo: r}
}

var (
	semIDPattern       = regexp.MustCompile(`^\d{4}[123]$`)
	tahunAjaranPattern = regexp.MustCompile(`^\d{4}/\d{4}$`)
	termSet            = map[string]struct{}{"Ganjil": {}, "Genap": {}, "Antara": {}}
)

func parseYearFromID(id string) (int, int, error) {
	if !semIDPattern.MatchString(id) {
		return 0, 0, ErrInvalidInput
	}
	yearStr := id[:4]
	termDigit := id[4]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return 0, 0, ErrInvalidInput
	}
	return year, int(termDigit - '0'), nil
}

func termOfDigit(d int) (string, error) {
	switch d {
	case 1:
		return "Ganjil", nil
	case 2:
		return "Genap", nil
	case 3:
		return "Antara", nil
	default:
		return "", ErrInvalidInput
	}
}

func validateTahunAjaranConsistent(tahunAjaran string, id string) error {
	if !tahunAjaranPattern.MatchString(tahunAjaran) {
		return ErrInvalidInput
	}
	part := strings.Split(tahunAjaran, "/")
	y1, err1 := strconv.Atoi(part[0])
	y2, err2 := strconv.Atoi(part[1])
	if err1 != nil || err2 != nil || y2 != y1+1 {
		return ErrInvalidInput
	}
	if id != "" {
		yFromID, _, err := parseYearFromID(id)
		if err != nil {
			return ErrInvalidInput
		}
		if yFromID != y1 {
			return ErrInvalidInput
		}
	}
	return nil
}

func validateTermConsistent(term string, id string) error {
	if _, ok := termSet[term]; !ok {
		return ErrInvalidInput
	}
	if id != "" {
		_, d, err := parseYearFromID(id)
		if err != nil {
			return ErrInvalidInput
		}
		t, _ := termOfDigit(d)
		if t != term {
			return ErrInvalidInput
		}
	}
	return nil
}

func validateDates(tglMulai, tglSelesai *time.Time) error {
	if tglMulai != nil && tglSelesai != nil {
		if !tglMulai.Before(*tglSelesai) {
			return ErrInvalidInput
		}
	}
	return nil
}

// List with filters and pagination
func (s *SemesterService) List(ctx context.Context, q string, tahunAjaran, term *string, limit, offset int, orderBy string) ([]model.Semester, error) {
	// sanitize
	if limit < 0 || offset < 0 {
		return nil, ErrInvalidInput
	}
	// allowed order by columns
	allowed := map[string]bool{
		"id_semester asc": true, "id_semester desc": true,
		"tahun_ajaran asc": true, "tahun_ajaran desc": true,
		"term asc": true, "term desc": true,
		"tanggal_mulai asc": true, "tanggal_mulai desc": true,
		"tanggal_selesai asc": true, "tanggal_selesai desc": true,
		"created_at asc": true, "created_at desc": true,
		"updated_at asc": true, "updated_at desc": true,
	}
	key := strings.ToLower(strings.TrimSpace(orderBy))
	if key == "" {
		key = "id_semester desc"
	}
	if !allowed[key] {
		key = "id_semester desc"
	}

	// optional filter validation
	if tahunAjaran != nil {
		if err := validateTahunAjaranConsistent(*tahunAjaran, ""); err != nil {
			return nil, ErrInvalidInput
		}
	}
	if term != nil {
		if err := validateTermConsistent(*term, ""); err != nil {
			return nil, ErrInvalidInput
		}
	}

	return s.repo.List(ctx, strings.TrimSpace(q), tahunAjaran, term, limit, offset, key)
}

func (s *SemesterService) Get(ctx context.Context, id string) (*model.Semester, error) {
	id = strings.TrimSpace(id)
	if !semIDPattern.MatchString(id) {
		return nil, ErrInvalidInput
	}
	out, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SemesterService) Create(ctx context.Context, sem *model.Semester) (*model.Semester, error) {
	if sem == nil {
		return nil, ErrInvalidInput
	}
	sem.IDSemester = strings.TrimSpace(sem.IDSemester)
	sem.TahunAjaran = strings.TrimSpace(sem.TahunAjaran)
	sem.Term = strings.TrimSpace(sem.Term)

	if !semIDPattern.MatchString(sem.IDSemester) {
		return nil, ErrInvalidInput
	}
	if err := validateTahunAjaranConsistent(sem.TahunAjaran, sem.IDSemester); err != nil {
		return nil, ErrInvalidInput
	}
	if err := validateTermConsistent(sem.Term, sem.IDSemester); err != nil {
		return nil, ErrInvalidInput
	}
	if err := validateDates(sem.TanggalMulai, sem.TanggalSelesai); err != nil {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.ExistsID(ctx, sem.IDSemester)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrConflict
	}

	out, err := s.repo.Create(ctx, sem)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateForCreate melakukan validasi lengkap yang sama dengan Create,
// namun tidak melakukan operasi insert. Jika checkConflict=true, akan
// melakukan pengecekan bahwa id_semester belum ada.
func (s *SemesterService) ValidateForCreate(ctx context.Context, sem *model.Semester, checkConflict bool) error {
	if sem == nil {
		return ErrInvalidInput
	}
	sem.IDSemester = strings.TrimSpace(sem.IDSemester)
	sem.TahunAjaran = strings.TrimSpace(sem.TahunAjaran)
	sem.Term = strings.TrimSpace(sem.Term)

	if !semIDPattern.MatchString(sem.IDSemester) {
		return ErrInvalidInput
	}
	if err := validateTahunAjaranConsistent(sem.TahunAjaran, sem.IDSemester); err != nil {
		return ErrInvalidInput
	}
	if err := validateTermConsistent(sem.Term, sem.IDSemester); err != nil {
		return ErrInvalidInput
	}
	if err := validateDates(sem.TanggalMulai, sem.TanggalSelesai); err != nil {
		return ErrInvalidInput
	}
	if checkConflict {
		exists, err := s.repo.ExistsID(ctx, sem.IDSemester)
		if err != nil {
			return err
		}
		if exists {
			return ErrConflict
		}
	}
	return nil
}

func (s *SemesterService) UpdatePut(ctx context.Context, id string, sem *model.Semester) (*model.Semester, error) {
	id = strings.TrimSpace(id)
	if !semIDPattern.MatchString(id) || sem == nil {
		return nil, ErrInvalidInput
	}

	sem.TahunAjaran = strings.TrimSpace(sem.TahunAjaran)
	sem.Term = strings.TrimSpace(sem.Term)

	if err := validateTahunAjaranConsistent(sem.TahunAjaran, id); err != nil {
		return nil, ErrInvalidInput
	}
	if err := validateTermConsistent(sem.Term, id); err != nil {
		return nil, ErrInvalidInput
	}
	if err := validateDates(sem.TanggalMulai, sem.TanggalSelesai); err != nil {
		return nil, ErrInvalidInput
	}

	// ensure exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	out, err := s.repo.UpdatePut(ctx, id, sem)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SemesterService) UpdatePatch(ctx context.Context, id string, tahunAjaran, term *string, tglMulai, tglSelesai *time.Time) (*model.Semester, error) {
	id = strings.TrimSpace(id)
	if !semIDPattern.MatchString(id) {
		return nil, ErrInvalidInput
	}

	if tahunAjaran != nil {
		v := strings.TrimSpace(*tahunAjaran)
		if err := validateTahunAjaranConsistent(v, id); err != nil {
			return nil, ErrInvalidInput
		}
		*tahunAjaran = v
	}
	if term != nil {
		v := strings.TrimSpace(*term)
		if err := validateTermConsistent(v, id); err != nil {
			return nil, ErrInvalidInput
		}
		*term = v
	}
	if err := validateDates(tglMulai, tglSelesai); err != nil {
		return nil, ErrInvalidInput
	}

	// ensure exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	out, err := s.repo.UpdatePatch(ctx, id, tahunAjaran, term, tglMulai, tglSelesai)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SemesterService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if !semIDPattern.MatchString(id) {
		return ErrInvalidInput
	}

	// ensure exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	// check relations
	hasKelas, err := s.repo.HasKelasRelated(ctx, id)
	if err != nil {
		return err
	}
	if hasKelas {
		return ErrConflict
	}
	hasKRS, err := s.repo.HasKRSRelated(ctx, id)
	if err != nil {
		return err
	}
	if hasKRS {
		return ErrConflict
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		return err
	}
	return nil
}
