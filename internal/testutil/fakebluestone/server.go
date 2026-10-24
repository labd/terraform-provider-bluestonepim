// Package fakebluestone is an in-memory fake of the parts of the Bluestone PIM
// API used by the provider, for testing the provider with Terraform without a
// Bluestone environment.
//
// The behaviour follows the Bluestone API documentation. It only proves the
// provider works against our reading of that documentation, not against the
// real API.
package fakebluestone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const defaultContext = "en"

// Request is a write request received by the fake, for asserting which calls
// the provider makes.
type Request struct {
	Method  string
	Path    string
	Context string
	Body    map[string]any
}

type Server struct {
	*httptest.Server

	mu          sync.Mutex
	nextID      int
	definitions map[string]map[string]any
	categories  map[string]*category
	clas        map[claKey]*cla
	requests    []Request
}

type category struct {
	id       string
	number   string
	parentId string
	// values per context: name and description
	values map[string]map[string]string
}

type claKey struct {
	category   string
	definition string
}

type cla struct {
	attributeValue *string
	column         map[string]any
	matrix         map[string]any
	dictionary     []any
	mandatorySetOn string
	copySetOn      string
	lockedSetOn    string
}

// New starts a fake server, which is closed when the test finishes.
func New(t *testing.T) *Server {
	t.Helper()

	s := &Server{
		definitions: map[string]map[string]any{},
		categories:  map[string]*category{},
		clas:        map[claKey]*cla{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /token", s.token)
	s.registerDefinitions(mux)
	s.registerCategories(mux)
	s.registerCategoryAttributes(mux)

	s.Server = httptest.NewServer(s.logWrites(mux))
	t.Cleanup(s.Close)
	return s
}

// ProviderConfig returns the provider block pointing to the fake.
func (s *Server) ProviderConfig() string {
	return fmt.Sprintf(`
provider "bluestonepim" {
  client_id     = "test"
  client_secret = "test"
  auth_url      = "%[1]s/token"
  api_url       = "%[1]s"
}
`, s.URL)
}

// Requests returns the write requests received so far.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

func (s *Server) newID() string {
	s.nextID++
	return fmt.Sprintf("%024x", s.nextID)
}

func (s *Server) token(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"access_token": "test", "token_type": "Bearer", "expires_in": 3600})
}

func (s *Server) logWrites(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.URL.Path != "/token" {
			var body map[string]any
			if r.Body != nil && r.ContentLength != 0 {
				_ = json.NewDecoder(r.Body).Decode(&body)
			}
			r.Body = http.NoBody

			s.mu.Lock()
			s.requests = append(s.requests, Request{Method: r.Method, Path: r.URL.Path, Context: requestContext(r), Body: body})
			s.mu.Unlock()

			r = r.WithContext(withBody(r.Context(), body))
		}
		next.ServeHTTP(w, r)
	})
}

func requestContext(r *http.Request) string {
	if c := r.Header.Get("context"); c != "" {
		return c
	}
	return defaultContext
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func created(w http.ResponseWriter, id string) {
	w.Header().Set("Resource-Id", id)
	w.WriteHeader(http.StatusCreated)
}
