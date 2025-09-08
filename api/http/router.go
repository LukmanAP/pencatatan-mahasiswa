package http

import (
	"github.com/gin-gonic/gin"
)

type Router = gin.Engine

type route struct{ *gin.Engine }

func NewRouter() *gin.Engine {
	r := gin.Default()
	// middleware bisa ditambah di sini
	return r
}
