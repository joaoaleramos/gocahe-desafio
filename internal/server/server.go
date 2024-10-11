package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"gocache-block-ips/internal/database"
)

type Server struct {
	port  int
	db    database.Service
	proxy *httputil.ReverseProxy
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	targetURL, _ := url.Parse("https://www.google.com") // URL do servidor de destino

	NewServer := &Server{
		port:  port,
		db:    database.New(),
		proxy: httputil.NewSingleHostReverseProxy(targetURL), // Inicializa o proxy reverso
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
