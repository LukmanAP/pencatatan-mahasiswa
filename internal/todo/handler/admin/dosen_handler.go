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

type DosenHandler struct {
    service *service.DosenService
}

func NewDosenHandler(cfg *config.Config, pool *db.Pool) *DosenHandler {
    r := repo.NewDosenRepository((*db.Pool)(pool))
    s := service.NewDosenService(r)
    return &DosenHandler{service: s}
}

// Request payloads

type dosenCreateRequest struct {
    IDDosen         *string `json:"id_dosen"`
    NIDN            *string `json:"nidn"`
    NamaDosen       string  `json:"nama_dosen"`
    Email           *string `json:"email"`
    NoHP            *string `json:"no_hp"`
    JabatanAkademik *string `json:"jabatan_akademik"`
}

type dosenPutRequest struct {
    NIDN            *string `json:"nidn"`
    NamaDosen       string  `json:"nama_dosen"`
    Email           *string `json:"email"`
    NoHP            *string `json:"no_hp"`
    JabatanAkademik *string `json:"jabatan_akademik"`
}

type dosenPatchRequest struct {
    NIDN            *string `json:"nidn"`
    NamaDosen       *string `json:"nama_dosen"`
    Email           *string `json:"email"`
    NoHP            *string `json:"no_hp"`
    JabatanAkademik *string `json:"jabatan_akademik"`
}

// List: GET /api/v1/dosen
func (h *DosenHandler) List(c *gin.Context) {
    q := strings.TrimSpace(c.Query("q"))

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
    sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_by", "nama_dosen")))
    sortDir := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort_dir", "asc")))
    allowedCols := map[string]string{
        "nama_dosen": "nama_dosen",
        "nidn":       "nidn",
        "email":      "email",
        "created_at": "created_at",
        "updated_at": "updated_at",
    }
    col, ok := allowedCols[sortBy]
    if !ok {
        col = "nama_dosen"
    }
    dir := "ASC"
    if sortDir == "desc" {
        dir = "DESC"
    }
    orderBy := col + " " + dir

    data, err := h.service.List(c.Request.Context(), q, limit, offset, orderBy)
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

// Get: GET /api/v1/dosen/:id
func (h *DosenHandler) Get(c *gin.Context) {
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

// Create: POST /api/v1/dosen
func (h *DosenHandler) Create(c *gin.Context) {
    var req dosenCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    d := &model.Dosen{
        NamaDosen:       req.NamaDosen,
        NIDN:            req.NIDN,
        Email:           req.Email,
        NoHP:            req.NoHP,
        JabatanAkademik: req.JabatanAkademik,
    }
    if req.IDDosen != nil {
        d.IDDosen = *req.IDDosen
    }

    out, err := h.service.Create(c.Request.Context(), d)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate id, nidn, or email"})
            return
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusCreated, gin.H{"message": "created", "data": out})
}

// UpdatePut: PUT /api/v1/dosen/:id
func (h *DosenHandler) UpdatePut(c *gin.Context) {
    id := c.Param("id")
    var req dosenPutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    d := &model.Dosen{
        NamaDosen:       req.NamaDosen,
        NIDN:            req.NIDN,
        Email:           req.Email,
        NoHP:            req.NoHP,
        JabatanAkademik: req.JabatanAkademik,
    }

    out, err := h.service.UpdatePut(c.Request.Context(), id, d)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate nidn or email"})
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

// UpdatePatch: PATCH /api/v1/dosen/:id
func (h *DosenHandler) UpdatePatch(c *gin.Context) {
    id := c.Param("id")
    var req dosenPatchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    out, err := h.service.UpdatePatch(c.Request.Context(), id, req.NIDN, req.NamaDosen, req.Email, req.NoHP, req.JabatanAkademik)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate nidn or email"})
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

// Delete: DELETE /api/v1/dosen/:id
func (h *DosenHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c.Request.Context(), id); err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete: related mata_kuliah or kelas_kuliah exists"})
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
    c.JSON(http.StatusOK, gin.H{"message": "deleted", "data": gin.H{"id_dosen": id}})
}