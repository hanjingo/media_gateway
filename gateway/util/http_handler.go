package util

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ArgHash string = "hash"
	ArgTag  string = "tag"
)

type HttpHandler struct {
	conf   *HttpConfig
	brun   bool
	router *gin.Engine
}

func NewHttpHandler(conf *HttpConfig) *HttpHandler {
	return &HttpHandler{
		brun:   false,
		conf:   conf,
		router: gin.Default(),
	}
}

func (h *HttpHandler) Service() {
	if h.brun {
		return
	}

	h.brun = true
	path := h.conf.StaticSrcPath
	if path == "" {
		path = "./statics/*"
	}
	h.router.LoadHTMLGlob(path)
	go h.router.Run(h.conf.Addr)
}

func (h *HttpHandler) SetHandler(typ string, url string, f func(ctx *gin.Context)) error {
	switch strings.ToUpper(typ) {
	case "GET":
		h.router.GET(url, f)
		return nil
	case "POST":
		h.router.POST(url, f)
		return nil
	}
	return fmt.Errorf("not support http request type:%v", typ)
}
