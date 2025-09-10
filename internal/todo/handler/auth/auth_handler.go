package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"pencatatan-data-mahasiswa/internal/config"
	"pencatatan-data-mahasiswa/internal/db"
	repo "pencatatan-data-mahasiswa/internal/todo/repository/auth"
	service "pencatatan-data-mahasiswa/internal/todo/service/auth"
)

type Handler struct {
	service   *service.Service
	jwtSecret string
}

func NewHandler(cfg *config.Config, pool *db.Pool) *Handler {
	r := repo.NewRepository(pool)
	s := service.NewService(r, cfg.JWTSecret)
	return &Handler{service: s, jwtSecret: cfg.JWTSecret}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required"`
	RefID    *string `json:"ref_id"`
}

// Register godoc
// @Summary Register
// @Description Registrasi user baru
// @Accept json
// @Produce json
// @Param body body registerRequest true "Register payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	created, err := h.service.Register(c.Request.Context(), req.Username, req.Password, req.Role, req.RefID)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
			return
		case "username already taken":
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "created",
		"user": gin.H{
			"id_user":    created.IDUser,
			"username":   created.Username,
			"role":       created.Role,
			"ref_id":     created.RefID,
			"created_at": created.CreatedAt,
		},
	})
}

// Login godoc
// @Summary Login
// @Description Autentikasi user dan menghasilkan JWT
// @Accept json
// @Produce json
// @Param body body loginRequest true "Login payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	token, exp, user, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": exp,
		"user": gin.H{
			"id_user":  user.IDUser,
			"username": user.Username,
			"role":     user.Role,
			"ref_id":   user.RefID,
		},
	})
}

// RequireAuth memvalidasi Bearer JWT dan memastikan role user termasuk ke dalam allowedRoles
func RequireAuth(jwtSecret string, allowedRoles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
		claims := jwt.MapClaims{}
		tok, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		// cek role
		role, _ := claims["role"].(string)
		if len(allowed) > 0 {
			if _, ok := allowed[role]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}
		c.Set("user", claims)
		c.Next()
	}
}
