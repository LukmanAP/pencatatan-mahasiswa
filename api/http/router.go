package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pencatatan-data-mahasiswa/internal/config"
	"pencatatan-data-mahasiswa/internal/db"
	admin "pencatatan-data-mahasiswa/internal/todo/handler/admin"
	auth "pencatatan-data-mahasiswa/internal/todo/handler/auth"
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
	dosenHandler := admin.NewDosenHandler(cfg, pool)
	mahasiswaHandler := admin.NewMahasiswaHandler(cfg, pool)
	semesterHandler := admin.NewSemesterHandler(cfg, pool)
	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
		}

		// Semester routes
		semesterReadGroup := v1.Group("/semester", auth.RequireAuth(cfg.JWTSecret, "admin", "operator", "dosen", "mahasiswa"))
		{
			semesterReadGroup.GET("/", semesterHandler.List)
			semesterReadGroup.GET("/:id", semesterHandler.Get)
		}
		semesterWriteGroup := v1.Group("/semester", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"))
		{
			semesterWriteGroup.POST("/", semesterHandler.Create)
			semesterWriteGroup.PUT("/:id", semesterHandler.UpdatePut)
			semesterWriteGroup.PATCH("/:id", semesterHandler.UpdatePatch)
			semesterWriteGroup.DELETE("/:id", semesterHandler.Delete)
			semesterWriteGroup.POST("/import", semesterHandler.ImportCSV)
		}

		mahasiswaGroup := v1.Group("/mahasiswa", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"))
		{
			mahasiswaGroup.GET("/", mahasiswaHandler.List)
			mahasiswaGroup.GET("/:id", mahasiswaHandler.Get)
			mahasiswaGroup.POST("/", mahasiswaHandler.Create)
			mahasiswaGroup.PUT("/:id", mahasiswaHandler.UpdatePut)
			mahasiswaGroup.PATCH("/:id", mahasiswaHandler.UpdatePatch)
			mahasiswaGroup.DELETE("/:id", mahasiswaHandler.Delete)
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

		// Dosen routes (protected by RequireAuth for admin/operator)
		dosenGroup := v1.Group("/dosen", auth.RequireAuth(cfg.JWTSecret, "admin", "operator"))
		{
			dosenGroup.GET("/", dosenHandler.List)
			dosenGroup.GET("/:id", dosenHandler.Get)
			dosenGroup.POST("/", dosenHandler.Create)
			dosenGroup.PUT("/:id", dosenHandler.UpdatePut)
			dosenGroup.PATCH("/:id", dosenHandler.UpdatePatch)
			dosenGroup.DELETE("/:id", dosenHandler.Delete)
		}
	}

	return r
}

func NewRouter() *gin.Engine {
	return gin.Default()
}
