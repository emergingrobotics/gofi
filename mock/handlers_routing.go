package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleRouting routes routing-related requests.
func (s *Server) handleRouting(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Routing endpoints: /rest/routing
	if strings.Contains(path, "/rest/routing") {
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "routing" && i+1 < len(parts) && parts[i+1] != "" {
				id = parts[i+1]
				break
			}
		}

		switch r.Method {
		case "GET":
			if id != "" {
				s.handleGetRoute(w, r, site, id)
			} else {
				s.handleListRoutes(w, r, site)
			}
		case "POST":
			s.handleCreateRoute(w, r, site)
		case "PUT":
			if id != "" {
				s.handleUpdateRoute(w, r, site, id)
			} else {
				writeBadRequest(w, "Route ID required for update")
			}
		case "DELETE":
			if id != "" {
				s.handleDeleteRoute(w, r, site, id)
			} else {
				writeBadRequest(w, "Route ID required for delete")
			}
		default:
			writeNotFound(w)
		}
		return
	}

	writeNotFound(w)
}

// handleListRoutes returns all routes.
func (s *Server) handleListRoutes(w http.ResponseWriter, r *http.Request, site string) {
	routes := s.state.ListRoutes()

	data := make([]interface{}, len(routes))
	for i, route := range routes {
		data[i] = *route
	}

	writeAPIResponse(w, data)
}

// handleGetRoute returns a specific route by ID.
func (s *Server) handleGetRoute(w http.ResponseWriter, r *http.Request, site, id string) {
	route := s.state.GetRoute(id)
	if route == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*route})
}

// handleCreateRoute creates a new route.
func (s *Server) handleCreateRoute(w http.ResponseWriter, r *http.Request, site string) {
	var route types.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if route.ID == "" {
		route.ID = generateID()
	}

	// Set site ID
	if route.SiteID == "" {
		route.SiteID = site
	}

	s.state.AddRoute(&route)
	writeAPIResponse(w, []interface{}{route})
}

// handleUpdateRoute updates an existing route.
func (s *Server) handleUpdateRoute(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetRoute(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var route types.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID and site ID
	route.ID = id
	route.SiteID = existing.SiteID

	s.state.UpdateRoute(&route)
	writeAPIResponse(w, []interface{}{route})
}

// handleDeleteRoute deletes a route.
func (s *Server) handleDeleteRoute(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetRoute(id) == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteRoute(id)
	writeAPIResponse(w, []interface{}{})
}
