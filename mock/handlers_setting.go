package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleSettings routes settings-related requests.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// RADIUS profile endpoints: /rest/radiusprofile
	if strings.Contains(path, "/rest/radiusprofile") {
		s.handleRADIUSProfiles(w, r, site)
		return
	}

	// Dynamic DNS endpoints: /rest/dynamicdns
	if strings.Contains(path, "/rest/dynamicdns") {
		s.handleDynamicDNS(w, r, site)
		return
	}

	// Setting endpoints: /rest/setting/{key}
	if strings.Contains(path, "/rest/setting") {
		parts := strings.Split(path, "/")
		var key string
		for i, part := range parts {
			if part == "setting" && i+1 < len(parts) && parts[i+1] != "" {
				key = parts[i+1]
				break
			}
		}

		switch r.Method {
		case "GET":
			s.handleGetSetting(w, r, site, key)
		case "PUT":
			s.handleUpdateSetting(w, r, site, key)
		default:
			writeNotFound(w)
		}
		return
	}

	writeNotFound(w)
}

// handleGetSetting returns a setting by key.
func (s *Server) handleGetSetting(w http.ResponseWriter, r *http.Request, site, key string) {
	setting := s.state.GetSetting(key)
	if setting == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*setting})
}

// handleUpdateSetting updates a setting.
func (s *Server) handleUpdateSetting(w http.ResponseWriter, r *http.Request, site, key string) {
	var setting types.Setting
	if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Ensure key matches
	setting.Key = key

	// Set site ID if not provided
	if setting.SiteID == "" {
		setting.SiteID = site
	}

	s.state.UpdateSetting(&setting)
	writeAPIResponse(w, []interface{}{setting})
}

// handleRADIUSProfiles routes RADIUS profile requests.
func (s *Server) handleRADIUSProfiles(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "radiusprofile" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetRADIUSProfile(w, r, site, id)
		} else {
			s.handleListRADIUSProfiles(w, r, site)
		}
	case "POST":
		s.handleCreateRADIUSProfile(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdateRADIUSProfile(w, r, site, id)
		} else {
			writeBadRequest(w, "RADIUS profile ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeleteRADIUSProfile(w, r, site, id)
		} else {
			writeBadRequest(w, "RADIUS profile ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListRADIUSProfiles returns all RADIUS profiles.
func (s *Server) handleListRADIUSProfiles(w http.ResponseWriter, r *http.Request, site string) {
	profiles := s.state.ListRADIUSProfiles()

	data := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		data[i] = *profile
	}

	writeAPIResponse(w, data)
}

// handleGetRADIUSProfile returns a specific RADIUS profile by ID.
func (s *Server) handleGetRADIUSProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	profile := s.state.GetRADIUSProfile(id)
	if profile == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*profile})
}

// handleCreateRADIUSProfile creates a new RADIUS profile.
func (s *Server) handleCreateRADIUSProfile(w http.ResponseWriter, r *http.Request, site string) {
	var profile types.RADIUSProfile
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

	s.state.AddRADIUSProfile(&profile)
	writeAPIResponse(w, []interface{}{profile})
}

// handleUpdateRADIUSProfile updates an existing RADIUS profile.
func (s *Server) handleUpdateRADIUSProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetRADIUSProfile(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var profile types.RADIUSProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID and site ID
	profile.ID = id
	profile.SiteID = existing.SiteID

	s.state.UpdateRADIUSProfile(&profile)
	writeAPIResponse(w, []interface{}{profile})
}

// handleDeleteRADIUSProfile deletes a RADIUS profile.
func (s *Server) handleDeleteRADIUSProfile(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetRADIUSProfile(id) == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteRADIUSProfile(id)
	writeAPIResponse(w, []interface{}{})
}

// handleDynamicDNS routes Dynamic DNS requests (GET/PUT only, singleton).
func (s *Server) handleDynamicDNS(w http.ResponseWriter, r *http.Request, site string) {
	switch r.Method {
	case "GET":
		s.handleGetDynamicDNS(w, r, site)
	case "PUT":
		s.handleUpdateDynamicDNS(w, r, site)
	default:
		writeNotFound(w)
	}
}

// handleGetDynamicDNS returns the Dynamic DNS configuration.
func (s *Server) handleGetDynamicDNS(w http.ResponseWriter, r *http.Request, site string) {
	ddns := s.state.GetDynamicDNS()
	if ddns == nil {
		// Return empty array if not configured
		writeAPIResponse(w, []interface{}{})
		return
	}

	writeAPIResponse(w, []interface{}{*ddns})
}

// handleUpdateDynamicDNS updates the Dynamic DNS configuration.
func (s *Server) handleUpdateDynamicDNS(w http.ResponseWriter, r *http.Request, site string) {
	var ddns types.DynamicDNS
	if err := json.NewDecoder(r.Body).Decode(&ddns); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if ddns.ID == "" {
		ddns.ID = generateID()
	}

	// Set site ID
	if ddns.SiteID == "" {
		ddns.SiteID = site
	}

	s.state.SetDynamicDNS(&ddns)
	writeAPIResponse(w, []interface{}{ddns})
}
