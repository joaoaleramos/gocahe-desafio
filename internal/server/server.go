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
	targetURL, _ := url.Parse("https://www.google.com") // Target server URL

	NewServer := &Server{
		port:  port,
		db:    database.New(),
		proxy: httputil.NewSingleHostReverseProxy(targetURL), // Initialize the reverse proxy
	}

	// Declare server configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
