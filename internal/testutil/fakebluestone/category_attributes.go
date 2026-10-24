package fakebluestone

import (
	"net/http"
)

func (s *Server) registerCategoryAttributes(mux *http.ServeMux) {
	base := "/pim/catalogs/nodes/{id}/attributes"
	mux.HandleFunc("GET "+base, s.listCategoryAttributes)
	mux.HandleFunc("POST "+base+"/{def}", s.createCategoryAttribute("simple"))
	mux.HandleFunc("POST "+base+"/{def}/column", s.createCategoryAttribute("column"))
	mux.HandleFunc("POST "+base+"/{def}/matrix", s.createCategoryAttribute("matrix"))
	mux.HandleFunc("POST "+base+"/{def}/dictionary", s.createCategoryAttribute("dictionary"))
	mux.HandleFunc("POST "+base+"/{def}/compound", s.createCategoryAttribute("compound"))
	mux.HandleFunc("PATCH "+base+"/{def}", s.updateCategoryAttribute(""))
	mux.HandleFunc("PUT "+base+"/{def}/column", s.updateCategoryAttribute("column"))
	mux.HandleFunc("PUT "+base+"/{def}/matrix", s.updateCategoryAttribute("matrix"))
	mux.HandleFunc("POST "+base+"/{def}/dictionary/values", s.setDictionaryValues)
	mux.HandleFunc("DELETE "+base+"/{def}", s.deleteCategoryAttribute)
}

// definitionKind returns the kind of a definition: simple, column, matrix,
// dictionary or compound.
func (s *Server) definitionKind(id string) string {
	definition, ok := s.definitions[id]
	if !ok {
		return ""
	}
	if compound, _ := definition["isCompound"].(bool); compound {
		return "compound"
	}
	switch definition["dataType"] {
	case "column", "matrix", "dictionary":
		return definition["dataType"].(string)
	}
	return "simple"
}

func (s *Server) listCategoryAttributes(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	categoryId := r.PathValue("id")
	data := []map[string]any{}
	for key, c := range s.clas {
		if key.category != categoryId {
			continue
		}
		item := map[string]any{
			"attributeDefinitionId": key.definition,
			"assignedOn":            categoryId,
			"column":                c.column,
			"matrix":                c.matrix,
			"dictionary":            c.dictionary,
		}
		if c.attributeValue != nil {
			item["attributeValue"] = *c.attributeValue
		}
		for field, value := range map[string]string{"mandatorySetOn": c.mandatorySetOn, "copySetOn": c.copySetOn, "lockedSetOn": c.lockedSetOn} {
			if value != "" {
				item[field] = value
			}
		}
		data = append(data, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) createCategoryAttribute(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()

		key := claKey{category: r.PathValue("id"), definition: r.PathValue("def")}
		// A definition of the wrong type is reported as not found
		if s.definitionKind(key.definition) != kind {
			writeError(w, http.StatusNotFound, "attribute definition not found")
			return
		}
		if _, exists := s.clas[key]; exists {
			writeError(w, http.StatusConflict, "attribute is already assigned to the category")
			return
		}

		body := bodyFrom(r.Context())
		c := &cla{}
		switch kind {
		case "simple":
			if value, ok := body["value"].(string); ok {
				c.attributeValue = &value
			}
		case "column", "matrix":
			values, ok := body[kind].(map[string]any)
			if !ok {
				writeError(w, http.StatusBadRequest, kind+" is required")
				return
			}
			if kind == "column" {
				c.column = values
			} else {
				c.matrix = values
			}
		case "compound":
			// Assigning a compound attribute also assigns its sub attributes
			subs, _ := s.definitions[key.definition]["subDefinitions"].([]any)
			for _, sub := range subs {
				subId, _ := sub.(map[string]any)["id"].(string)
				subKey := claKey{category: key.category, definition: subId}
				if _, exists := s.clas[subKey]; !exists {
					s.clas[subKey] = &cla{}
				}
			}
		}
		s.clas[key] = c
		w.WriteHeader(http.StatusAccepted)
	}
}

func (s *Server) updateCategoryAttribute(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()

		key := claKey{category: r.PathValue("id"), definition: r.PathValue("def")}
		c, ok := s.clas[key]
		if !ok {
			writeError(w, http.StatusBadRequest, "attribute is not assigned to the category")
			return
		}

		definitionKind := s.definitionKind(key.definition)
		body := bodyFrom(r.Context())

		if _, ok := body["locked"]; ok && (definitionKind == "column" || definitionKind == "matrix" || definitionKind == "compound") {
			writeError(w, http.StatusBadRequest, "locked is not supported for "+definitionKind+" definitions")
			return
		}
		if definitionKind == "compound" && r.URL.Query().Has("forceCla") {
			writeError(w, http.StatusBadRequest, "forceCla is not supported for compound definitions")
			return
		}

		setFlag := func(field string, target *string) {
			if v, ok := body[field].(bool); ok {
				*target = ""
				if v {
					*target = key.category
				}
			}
		}
		setFlag("mandatory", &c.mandatorySetOn)
		setFlag("copyAttribute", &c.copySetOn)
		setFlag("locked", &c.lockedSetOn)
		// Locking an attribute also makes it mandatory
		if c.lockedSetOn != "" {
			c.mandatorySetOn = key.category
		}

		if value, ok := propertyUpdate(body, "attributeValue"); ok {
			if value == nil {
				c.attributeValue = nil
			} else {
				v := value.(string)
				c.attributeValue = &v
			}
		}

		// The column and matrix endpoints expect the complete set of values
		switch kind {
		case "column":
			c.column, _ = body["column"].(map[string]any)
		case "matrix":
			c.matrix, _ = body["matrix"].(map[string]any)
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func (s *Server) setDictionaryValues(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := claKey{category: r.PathValue("id"), definition: r.PathValue("def")}
	if s.definitionKind(key.definition) != "dictionary" {
		writeError(w, http.StatusNotFound, "attribute definition not found")
		return
	}

	// Setting the values also assigns the attribute when it is missing
	c, ok := s.clas[key]
	if !ok {
		c = &cla{}
		s.clas[key] = c
	}
	c.dictionary, _ = bodyFrom(r.Context())["valueIds"].([]any)
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) deleteCategoryAttribute(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clas, claKey{category: r.PathValue("id"), definition: r.PathValue("def")})
	w.WriteHeader(http.StatusNoContent)
}

// CategoryAttribute returns the value of an assigned attribute as the API
// lists it, or nil when it is not assigned.
func (s *Server) CategoryAttribute(categoryId, definitionId string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.clas[claKey{category: categoryId, definition: definitionId}]
	if !ok {
		return nil
	}
	result := map[string]any{
		"column":     c.column,
		"matrix":     c.matrix,
		"dictionary": c.dictionary,
		"mandatory":  c.mandatorySetOn != "",
		"propagate":  c.copySetOn != "",
		"locked":     c.lockedSetOn != "",
	}
	if c.attributeValue != nil {
		result["value"] = *c.attributeValue
	}
	return result
}
