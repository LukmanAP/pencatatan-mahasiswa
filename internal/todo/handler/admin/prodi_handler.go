package admin

import (
    "errors"
    "net/http"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    "pencatatan-data-mahasiswa/internal/config"
    "pencatatan-data-mahasiswa/internal/db"
    model "pencatatan-data-mahasiswa/internal/todo/model/admin"
    repo "pencatatan-data-mahasiswa/internal/todo/repository/admin"
    service "pencatatan-data-mahasiswa/internal/todo/service/admin"
)

type ProdiHandler struct {
    service *service.ProdiService
}

func NewProdiHandler(cfg *config.Config, pool *db.Pool) *ProdiHandler {
    r := repo.NewProdiRepository(pool)
    s := service.NewProdiService(r)
    return &ProdiHandler{service: s}
}

// Request payloads

type prodiCreateRequest struct {
    IDProdi    *string `json:"id_prodi"`
    IDFakultas string  `json:"id_fakultas"`
    NamaProdi  string  `json:"nama_prodi"`
    Jenjang    string  `json:"jenjang"`
    KodeProdi  string  `json:"kode_prodi"`
    Akreditasi *string `json:"akreditasi"`
}

type prodiPutRequest struct {
    IDFakultas string  `json:"id_fakultas"`
    NamaProdi  string  `json:"nama_prodi"`
    Jenjang    string  `json:"jenjang"`
    KodeProdi  string  `json:"kode_prodi"`
    Akreditasi *string `json:"akreditasi"`
}

type prodiPatchRequest struct {
    IDFakultas *string `json:"id_fakultas"`
    NamaProdi  *string `json:"nama_prodi"`
    Jenjang    *string `json:"jenjang"`
    KodeProdi  *string `json:"kode_prodi"`
    Akreditasi *string `json:"akreditasi"`
}

// List: GET /api/v1/prodi
func (h *ProdiHandler) List(c *gin.Context) {
    q := strings.TrimSpace(c.Query("q"))
    idF := strings.TrimSpace(c.Query("id_fakultas"))
    jen := strings.TrimSpace(c.Query("jenjang"))
    akr := strings.TrimSpace(c.Query("akreditasi"))

    var idFPtr, jenPtr, akrPtr *string
    if idF != "" {
        idFPtr = &idF
    }
    if jen != "" {
        jenPtr = &jen
    }
    if akr != "" {
        akrPtr = &akr
    }

    limitStr := c.DefaultQuery("limit", "20")
    offsetStr := c.DefaultQuery("offset", "0")
    limit, err := strconv.Atoi(limitStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
        return
    }
    offset, err := strconv.Atoi(offsetStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
        return
    }

    // Sorting sanitization
    sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "nama_prodi")))
    sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "asc")))
    allowedCols := map[string]string{
        "nama_prodi":  "nama_prodi",
        "kode_prodi":  "kode_prodi",
        "jenjang":     "jenjang",
        "akreditasi":  "akreditasi",
        "created_at":  "created_at",
        "updated_at":  "updated_at",
    }
    col, ok := allowedCols[sortBy]
    if !ok {
        col = "nama_prodi"
    }
    dir := "ASC"
    if sortDir == "desc" {
        dir = "DESC"
    }
    orderBy := col + " " + dir

    data, err := h.service.List(c.Request.Context(), q, idFPtr, jenPtr, akrPtr, limit, offset, orderBy)
    if err != nil {
        if err.Error() == "invalid input" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": data})
}

// Get: GET /api/v1/prodi/:id
func (h *ProdiHandler) Get(c *gin.Context) {
    id := c.Param("id")
    out, err := h.service.Get(c.Request.Context(), id)
    if err != nil {
        if err.Error() == "invalid input" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        }
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": out})
}

// Create: POST /api/v1/prodi
func (h *ProdiHandler) Create(c *gin.Context) {
    var req prodiCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    p := &model.Prodi{
        IDFakultas: req.IDFakultas,
        NamaProdi:  req.NamaProdi,
        Jenjang:    req.Jenjang,
        KodeProdi:  req.KodeProdi,
        Akreditasi: req.Akreditasi,
    }
    if req.IDProdi != nil {
        p.IDProdi = *req.IDProdi
    }

    out, err := h.service.Create(c.Request.Context(), p)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate id, kode, or (nama+jenjang+fakultas)"})
            return
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusCreated, gin.H{"message": "created", "data": out})
}

// UpdatePut: PUT /api/v1/prodi/:id
func (h *ProdiHandler) UpdatePut(c *gin.Context) {
    id := c.Param("id")
    var req prodiPutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    p := &model.Prodi{
        IDFakultas: req.IDFakultas,
        NamaProdi:  req.NamaProdi,
        Jenjang:    req.Jenjang,
        KodeProdi:  req.KodeProdi,
        Akreditasi: req.Akreditasi,
    }

    out, err := h.service.UpdatePut(c.Request.Context(), id, p)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate kode or (nama+jenjang+fakultas)"})
            return
        default:
            if errors.Is(err, pgx.ErrNoRows) {
                c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// UpdatePatch: PATCH /api/v1/prodi/:id
func (h *ProdiHandler) UpdatePatch(c *gin.Context) {
    id := c.Param("id")
    var req prodiPatchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    out, err := h.service.UpdatePatch(c.Request.Context(), id, req.IDFakultas, req.NamaProdi, req.Jenjang, req.KodeProdi, req.Akreditasi)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate kode or (nama+jenjang+fakultas)"})
            return
        default:
            if errors.Is(err, pgx.ErrNoRows) {
                c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "updated", "data": out})
}

// Delete: DELETE /api/v1/prodi/:id
func (h *ProdiHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c.Request.Context(), id); err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete: related mahasiswa or mata_kuliah exists"})
            return
        default:
            if errors.Is(err, pgx.ErrNoRows) {
                c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "deleted", "data": gin.H{"id_prodi": id}})
}