package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleTrafficRules routes traffic rule requests (v2 API).
func (s *Server) handleTrafficRules(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Extract ID if present: /v2/api/site/{site}/trafficrule/{id}
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "trafficrule" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetTrafficRule(w, r, site, id)
		} else {
			s.handleListTrafficRules(w, r, site)
		}
	case "POST":
		s.handleCreateTrafficRule(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdateTrafficRule(w, r, site, id)
		} else {
			writeBadRequest(w, "Traffic rule ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeleteTrafficRule(w, r, site, id)
		} else {
			writeBadRequest(w, "Traffic rule ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListTrafficRules returns all traffic rules for a site.
func (s *Server) handleListTrafficRules(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	rules := s.state.ListTrafficRules()

	data := make([]interface{}, len(rules))
	for i, rule := range rules {
		data[i] = *rule
	}

	writeAPIResponse(w, data)
}

// handleGetTrafficRule returns a specific traffic rule by ID.
func (s *Server) handleGetTrafficRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	rule := s.state.GetTrafficRule(id)
	if rule == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*rule})
}

// handleCreateTrafficRule creates a new traffic rule.
func (s *Server) handleCreateTrafficRule(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	var rule types.TrafficRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Validate required fields
	if rule.Name == "" {
		writeBadRequest(w, "Traffic rule name is required")
		return
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = generateID()
	}
	rule.SiteID = site

	s.state.AddTrafficRule(&rule)

	writeAPIResponse(w, []interface{}{rule})
}

// handleUpdateTrafficRule updates an existing traffic rule.
// Note: PUT returns 201 for traffic rules (v2 API quirk).
func (s *Server) handleUpdateTrafficRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "PUT" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetTrafficRule(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var rule types.TrafficRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Preserve ID and site
	rule.ID = id
	rule.SiteID = site

	s.state.UpdateTrafficRule(&rule)

	// Note: v2 API returns 201 for PUT operations
	writeAPIResponseWithStatus(w, []interface{}{rule}, http.StatusCreated)
}

// handleDeleteTrafficRule deletes a traffic rule.
func (s *Server) handleDeleteTrafficRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "DELETE" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetTrafficRule(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteTrafficRule(id)

	writeAPIResponse(w, []interface{}{})
}
