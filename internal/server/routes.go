package server

import (
	"context"
	"encoding/json"
	"gocache-block-ips/internal/util"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

var blockRegex = regexp.MustCompile(`(?i)<(?:a|abbr|acronym|address|applet|area|audioscope|b|base|basefront|bdo|bgsound|big|blackface|blink|blockquote|body|bq|br|button|caption|center|cite|code|col|colgroup|comment|dd|del|dfn|dir|div|dl|dt|em|embed|fieldset|fn|font|form|frame|frameset|h1|head|hr|html|i|iframe|ilayer|img|input|ins|isindex|kdb|keygen|label|layer|legend|li|limittext|link|listing|map|marquee|menu|meta|multicol|nobr|noembed|noframes|noscript|nosmartquotes|object|ol|optgroup|option|p|param|plaintext|pre|q|rt|ruby|s|samp|script|select|server|shadow|sidebar|small|spacer|span|strike|strong|style|sub|sup|table|tbody|td|textarea|tfoot|th|thead|title|tr|tt|u|ul|var|wbr|xml|xmp)`)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Define the main route with handling logic
	r.HandleFunc("/", s.proxyHandler).Methods("GET", "POST")

	// Define health check route
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/blockips", s.blockIPHandler).Methods("POST")

	return r
}

// Proxy handler to deal with requests and block IPs if necessary
func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract client IP
	log.Println(r.Method)
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	clientIP := util.GetClientIP(ip)

	// Check if the IP is blocked
	if util.IsIPBlocked(clientIP, s.db.GetCollection()) {
		http.Error(w, "Forbidden: IP blocked", http.StatusForbidden)
		return
	}

	// Check subdomain and update IP format if necessary
	hostParts := strings.Split(r.Host, ".")
	if len(hostParts) > 0 && hostParts[0] == "www" {
		r.URL.Path = "/site/www" + r.URL.Path
	}

	// Validate query string
	if util.IsPayloadInvalid(r.URL.RawQuery) || util.IsPayloadInvalid(r.Form.Encode()) {
		http.Error(w, "Forbidden: Invalid content", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Error decoding JSON: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		data, ok := payload["data"].(string)
		log.Println(ok)
		if !ok {
			log.Println("Payload does not contain 'data' field or is not a string")
			http.Error(w, "Forbidden: Invalid payload format", http.StatusForbidden)
			return
		}
		if blockRegex.MatchString(data) {
			log.Printf("Blocked payload content: %s", data)
			http.Error(w, "Forbidden: Invalid payload", http.StatusForbidden)
			return
		}
	}
	// Redirect the request to the target server
	s.proxy.ServeHTTP(w, r)
}

// Block IPs handler
func (s *Server) blockIPHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ips struct {
		IPs []string `json:"ips"`
	}

	err := json.NewDecoder(r.Body).Decode(&ips)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var documents []interface{}
	for _, ip := range ips.IPs {
		documents = append(documents, bson.M{"ip": ip})
	}

	collection := s.db.GetCollection()

	// Insert IPs into MongoDB
	_, err = collection.InsertMany(ctx, documents)
	if err != nil {
		log.Printf("Error inserting IPs: %v", err)
		http.Error(w, "Failed to insert IPs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("IPs blocked successfully"))
}

// Check database health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("Error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
