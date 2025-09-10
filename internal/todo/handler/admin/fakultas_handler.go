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

type Handler struct {
    service *service.Service
}

func NewHandler(cfg *config.Config, pool *db.Pool) *Handler {
    r := repo.NewFakultasRepository(pool)
    s := service.NewService(r)
    return &Handler{service: s}
}

// request payloads

type createRequest struct {
    NamaFakultas string  `json:"nama_fakultas" binding:"required"`
    Singkatan    *string `json:"singkatan"`
}

type updateRequest struct {
    NamaFakultas *string `json:"nama_fakultas"`
    Singkatan    *string `json:"singkatan"`
}

// List: GET /api/v1/fakultas?search=...&limit=..&offset=..
func (h *Handler) List(c *gin.Context) {
    search := strings.TrimSpace(c.Query("search"))
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

    data, err := h.service.List(c.Request.Context(), search, limit, offset)
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

// Get: GET /api/v1/fakultas/:id
func (h *Handler) Get(c *gin.Context) {
    id := c.Param("id")
    f, err := h.service.Get(c.Request.Context(), id)
    if err != nil {
        if err.Error() == "invalid input" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        }
        // treat not found
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": f})
}

// Create: POST /api/v1/fakultas
func (h *Handler) Create(c *gin.Context) {
    var req createRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    f := &model.Fakultas{NamaFakultas: req.NamaFakultas, Singkatan: req.Singkatan}
    out, err := h.service.Create(c.Request.Context(), f)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate id or name"})
            return
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
            return
        }
    }
    c.JSON(http.StatusCreated, gin.H{"message": "created", "data": out})
}

// Update: PUT /api/v1/fakultas/:id
func (h *Handler) Update(c *gin.Context) {
    id := c.Param("id")
    var req updateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    out, err := h.service.Update(c.Request.Context(), id, req.NamaFakultas, req.Singkatan)
    if err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusConflict, gin.H{"error": "duplicate name"})
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

// Delete: DELETE /api/v1/fakultas/:id
func (h *Handler) Delete(c *gin.Context) {
    id := c.Param("id")
    if err := h.service.Delete(c.Request.Context(), id); err != nil {
        switch err.Error() {
        case "invalid input":
            c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
            return
        case "conflict":
            c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete: related prodi exists"})
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
    c.JSON(http.StatusOK, gin.H{"message": "deleted", "data": gin.H{"id_fakultas": id}})
}
