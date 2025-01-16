package server

import (
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"rapid-api/pkg/data"
)

type Server struct {
	Config        Config
	DataInterface data.Interface
}

func (s *Server) Run() error {
	r := mux.NewRouter()

	if log.GetLevel() >= log.TraceLevel {
		r.Use(s.TraceLogMiddleWare)
	}

	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		MaxAge:           86400,
	})

	if log.GetLevel() >= log.TraceLevel {
		c.Log = log.New()
	}

	s.setupRestApi(r)

	handler := c.Handler(r)

	srv := &http.Server{
		Addr:    s.Config.ListenAddr,
		Handler: handler,
	}

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *Server) TraceLogMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		x, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Warn(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
		log.Tracef("Request: \n===============\n%s\n===============", string(x))

		next.ServeHTTP(w, req)
	})
}
