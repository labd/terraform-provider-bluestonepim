package fakebluestone

import (
	"net/http"
	"slices"
	"sort"
	"strconv"
)

type attributeGroup struct {
	id          string
	name        string
	number      string
	orderByName bool
	// manual order of the definitions in the group
	definitionOrder []string
}

func (s *Server) registerAttributeGroups(mux *http.ServeMux) {
	mux.HandleFunc("GET /pim/attributeGroups", s.listAttributeGroups)
	mux.HandleFunc("POST /pim/attributeGroups", s.createAttributeGroup)
	mux.HandleFunc("PUT /pim/attributeGroups/{id}", s.moveAttributeGroup)
	mux.HandleFunc("PUT /pim/attributeGroups/{id}/name", s.renameAttributeGroup)
	mux.HandleFunc("PUT /pim/attributeGroups/{id}/number", s.updateAttributeGroupNumber)
	mux.HandleFunc("DELETE /pim/attributeGroups/{id}", s.deleteAttributeGroup)
	mux.HandleFunc("GET /pim/attributeGroups/{id}/definitions", s.listGroupDefinitions)
	mux.HandleFunc("PUT /pim/attributeGroups/{id}/definitions/order", s.setGroupOrderStrategy)
	mux.HandleFunc("PUT /pim/attributeGroups/{id}/definitions/{definitionId}", s.moveGroupDefinition)
}

// AddAttributeGroup stores a group at the end of the order, as if it was
// created outside Terraform.
func (s *Server) AddAttributeGroup(id, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups = append(s.groups, &attributeGroup{id: id, name: name, number: id})
}

// AttributeGroupOrder returns the IDs of the groups in display order.
func (s *Server) AttributeGroupOrder() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := []string{}
	for _, g := range s.groups {
		ids = append(ids, g.id)
	}
	return ids
}

// GroupDefinitionOrder returns the IDs of the definitions in a group in
// display order.
func (s *Server) GroupDefinitionOrder(groupId string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, g := s.group(groupId)
	if g == nil {
		return nil
	}
	return s.groupDefinitions(g)
}

func (s *Server) group(id string) (int, *attributeGroup) {
	for i, g := range s.groups {
		if g.id == id {
			return i, g
		}
	}
	return -1, nil
}

// groupDefinitions returns the definitions with the group as groupId: by name
// when ordered by name, otherwise in the manual order, with definitions not
// moved yet at the end in creation order.
func (s *Server) groupDefinitions(g *attributeGroup) []string {
	var ids []string
	for _, id := range g.definitionOrder {
		if d, ok := s.definitions[id]; ok && d["groupId"] == g.id {
			ids = append(ids, id)
		}
	}
	var rest []string
	for id, d := range s.definitions {
		if d["groupId"] == g.id && !slices.Contains(ids, id) {
			rest = append(rest, id)
		}
	}
	sort.Strings(rest) // IDs are generated in increasing order
	ids = append(ids, rest...)

	if g.orderByName {
		sort.SliceStable(ids, func(i, j int) bool {
			return s.definitions[ids[i]]["name"].(string) < s.definitions[ids[j]]["name"].(string)
		})
	}
	return ids
}

func (s *Server) listAttributeGroups(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := []map[string]any{}
	for _, g := range s.groups {
		data = append(data, map[string]any{"id": g.id, "name": g.name, "number": g.number, "orderByName": g.orderByName})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) createAttributeGroup(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	body := bodyFrom(r.Context())
	id := s.newID()
	g := &attributeGroup{id: id, number: id, orderByName: true}
	g.name, _ = body["name"].(string)
	if number, _ := body["number"].(string); number != "" {
		g.number = number
	}
	s.groups = append(s.groups, g)
	created(w, id)
}

func (s *Server) moveAttributeGroup(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, g := s.group(r.PathValue("id"))
	position, err := strconv.Atoi(r.URL.Query().Get("targetPosition"))
	if g == nil || err != nil || position < 0 || position >= len(s.groups) {
		writeError(w, http.StatusBadRequest, "invalid group or position")
		return
	}
	s.groups = slices.Insert(slices.Delete(s.groups, index, index+1), position, g)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) renameAttributeGroup(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	g.name, _ = bodyFrom(r.Context())["name"].(string)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateAttributeGroupNumber(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	g.number, _ = bodyFrom(r.Context())["number"].(string)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteAttributeGroup(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	s.groups = slices.Delete(s.groups, index, index+1)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listGroupDefinitions(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	data := []map[string]any{}
	for _, id := range s.groupDefinitions(g) {
		data = append(data, s.definitions[id])
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) setGroupOrderStrategy(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	g.orderByName = r.URL.Query().Get("orderByName") == "true"
	w.WriteHeader(http.StatusOK)
}

func (s *Server) moveGroupDefinition(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, g := s.group(r.PathValue("id"))
	if g == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	order := s.groupDefinitions(g)
	index := slices.Index(order, r.PathValue("definitionId"))
	position, err := strconv.Atoi(r.URL.Query().Get("targetPosition"))
	if index == -1 || err != nil || position < 0 || position >= len(order) {
		writeError(w, http.StatusBadRequest, "invalid definition or position")
		return
	}
	id := order[index]
	g.definitionOrder = slices.Insert(slices.Delete(order, index, index+1), position, id)
	w.WriteHeader(http.StatusOK)
}
