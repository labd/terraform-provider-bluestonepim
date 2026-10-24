package fakebluestone

import (
	"net/http"
)

func (s *Server) registerCategories(mux *http.ServeMux) {
	mux.HandleFunc("POST /pim/catalogs/nodes", s.createCategory)
	mux.HandleFunc("GET /pim/catalogs/nodes/{id}", s.getCategory)
	mux.HandleFunc("PATCH /pim/catalogs/nodes/{id}", s.updateCategory)
	mux.HandleFunc("PUT /pim/catalogs/nodes/{id}/move", s.moveCategory)
	mux.HandleFunc("DELETE /pim/catalogs/nodes/{id}", s.deleteCategory)
}

// CategoryValues returns the name and description stored for the context
// itself (without fallback), or nil when the category does not exist.
func (s *Server) CategoryValues(id, context string) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.categories[id]
	if !ok {
		return nil
	}
	result := map[string]string{}
	for k, v := range c.values[context] {
		result[k] = v
	}
	return result
}

// DeleteCategory removes a category as if it was deleted outside Terraform.
func (s *Server) DeleteCategory(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.categories, id)
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	body := bodyFrom(r.Context())
	id := s.newID()
	c := &category{
		id:     id,
		number: id,
		values: map[string]map[string]string{defaultContext: {}},
	}
	if number, _ := body["number"].(string); number != "" {
		c.number = number
	}
	if parentId, _ := body["parentId"].(string); parentId != "" {
		c.parentId = parentId
	}
	if name, _ := body["name"].(string); name != "" {
		c.values[requestContext(r)] = map[string]string{"name": name}
	}
	s.categories[id] = c
	created(w, id)
}

func (s *Server) getCategory(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.categories[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	values := map[string]string{}
	if r.Header.Get("context-fallback") != "false" {
		for k, v := range c.values[defaultContext] {
			values[k] = v
		}
	}
	for k, v := range c.values[requestContext(r)] {
		values[k] = v
	}

	response := map[string]any{"id": c.id, "number": c.number, "name": values["name"]}
	// The API returns an empty description for categories without one
	response["description"] = values["description"]
	if c.parentId != "" {
		response["parentId"] = c.parentId
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) updateCategory(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.categories[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	body := bodyFrom(r.Context())
	context := requestContext(r)
	if c.values[context] == nil {
		c.values[context] = map[string]string{}
	}
	for _, field := range []string{"name", "description"} {
		if value, ok := propertyUpdate(body, field); ok {
			if value == nil {
				delete(c.values[context], field)
			} else {
				c.values[context][field] = value.(string)
			}
		}
	}
	if number, ok := body["number"].(string); ok {
		c.number = number
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) moveCategory(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.categories[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	c.parentId, _ = bodyFrom(r.Context())["parentId"].(string)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.categories[r.PathValue("id")]; !ok {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}
	delete(s.categories, r.PathValue("id"))
	// Categories are removed asynchronously
	w.WriteHeader(http.StatusAccepted)
}
