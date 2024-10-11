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

var blockRegex = regexp.MustCompile(`<(?:a|abbr|acronym|address|applet|area|audioscope|b|base|basefront|bdo|bgsound|big|blackface|blink|blockquote|body|bq|br|button|caption|center|cite|code|col|colgroup|comment|dd|del|dfn|dir|div|dl|dt|em|embed|fieldset|fn|font|form|frame|frameset|h1|head|hr|html|i|iframe|ilayer|img|input|ins|isindex|kdb|keygen|label|layer|legend|li|limittext|link|listing|map|marquee|menu|meta|multicol|nobr|noembed|noframes|noscript|nosmartquotes|object|ol|optgroup|option|p|param|plaintext|pre|q|rt|ruby|s|samp|script|select|server|shadow|sidebar|small|spacer|span|strike|strong|style|sub|sup|table|tbody|td|textarea|tfoot|th|thead|title|tr|tt|u|ul|var|wbr|xml|xmp)\\W`)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Definir rota principal com lógica de manipulação
	r.HandleFunc("/", s.proxyHandler).Methods("GET", "POST")

	// Definir rota de health
	r.HandleFunc("/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/blockips", s.blockIPHandler).Methods("POST")

	return r
}

// Proxy handler to deal with requests and block IPs if necessary
func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request) {

	// Extract client IP
	clientIP := util.GetClientIP(r.Header.Get("X-Forwarded-For"))

	if util.IsIPBlocked(clientIP, s.db.GetCollection()) {
		http.Error(w, "Forbidden: IP blocked", http.StatusForbidden)
		return
	}

	// Verify subdomain and update ip format if necessary
	hostParts := strings.Split(r.Host, ".")
	if len(hostParts) > 0 && hostParts[0] == "www" {
		r.URL.Path = "/site/www" + r.URL.Path
	}

	// Validate query string
	if util.IsPayloadInvalid(r.URL.RawQuery) || util.IsPayloadInvalid(r.Form.Encode()) {
		http.Error(w, "Forbidden: Invalid content", http.StatusForbidden)
		return
	}

	// Check Paylod if http method is POST
	if r.Method == http.MethodPost {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			if blockRegex.MatchString(payload["data"].(string)) { // Supondo que o payload tem um campo "data"
				http.Error(w, "Forbidden: Invalid payload", http.StatusForbidden)
				return
			}
		}

	}

	// Send the response after validation
	resp := map[string]string{
		"message": "Request processed successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error while processing request", http.StatusInternalServerError)
		log.Printf("Error serializing JSON: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)
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

	// Insert IPs inside mongo
	_, err = collection.InsertMany(ctx, documents)
	if err != nil {
		log.Printf("Error inserting IPs: %v", err)
		http.Error(w, "Failed to insert IPs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("IPs blocked with success"))
}

// Check database Health Status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
