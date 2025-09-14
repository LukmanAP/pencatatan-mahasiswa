package admin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"pencatatan-data-mahasiswa/internal/config"
	"pencatatan-data-mahasiswa/internal/db"
	model "pencatatan-data-mahasiswa/internal/todo/model/admin"
	repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
	service "pencatatan-data-mahasiswa/internal/todo/service/admin"
)

type MahasiswaHandler struct {
	service *service.MahasiswaService
}

func NewMahasiswaHandler(cfg *config.Config, pool *db.Pool) *MahasiswaHandler {
	r := repo.NewMahasiswaRepository(pool)
	s := service.NewMahasiswaService(r)
	return &MahasiswaHandler{service: s}
}

// Request payloads

type mhsCreateRequest struct {
	IDMahasiswa  string  `json:"id_mahasiswa"`
	IDProdi      string  `json:"id_prodi"`
	NIK          *string `json:"nik"`
	NamaLengkap  string  `json:"nama_lengkap"`
	JenisKelamin string  `json:"jenis_kelamin"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"` // format YYYY-MM-DD
	Alamat       *string `json:"alamat"`
	Email        *string `json:"email"`
	NoHP         *string `json:"no_hp"`
	TahunMasuk   int     `json:"tahun_masuk"`
	Status       *string `json:"status"`
}

type mhsPutRequest struct {
	IDProdi      string  `json:"id_prodi"`
	NIK          *string `json:"nik"`
	NamaLengkap  string  `json:"nama_lengkap"`
	JenisKelamin string  `json:"jenis_kelamin"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"`
	Alamat       *string `json:"alamat"`
	Email        *string `json:"email"`
	NoHP         *string `json:"no_hp"`
	TahunMasuk   int     `json:"tahun_masuk"`
	Status       *string `json:"status"`
}

type mhsPatchRequest struct {
	IDProdi      *string `json:"id_prodi"`
	NIK          *string `json:"nik"`
	NamaLengkap  *string `json:"nama_lengkap"`
	JenisKelamin *string `json:"jenis_kelamin"`
	TempatLahir  *string `json:"tempat_lahir"`
	TanggalLahir *string `json:"tanggal_lahir"`
	Alamat       *string `json:"alamat"`
	Email        *string `json:"email"`
	NoHP         *string `json:"no_hp"`
	TahunMasuk   *int    `json:"tahun_masuk"`
	Status       *string `json:"status"`
}

// List: GET /api/v1/mahasiswa
func (h *MahasiswaHandler) List(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	idProdi := strings.TrimSpace(c.Query("id_prodi"))
	angkatanStr := strings.TrimSpace(c.Query("angkatan"))
	status := strings.TrimSpace(c.Query("status"))

	var idProdiPtr *string
	if idProdi != "" {
		idProdiPtr = &idProdi
	}
	var angkatanPtr *int
	if angkatanStr != "" {
		if v, err := strconv.Atoi(angkatanStr); err == nil {
			angkatanPtr = &v
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"angkatan": "must be integer"}})
			return
		}
	}
	var statusPtr *string
	if status != "" {
		statusPtr = &status
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
	sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "nama_lengkap")))
	sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "asc")))
	allowedCols := map[string]string{
		"nama_lengkap": "nama_lengkap",
		"tahun_masuk":  "tahun_masuk",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
	}
	col, ok := allowedCols[sortBy]
	if !ok {
		col = "nama_lengkap"
	}
	dir := "ASC"
	if sortDir == "desc" {
		dir = "DESC"
	}
	orderBy := col + " " + dir

	data, err := h.service.List(c.Request.Context(), q, idProdiPtr, angkatanPtr, statusPtr, limit, offset, orderBy)
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

// Get: GET /api/v1/mahasiswa/:id
func (h *MahasiswaHandler) Get(c *gin.Context) {
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

// Create: POST /api/v1/mahasiswa
func (h *MahasiswaHandler) Create(c *gin.Context) {
	var req mhsCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}

	var tglPtr *time.Time
	if req.TanggalLahir != nil && strings.TrimSpace(*req.TanggalLahir) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalLahir))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_lahir": "invalid date format (YYYY-MM-DD)"}})
			return
		}
		tglPtr = &t
	}

	m := &model.Mahasiswa{
		IDMahasiswa:  strings.TrimSpace(req.IDMahasiswa),
		IDProdi:      strings.TrimSpace(req.IDProdi),
		NIK:          req.NIK,
		NamaLengkap:  strings.TrimSpace(req.NamaLengkap),
		JenisKelamin: strings.TrimSpace(req.JenisKelamin),
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglPtr,
		Alamat:       req.Alamat,
		Email:        req.Email,
		NoHP:         req.NoHP,
		TahunMasuk:   req.TahunMasuk,
	}
	if req.Status != nil {
		m.Status = strings.TrimSpace(*req.Status)
	}

	out, err := h.service.Create(c.Request.Context(), m)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		case "unprocessable":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unprocessable", "fields": gin.H{"id_prodi": "not found"}})
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

// UpdatePut: PUT /api/v1/mahasiswa/:id
func (h *MahasiswaHandler) UpdatePut(c *gin.Context) {
	id := c.Param("id")
	var req mhsPutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}

	var tglPtr *time.Time
	if req.TanggalLahir != nil && strings.TrimSpace(*req.TanggalLahir) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalLahir))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_lahir": "invalid date format (YYYY-MM-DD)"}})
			return
		}
		tglPtr = &t
	}

	m := &model.Mahasiswa{
		IDProdi:      strings.TrimSpace(req.IDProdi),
		NIK:          req.NIK,
		NamaLengkap:  strings.TrimSpace(req.NamaLengkap),
		JenisKelamin: strings.TrimSpace(req.JenisKelamin),
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglPtr,
		Alamat:       req.Alamat,
		Email:        req.Email,
		NoHP:         req.NoHP,
		TahunMasuk:   req.TahunMasuk,
	}
	if req.Status != nil {
		m.Status = strings.TrimSpace(*req.Status)
	}

	out, err := h.service.UpdatePut(c.Request.Context(), id, m)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		case "unprocessable":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unprocessable", "fields": gin.H{"id_prodi": "not found"}})
			return
		case "conflict":
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
			return
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// UpdatePatch: PATCH /api/v1/mahasiswa/:id
func (h *MahasiswaHandler) UpdatePatch(c *gin.Context) {
	id := c.Param("id")
	var req mhsPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}

	var tglPtr *time.Time
	if req.TanggalLahir != nil && strings.TrimSpace(*req.TanggalLahir) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TanggalLahir))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "fields": gin.H{"tanggal_lahir": "invalid date format (YYYY-MM-DD)"}})
			return
		}
		tglPtr = &t
	}

	out, err := h.service.UpdatePatch(c.Request.Context(), id, req.IDProdi, req.NIK, req.NamaLengkap, req.JenisKelamin, req.TempatLahir, req.Alamat, req.Email, req.NoHP, req.Status, tglPtr, req.TahunMasuk)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		case "unprocessable":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unprocessable", "fields": gin.H{"id_prodi": "not found"}})
			return
		case "conflict":
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
			return
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// Delete: DELETE /api/v1/mahasiswa/:id
func (h *MahasiswaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		switch err.Error() {
		case "invalid input":
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		case "conflict":
			c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
			return
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted", "data": gin.H{"id_mahasiswa": id}})
}
