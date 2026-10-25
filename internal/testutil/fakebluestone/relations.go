package fakebluestone

import (
	"net/http"
)

type relation struct {
	fields map[string]any
	// number of product connections using the relation
	connections int
}

func (s *Server) registerRelations(mux *http.ServeMux) {
	mux.HandleFunc("POST /pim/relations", s.createRelation)
	mux.HandleFunc("GET /pim/relations/{id}", s.getRelation)
	mux.HandleFunc("PATCH /pim/relations/{id}", s.updateRelation)
	mux.HandleFunc("POST /pim/relations/{id}/direction", s.updateRelationDirection)
	mux.HandleFunc("DELETE /pim/relations/{id}", s.deleteRelation)
}

// AddRelation stores a relation as if it was created outside Terraform.
func (s *Server) AddRelation(id string, fields map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fields["id"] = id
	s.relations[id] = &relation{fields: fields}
}

// AddRelationConnection marks the relation as used by a product connection.
func (s *Server) AddRelationConnection(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.relations[id].connections++
}

// Relation returns the fields of a relation, or nil.
func (s *Server) Relation(id string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.relations[id]; ok {
		return r.fields
	}
	return nil
}

func (s *Server) createRelation(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	body := bodyFrom(r.Context())
	if name, _ := body["name"].(string); name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	id := s.newID()
	fields := map[string]any{"id": id, "number": id, "direction": "ONE_WAY", "quantityEnabled": false}
	for _, field := range []string{"name", "number", "description", "reverseName", "direction", "quantityEnabled"} {
		if v, ok := body[field]; ok && v != nil {
			fields[field] = v
		}
	}
	s.relations[id] = &relation{fields: fields}
	created(w, id)
}

func (s *Server) getRelation(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, ok := s.relations[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "relation not found")
		return
	}

	response := map[string]any{}
	for k, v := range rel.fields {
		response[k] = v
	}
	// Unset texts are returned as empty strings
	for _, field := range []string{"description", "reverseName"} {
		if _, ok := response[field]; !ok {
			response[field] = ""
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) updateRelation(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, ok := s.relations[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "relation not found")
		return
	}

	body := bodyFrom(r.Context())
	for _, field := range []string{"name", "description", "reverseName"} {
		if value, ok := propertyUpdate(body, field); ok {
			if value == nil {
				delete(rel.fields, field)
			} else {
				rel.fields[field] = value
			}
		}
	}
	for _, field := range []string{"number", "quantityEnabled"} {
		if v, ok := body[field]; ok {
			rel.fields[field] = v
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateRelationDirection(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, ok := s.relations[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "relation not found")
		return
	}

	direction, _ := bodyFrom(r.Context())["direction"].(string)
	// Switching to two way requires the relation to have no product connections
	if direction == "TWO_WAY" && rel.fields["direction"] == "ONE_WAY" && rel.connections > 0 {
		writeError(w, http.StatusBadRequest, "relation has product connections")
		return
	}
	rel.fields["direction"] = direction
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteRelation(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, ok := s.relations[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "relation not found")
		return
	}
	if rel.connections > 0 && r.URL.Query().Get("force") != "true" {
		writeError(w, http.StatusBadRequest, "relation is in use by product connections")
		return
	}
	delete(s.relations, r.PathValue("id"))
	w.WriteHeader(http.StatusNoContent)
}
