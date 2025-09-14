package admin

import (
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"

    "bufio"
    "encoding/csv"
    "io"
    "pencatatan-data-mahasiswa/internal/config"
    "pencatatan-data-mahasiswa/internal/db"
    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
    service "pencatatan-data-mahasiswa/internal/todo/service/admin"
)

type SemesterHandler struct {
    service *service.SemesterService
}

func NewSemesterHandler(cfg *config.Config, pool *db.Pool) *SemesterHandler {
    r := repo.NewSemesterRepository(pool)
    s := service.NewSemesterService(r)
    return &SemesterHandler{service: s}
}

// Request payloads

type semesterCreateRequest struct {
    IDSemester     string  `json:"id_semester"`
    TahunAjaran    string  `json:"tahun_ajaran"`
    Term           string  `json:"term"`
    TanggalMulai   *string `json:"tanggal_mulai"`
    TanggalSelesai *string `json:"tanggal_selesai"`
}

type semesterPutRequest struct {
    TahunAjaran    string  `json:"tahun_ajaran"`
    Term           string  `json:"term"`
    TanggalMulai   *string `json:"tanggal_mulai"`
    TanggalSelesai *string `json:"tanggal_selesai"`
}

type semesterPatchRequest struct {
    TahunAjaran    *string `json:"tahun_ajaran"`
    Term           *string `json:"term"`
    TanggalMulai   *string `json:"tanggal_mulai"`
    TanggalSelesai *string `json:"tanggal_selesai"`
}

// List: GET /api/v1/semester
func (h *SemesterHandler) List(c *gin.Context) {
    q := strings.TrimSpace(c.Query("q"))
    tahunAjaran := strings.TrimSpace(c.Query("tahun_ajaran"))
    term := strings.TrimSpace(c.Query("term"))

    var tahunAjaranPtr *string
    if tahunAjaran != "" {
        tahunAjaranPtr = &tahunAjaran
    }
    var termPtr *string
    if term != "" {
        termPtr = &term
    }

    // pagination via page & per_page (cap 100)
    pageStr := c.DefaultQuery("page", "1")
    perPageStr := c.DefaultQuery("per_page", "20")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"page": "must be >= 1"}})
        return
    }
    perPage, err := strconv.Atoi(perPageStr)
    if err != nil || perPage < 1 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"per_page": "must be >= 1"}})
        return
    }
    if perPage > 100 {
        perPage = 100
    }
    limit := perPage
    offset := (page - 1) * perPage

    // sorting sanitization
    sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "id_semester")))
    sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "desc")))
    allowedCols := map[string]string{
        "id_semester":    "id_semester",
        "tahun_ajaran":   "tahun_ajaran",
        "term":           "term",
        "tanggal_mulai":  "tanggal_mulai",
        "tanggal_selesai": "tanggal_selesai",
        "created_at":     "created_at",
        "updated_at":     "updated_at",
    }
    col, ok := allowedCols[sortBy]
    if !ok {
        col = "id_semester"
    }
    dir := "ASC"
    if sortDir == "desc" {
        dir = "DESC"
    }
    orderBy := col + " " + dir

    data, err := h.service.List(c.Request.Context(), q, tahunAjaranPtr, termPtr, limit, offset, orderBy)
    if err != nil {
        if err.Error() == "invalid input" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": data})
}

// Get: GET /api/v1/semester/:id
func (h *SemesterHandler) Get(c *gin.Context) {
    id := c.Param("id")
    out, err := h.service.Get(c.Request.Context(), id)
    if err != nil {
        if err.Error() == "invalid input" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        }
        c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": out})
}

// Create: POST /api/v1/semester
func (h *SemesterHandler) Create(c *gin.Context) {
    var req semesterCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
        return
    }

    var tMulai, tSelesai *time.Time
    if req.TanggalMulai != nil && strings.TrimSpace(*req.TanggalMulai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalMulai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_mulai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tMulai = &t
    }
    if req.TanggalSelesai != nil && strings.TrimSpace(*req.TanggalSelesai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalSelesai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_selesai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tSelesai = &t
    }

    s := &model.Semester{
        IDSemester:     strings.TrimSpace(req.IDSemester),
        TahunAjaran:    strings.TrimSpace(req.TahunAjaran),
        Term:           strings.TrimSpace(req.Term),
        TanggalMulai:   tMulai,
        TanggalSelesai: tSelesai,
    }

    out, err := h.service.Create(c.Request.Context(), s)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
            return
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
            return
        }
    }
    c.JSON(http.StatusCreated, gin.H{"message": "created", "data": out})
}

// UpdatePut: PUT /api/v1/semester/:id
func (h *SemesterHandler) UpdatePut(c *gin.Context) {
    id := c.Param("id")
    var req semesterPutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
        return
    }

    var tMulai, tSelesai *time.Time
    if req.TanggalMulai != nil && strings.TrimSpace(*req.TanggalMulai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalMulai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_mulai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tMulai = &t
    }
    if req.TanggalSelesai != nil && strings.TrimSpace(*req.TanggalSelesai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalSelesai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_selesai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tSelesai = &t
    }

    s := &model.Semester{
        TahunAjaran:    strings.TrimSpace(req.TahunAjaran),
        Term:           strings.TrimSpace(req.Term),
        TanggalMulai:   tMulai,
        TanggalSelesai: tSelesai,
    }

    out, err := h.service.UpdatePut(c.Request.Context(), id, s)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        default:
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// UpdatePatch: PATCH /api/v1/semester/:id
func (h *SemesterHandler) UpdatePatch(c *gin.Context) {
    id := c.Param("id")
    var req semesterPatchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
        return
    }

    var tMulai, tSelesai *time.Time
    if req.TanggalMulai != nil && strings.TrimSpace(*req.TanggalMulai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalMulai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_mulai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tMulai = &t
    }
    if req.TanggalSelesai != nil && strings.TrimSpace(*req.TanggalSelesai) != "" {
        t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalSelesai))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_selesai": "invalid date format (YYYY-MM-DD)"}})
            return
        }
        tSelesai = &t
    }

    out, err := h.service.UpdatePatch(c.Request.Context(), id, req.TahunAjaran, req.Term, tMulai, tSelesai)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        default:
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// Delete: DELETE /api/v1/semester/:id
func (h *SemesterHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c.Request.Context(), id); err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
            return
        case "conflict":
            c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete: related kelas_kuliah or krs exists"})
            return
        default:
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "deleted", "data": gin.H{"id_semester": id}})
}

// ImportCSV: POST /api/v1/semester/import?dry_run=true
// CSV header: id_semester,tahun_ajaran,term,tanggal_mulai,tanggal_selesai
func (h *SemesterHandler) ImportCSV(c *gin.Context) {
    dryRun := strings.ToLower(c.DefaultQuery("dry_run", "true")) == "true"

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "missing file field 'file'"})
        return
    }
    f, err := file.Open()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "cannot open uploaded file"})
        return
    }
    defer f.Close()

    reader := csv.NewReader(bufio.NewReader(f))
    reader.TrimLeadingSpace = true

    // read header
    header, err := reader.Read()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "invalid CSV header"})
        return
    }
    if len(header) < 5 || strings.ToLower(header[0]) != "id_semester" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "expected header: id_semester,tahun_ajaran,term,tanggal_mulai,tanggal_selesai"})
        return
    }

    type rowError struct {
        Line   int      `json:"line"`
        Record []string `json:"record"`
        Error  string   `json:"error"`
    }

    var (
        imported   int
        errorsList []rowError
        records    []model.Semester
        line       = 1 // header counted as line 1; data starts at 2
    )

    for {
        rec, err := reader.Read()
        if err == io.EOF {
            break
        }
        line++
        if err != nil {
            errorsList = append(errorsList, rowError{Line: line, Record: rec, Error: "invalid csv row"})
            continue
        }
        if len(rec) < 5 {
            errorsList = append(errorsList, rowError{Line: line, Record: rec, Error: "not enough columns"})
            continue
        }

        id := strings.TrimSpace(rec[0])
        thn := strings.TrimSpace(rec[1])
        term := strings.TrimSpace(rec[2])
        var tMulai, tSelesai *time.Time
        if s := strings.TrimSpace(rec[3]); s != "" {
            tt, e := time.Parse("2006-01-02", s)
            if e != nil {
                errorsList = append(errorsList, rowError{Line: line, Record: rec, Error: "tanggal_mulai invalid (YYYY-MM-DD)"})
                continue
            }
            tMulai = &tt
        }
        if s := strings.TrimSpace(rec[4]); s != "" {
            tt, e := time.Parse("2006-01-02", s)
            if e != nil {
                errorsList = append(errorsList, rowError{Line: line, Record: rec, Error: "tanggal_selesai invalid (YYYY-MM-DD)"})
                continue
            }
            tSelesai = &tt
        }

        sObj := &model.Semester{
            IDSemester:     id,
            TahunAjaran:    thn,
            Term:           term,
            TanggalMulai:   tMulai,
            TanggalSelesai: tSelesai,
        }

        // validate only (and check conflict) when dry_run or before actual insert for early error reporting
        if err := h.service.ValidateForCreate(c.Request.Context(), sObj, true); err != nil {
            errorsList = append(errorsList, rowError{Line: line, Record: rec, Error: err.Error()})
            continue
        }

        records = append(records, *sObj)
    }

    // If dry run, just return summary
    if dryRun {
        c.JSON(http.StatusOK, gin.H{
            "dry_run":      true,
            "total_rows":   len(records) + len(errorsList),
            "valid_rows":   len(records),
            "invalid_rows": len(errorsList),
            "errors":       errorsList,
        })
        return
    }

    // Insert valid records
    for i := range records {
        if _, err := h.service.Create(c.Request.Context(), &records[i]); err != nil {
            // on insert error, accumulate as error; add +2 to map record index to CSV line (skip header)
            errorsList = append(errorsList, rowError{Line: i + 2, Error: err.Error()})
            continue
        }
        imported++
    }

    c.JSON(http.StatusOK, gin.H{
        "dry_run":  false,
        "imported": imported,
        "failed":   len(errorsList),
        "errors":   errorsList,
    })
}