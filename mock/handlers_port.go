package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handlePorts routes port-related requests (forwarding and profiles).
func (s *Server) handlePorts(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Port forwarding endpoints: /rest/portforward
	if strings.Contains(path, "/rest/portforward") {
		s.handlePortForward(w, r, site)
		return
	}

	// Port profile endpoints: /rest/portconf
	if strings.Contains(path, "/rest/portconf") {
		s.handlePortProfile(w, r, site)
		return
	}

	writeNotFound(w)
}

// handlePortForward routes port forwarding requests.
func (s *Server) handlePortForward(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "portforward" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetPortForward(w, r, site, id)
		} else {
			s.handleListPortForwards(w, r, site)
		}
	case "POST":
		s.handleCreatePortForward(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdatePortForward(w, r, site, id)
		} else {
			writeBadRequest(w, "Port forward ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeletePortForward(w, r, site, id)
		} else {
			writeBadRequest(w, "Port forward ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListPortForwards returns all port forwards.
func (s *Server) handleListPortForwards(w http.ResponseWriter, r *http.Request, site string) {
	forwards := s.state.ListPortForwards()

	data := make([]interface{}, len(forwards))
	for i, forward := range forwards {
		data[i] = *forward
	}

	writeAPIResponse(w, data)
}

// handleGetPortForward returns a specific port forward by ID.
func (s *Server) handleGetPortForward(w http.ResponseWriter, r *http.Request, site, id string) {
	forward := s.state.GetPortForward(id)
	if forward == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*forward})
}

// handleCreatePortForward creates a new port forward.
func (s *Server) handleCreatePortForward(w http.ResponseWriter, r *http.Request, site string) {
	var forward types.PortForward
	if err := json.NewDecoder(r.Body).Decode(&forward); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if forward.ID == "" {
		forward.ID = generateID()
	}

	// Set site ID
	if forward.SiteID == "" {
		forward.SiteID = site
	}

	s.state.AddPortForward(&forward)
	writeAPIResponse(w, []interface{}{forward})
}

// handleUpdatePortForward updates an existing port forward.
func (s *Server) handleUpdatePortForward(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetPortForward(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var forward types.PortForward
	if err := json.NewDecoder(r.Body).Decode(&forward); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID and site ID
	forward.ID = id
	forward.SiteID = existing.SiteID

	s.state.UpdatePortForward(&forward)
	writeAPIResponse(w, []interface{}{forward})
}

// handleDeletePortForward deletes a port forward.
func (s *Server) handleDeletePortForward(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetPortForward(id) == nil {
		writeNotFound(w)
		return
	}

	s.state.DeletePortForward(id)
	writeAPIResponse(w, []interface{}{})
}

// handlePortProfile routes port profile requests.
func (s *Server) handlePortProfile(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "portconf" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetPortProfile(w, r, site, id)
		} else {
			s.handleListPortProfiles(w, r, site)
		}
	case "POST":
		s.handleCreatePortProfile(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdatePortProfile(w, r, site, id)
		} else {
			writeBadRequest(w, "Port profile ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeletePortProfile(w, r, site, id)
		} else {
			writeBadRequest(w, "Port profile ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListPortProfiles returns all port profiles.
func (s *Server) handleListPortProfiles(w http.ResponseWriter, r *http.Request, site string) {
	profiles := s.state.ListPortProfiles()

	data := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		data[i] = *profile
	}

	writeAPIResponse(w, data)
}

// handleGetPortProfile returns a specific port profile by ID.
func (s *Server) handleGetPortProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	profile := s.state.GetPortProfile(id)
	if profile == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*profile})
}

// handleCreatePortProfile creates a new port profile.
func (s *Server) handleCreatePortProfile(w http.ResponseWriter, r *http.Request, site string) {
	var profile types.PortProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if profile.ID == "" {
		profile.ID = generateID()
	}

	// Set site ID
	if profile.SiteID == "" {
		profile.SiteID = site
	}

	s.state.AddPortProfile(&profile)
	writeAPIResponse(w, []interface{}{profile})
}

// handleUpdatePortProfile updates an existing port profile.
func (s *Server) handleUpdatePortProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetPortProfile(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var profile types.PortProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID and site ID
	profile.ID = id
	profile.SiteID = existing.SiteID

	s.state.UpdatePortProfile(&profile)
	writeAPIResponse(w, []interface{}{profile})
}

// handleDeletePortProfile deletes a port profile.
func (s *Server) handleDeletePortProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetPortProfile(id) == nil {
		writeNotFound(w)
		return
	}

	s.state.DeletePortProfile(id)
	writeAPIResponse(w, []interface{}{})
}
