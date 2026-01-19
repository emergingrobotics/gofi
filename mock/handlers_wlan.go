package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleWLANs routes WLAN-related requests.
func (s *Server) handleWLANs(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// WLAN Group endpoints: /rest/wlangroup
	if strings.Contains(path, "/rest/wlangroup") {
		s.handleWLANGroups(w, r, site)
		return
	}

	// WLAN endpoints: /rest/wlanconf
	if strings.Contains(path, "/rest/wlanconf") {
		// Extract ID if present
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "wlanconf" && i+1 < len(parts) && parts[i+1] != "" {
				id = parts[i+1]
				break
			}
		}

		switch r.Method {
		case "GET":
			if id != "" {
				s.handleGetWLAN(w, r, site, id)
			} else {
				s.handleListWLANs(w, r, site)
			}
		case "POST":
			s.handleCreateWLAN(w, r, site)
		case "PUT":
			if id != "" {
				s.handleUpdateWLAN(w, r, site, id)
			} else {
				writeBadRequest(w, "WLAN ID required for update")
			}
		case "DELETE":
			if id != "" {
				s.handleDeleteWLAN(w, r, site, id)
			} else {
				writeBadRequest(w, "WLAN ID required for delete")
			}
		default:
			writeNotFound(w)
		}
		return
	}

	writeNotFound(w)
}

// handleListWLANs returns all WLANs for a site.
func (s *Server) handleListWLANs(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	wlans := s.state.ListWLANs()

	data := make([]interface{}, len(wlans))
	for i, wlan := range wlans {
		data[i] = *wlan
	}

	writeAPIResponse(w, data)
}

// handleGetWLAN returns a specific WLAN by ID.
func (s *Server) handleGetWLAN(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	wlan := s.state.GetWLAN(id)
	if wlan == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*wlan})
}

// handleCreateWLAN creates a new WLAN.
func (s *Server) handleCreateWLAN(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	var wlan types.WLAN
	if err := json.NewDecoder(r.Body).Decode(&wlan); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Validate required fields
	if wlan.Name == "" {
		writeBadRequest(w, "WLAN name is required")
		return
	}

	// Generate ID if not provided
	if wlan.ID == "" {
		wlan.ID = generateID()
	}
	wlan.SiteID = site

	s.state.AddWLAN(&wlan)

	writeAPIResponse(w, []interface{}{wlan})
}

// handleUpdateWLAN updates an existing WLAN.
func (s *Server) handleUpdateWLAN(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "PUT" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetWLAN(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var wlan types.WLAN
	if err := json.NewDecoder(r.Body).Decode(&wlan); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Preserve ID and site
	wlan.ID = id
	wlan.SiteID = site

	s.state.UpdateWLAN(&wlan)

	writeAPIResponse(w, []interface{}{wlan})
}

// handleDeleteWLAN deletes a WLAN.
func (s *Server) handleDeleteWLAN(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "DELETE" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetWLAN(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteWLAN(id)

	writeAPIResponse(w, []interface{}{})
}

// handleWLANGroups routes WLAN group requests.
func (s *Server) handleWLANGroups(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Extract ID if present
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "wlangroup" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetWLANGroup(w, r, site, id)
		} else {
			s.handleListWLANGroups(w, r, site)
		}
	case "POST":
		s.handleCreateWLANGroup(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdateWLANGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "WLAN group ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeleteWLANGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "WLAN group ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListWLANGroups returns all WLAN groups for a site.
func (s *Server) handleListWLANGroups(w http.ResponseWriter, r *http.Request, site string) {
	groups := s.state.ListWLANGroups()

	data := make([]interface{}, len(groups))
	for i, group := range groups {
		data[i] = *group
	}

	writeAPIResponse(w, data)
}

// handleGetWLANGroup returns a specific WLAN group by ID.
func (s *Server) handleGetWLANGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	group := s.state.GetWLANGroup(id)
	if group == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*group})
}

// handleCreateWLANGroup creates a new WLAN group.
func (s *Server) handleCreateWLANGroup(w http.ResponseWriter, r *http.Request, site string) {
	var group types.WLANGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Validate required fields
	if group.Name == "" {
		writeBadRequest(w, "WLAN group name is required")
		return
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = generateID()
	}
	group.SiteID = site

	s.state.AddWLANGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleUpdateWLANGroup updates an existing WLAN group.
func (s *Server) handleUpdateWLANGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetWLANGroup(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var group types.WLANGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Preserve ID and site
	group.ID = id
	group.SiteID = site

	s.state.UpdateWLANGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleDeleteWLANGroup deletes a WLAN group.
func (s *Server) handleDeleteWLANGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetWLANGroup(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteWLANGroup(id)

	writeAPIResponse(w, []interface{}{})
}
