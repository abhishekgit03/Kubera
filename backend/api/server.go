package api

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/abhishekgit03/kubera/collector"
	"github.com/abhishekgit03/kubera/gemini"
	"github.com/abhishekgit03/kubera/prompt"
)

type Server struct {
	router     *gin.Engine
	collector  *collector.Collector
	gemini     *gemini.Client
	frontendFS embed.FS
}

func NewServer(col *collector.Collector, gem *gemini.Client, frontendFS embed.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
	}))

	s := &Server{router: router, collector: col, gemini: gem, frontendFS: frontendFS}
	s.registerRoutes()
	return s
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) registerRoutes() {
	s.router.GET("/healthz", s.handleHealth)
	s.router.GET("/api/snapshot", s.handleSnapshot)
	s.router.GET("/api/analyze", s.handleAnalyze)
	s.router.NoRoute(s.handleStatic)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleSnapshot(c *gin.Context) {
	snap := s.collector.Latest()
	if snap == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no data collected yet"})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (s *Server) handleAnalyze(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	snap := s.collector.Latest()
	if snap == nil {
		c.SSEvent("error", "no data collected yet")
		c.Writer.Flush()
		return
	}

	p := prompt.Build(snap)
	ch := make(chan string, 32)

	go func() {
		// StreamNarrative closes ch when done or on error
		_ = s.gemini.StreamNarrative(c.Request.Context(), p, ch)
	}()

	c.Stream(func(w io.Writer) bool {
		token, ok := <-ch
		if !ok {
			c.SSEvent("done", "")
			return false
		}
		c.SSEvent("token", token)
		return true
	})
}

func (s *Server) handleStatic(c *gin.Context) {
	sub, err := fs.Sub(s.frontendFS, "frontend/dist")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// For SPA routes that don't map to real files, serve index.html
	if _, err := sub.Open(path); err != nil {
		data, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		return
	}

	http.FileServer(http.FS(sub)).ServeHTTP(c.Writer, c.Request)
}
