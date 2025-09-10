package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pencatatan-data-mahasiswa/internal/config"
	"pencatatan-data-mahasiswa/internal/db"
	auth "pencatatan-data-mahasiswa/internal/todo/handler/auth"
	admin "pencatatan-data-mahasiswa/internal/todo/handler/admin"
)

type Router struct{}

func NewRouterWithDeps(cfg *config.Config, pool *db.Pool) *gin.Engine {
	r := gin.Default()

	// Simple health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	authHandler := auth.NewHandler(cfg, pool)
	fakultasHandler := admin.NewHandler(cfg, pool)
	prodiHandler := admin.NewProdiHandler(cfg, pool)
	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
		}

		mahasiswaGroup := v1.Group("/mahasiswa")
		{
			mahasiswaGroup.GET("/", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "protected mahasiswa route"})
			})
		}

		// Fakultas routes (protected by RequireAuth for admin/operator)
		fakultasGroup := v1.Group("/fakultas", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"))
		{
			fakultasGroup.GET("/", fakultasHandler.List)
			fakultasGroup.GET("/:id", fakultasHandler.Get)
			fakultasGroup.POST("/", fakultasHandler.Create)
			fakultasGroup.PUT("/:id", fakultasHandler.Update)
			fakultasGroup.DELETE("/:id", fakultasHandler.Delete)
		}

		// Prodi routes (protected by RequireAuth for admin/operator)
		prodiGroup := v1.Group("/prodi", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"))
		{
			prodiGroup.GET("/", prodiHandler.List)
			prodiGroup.GET("/:id", prodiHandler.Get)
			prodiGroup.POST("/", prodiHandler.Create)
			prodiGroup.PUT("/:id", prodiHandler.UpdatePut)
			prodiGroup.PATCH("/:id", prodiHandler.UpdatePatch)
			prodiGroup.DELETE("/:id", prodiHandler.Delete)
		}
	}

	return r
}

func NewRouter() *gin.Engine {
	return gin.Default()
}
