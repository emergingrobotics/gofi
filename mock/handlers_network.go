package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleNetworks routes network-related requests.
func (s *Server) handleNetworks(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Check if this is a specific network ID request
	if strings.Contains(path, "/rest/networkconf/") {
		// Extract ID from path
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "networkconf" && i+1 < len(parts) {
				id := parts[i+1]
				switch r.Method {
				case "GET":
					s.handleGetNetwork(w, r, site, id)
				case "PUT":
					s.handleUpdateNetwork(w, r, site, id)
				case "DELETE":
					s.handleDeleteNetwork(w, r, site, id)
				default:
					writeBadRequest(w, "Method not allowed")
				}
				return
			}
		}
	}

	// List or create
	if strings.HasSuffix(path, "/rest/networkconf") {
		switch r.Method {
		case "GET":
			s.handleListNetworks(w, r, site)
		case "POST":
			s.handleCreateNetwork(w, r, site)
		default:
			writeBadRequest(w, "Method not allowed")
		}
		return
	}

	writeNotFound(w)
}

// handleListNetworks returns all networks for a site.
func (s *Server) handleListNetworks(w http.ResponseWriter, r *http.Request, site string) {
	networks := s.state.ListNetworks()

	// Convert to interface slice
	data := make([]interface{}, len(networks))
	for i, network := range networks {
		data[i] = *network
	}

	writeAPIResponse(w, data)
}

// handleGetNetwork returns a specific network.
func (s *Server) handleGetNetwork(w http.ResponseWriter, r *http.Request, site, id string) {
	network, exists := s.state.GetNetwork(id)
	if !exists {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*network})
}

// handleCreateNetwork creates a new network.
func (s *Server) handleCreateNetwork(w http.ResponseWriter, r *http.Request, site string) {
	var network types.Network
	if err := json.NewDecoder(r.Body).Decode(&network); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if network.ID == "" {
		network.ID = generateToken()
	}
	network.SiteID = site

	s.state.AddNetwork(&network)

	writeAPIResponse(w, []interface{}{network})
}

// handleUpdateNetwork updates a network.
func (s *Server) handleUpdateNetwork(w http.ResponseWriter, r *http.Request, site, id string) {
	// Get existing network
	existing, exists := s.state.GetNetwork(id)
	if !exists {
		writeNotFound(w)
		return
	}

	// Parse update request
	var updateReq types.Network
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID and site
	updateReq.ID = existing.ID
	updateReq.SiteID = existing.SiteID

	// Save updated network
	s.state.AddNetwork(&updateReq)

	writeAPIResponse(w, []interface{}{updateReq})
}

// handleDeleteNetwork deletes a network.
func (s *Server) handleDeleteNetwork(w http.ResponseWriter, r *http.Request, site, id string) {
	_, exists := s.state.GetNetwork(id)
	if !exists {
		writeNotFound(w)
		return
	}

	s.state.DeleteNetwork(id)

	writeAPIResponse(w, []interface{}{})
}
