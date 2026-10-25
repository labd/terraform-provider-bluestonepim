package fakebluestone

import (
	"net/http"
)

// writableDefinitionFields are the fields of a simple attribute definition
// that are set by create and update.
var writableDefinitionFields = []string{
	"name", "number", "dataType", "contentType", "charset", "externalSource",
	"internal", "groupId", "unit", "restrictions",
}

func (s *Server) registerDefinitions(mux *http.ServeMux) {
	mux.HandleFunc("POST /pim/definitions", s.createDefinition)
	mux.HandleFunc("GET /pim/definitions/{id}", s.getDefinition)
	mux.HandleFunc("PUT /pim/definitions/{id}", s.updateDefinition)
	mux.HandleFunc("PATCH /pim/definitions/{id}", s.updateDefinitionMetadata)
	mux.HandleFunc("DELETE /pim/definitions/{id}", s.deleteDefinition)
}

// AddDefinition stores a definition as if it was created outside Terraform.
func (s *Server) AddDefinition(id string, definition map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	definition["id"] = id
	s.definitions[id] = definition
}

// RemoveDefinition deletes a definition as if it was deleted outside Terraform.
func (s *Server) RemoveDefinition(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.definitions, id)
}

// DefinitionValues returns the name and description stored for a context
// (without fallback).
func (s *Server) DefinitionValues(id, context string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	values, _ := s.definitions[id]["contextValues"].(map[string]map[string]any)
	result := map[string]any{}
	for k, v := range values[context] {
		result[k] = v
	}
	return result
}

// Definition returns a stored definition, or nil.
func (s *Server) Definition(id string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.definitions[id]
}

func (s *Server) createDefinition(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.newID()
	body := bodyFrom(r.Context())
	definition := map[string]any{"id": id, "externalSource": false, "internal": false, "contextAware": true}
	// Only set on create, changing it has a separate endpoint
	if contextAware, ok := body["contextAware"].(bool); ok {
		definition["contextAware"] = contextAware
	}
	// Only set on create, afterwards it is updated with the metadata endpoint
	if description, ok := body["description"].(string); ok {
		definition["description"] = description
	}
	s.applyDefinitionFields(definition, body)
	if _, ok := definition["number"]; !ok {
		definition["number"] = id
	}
	s.definitions[id] = definition
	created(w, id)
}

func (s *Server) getDefinition(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	definition, ok := s.definitions[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}

	context := requestContext(r)
	if context == defaultContext {
		writeJSON(w, http.StatusOK, definition)
		return
	}

	// Names and descriptions are stored per context
	response := map[string]any{}
	for k, v := range definition {
		if k != "contextValues" {
			response[k] = v
		}
	}
	values, _ := definition["contextValues"].(map[string]map[string]any)
	for _, field := range []string{"name", "description"} {
		if v, ok := values[context][field]; ok {
			response[field] = v
		} else if r.Header.Get("context-fallback") == "false" {
			delete(response, field)
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) updateDefinition(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	definition, ok := s.definitions[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}

	for _, field := range writableDefinitionFields {
		delete(definition, field)
	}
	s.applyDefinitionFields(definition, bodyFrom(r.Context()))
	writeJSON(w, http.StatusOK, definition)
}

func (s *Server) updateDefinitionMetadata(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	definition, ok := s.definitions[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}

	body := bodyFrom(r.Context())
	target := definition
	if context := requestContext(r); context != defaultContext {
		values, _ := definition["contextValues"].(map[string]map[string]any)
		if values == nil {
			values = map[string]map[string]any{}
			definition["contextValues"] = values
		}
		if values[context] == nil {
			values[context] = map[string]any{}
		}
		target = values[context]
	}
	for _, field := range []string{"description", "name", "number"} {
		if value, ok := propertyUpdate(body, field); ok {
			if value == nil {
				delete(target, field)
			} else {
				target[field] = value
			}
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteDefinition(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.definitions[r.PathValue("id")]; !ok {
		writeError(w, http.StatusNotFound, "definition not found")
		return
	}
	delete(s.definitions, r.PathValue("id"))
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) applyDefinitionFields(definition, body map[string]any) {
	for _, field := range writableDefinitionFields {
		if value, ok := body[field]; ok && value != nil {
			definition[field] = value
		}
	}

	// The content type only applies to formatted text, where it defaults to
	// markdown. Other data types have no content type.
	if definition["dataType"] == "formatted_text" {
		if _, ok := definition["contentType"]; !ok {
			definition["contentType"] = "text/markdown"
		}
	} else {
		delete(definition, "contentType")
	}

	// Select values without an ID are new and get a generated ID
	restrictions, _ := definition["restrictions"].(map[string]any)
	enum, _ := restrictions["enum"].(map[string]any)
	values, _ := enum["values"].([]any)
	for _, v := range values {
		value, _ := v.(map[string]any)
		if id, _ := value["valueId"].(string); id == "" {
			value["valueId"] = s.newID()
		}
	}
}
