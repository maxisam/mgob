package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/config"
	"github.com/stefanprodan/mgob/pkg/db"
)

type HttpServer struct {
	Config  *config.AppConfig
	Modules *config.ModuleConfig
	Stats   *db.StatusStore
}

func (s *HttpServer) Start(version string) {

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	if s.Config.LogLevel == "debug" {
		r.Use(middleware.DefaultLogger)
	}

	r.Mount("/metrics", metricsRouter())

	r.Mount("/debug", middleware.Profiler())

	r.Route("/version", func(r chi.Router) {
		r.Use(appVersionCtx(version))
		r.Get("/", getVersion)
	})

	r.Route("/status", func(r chi.Router) {
		r.Use(statusCtx(s.Stats))
		r.Get("/", getStatus)
		r.Get("/{planID}", getPlanStatus)
	})

	r.Route("/backup", func(r chi.Router) {
		r.Use(configCtx(*s.Config, *s.Modules))
		r.Post("/{planID}", postBackup)
	})

	r.Route("/restore", func(r chi.Router) {
		r.Use(configCtx(*s.Config, *s.Modules))
		r.Post("/{planID}/{backupPath}", postRestore)
	})

	if s.Config.StoragePath != "" {
		FileServer(r, "/storage", http.Dir(s.Config.StoragePath))
	}

	log.Error(http.ListenAndServe(fmt.Sprintf("%s:%v", s.Config.Host, s.Config.Port), r))
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func configCtx(data config.AppConfig, modules config.ModuleConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "app.config", data))
			r = r.WithContext(context.WithValue(r.Context(), "app.modules", modules))
			next.ServeHTTP(w, r)
		})
	}
}
