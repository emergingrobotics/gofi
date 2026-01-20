package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleSites routes site-related requests.
func (s *Server) handleSites(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// /api/self/sites - list or create sites
	if strings.HasSuffix(path, "/api/self/sites") {
		if r.Method == "GET" {
			s.handleListSites(w, r)
		} else if r.Method == "POST" {
			s.handleCreateSite(w, r)
		} else {
			writeBadRequest(w, "Method not allowed")
		}
		return
	}

	// Health endpoint: /proxy/network/api/s/{site}/stat/health
	if strings.Contains(path, "/stat/health") {
		// Extract site from path
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "s" && i+1 < len(parts) {
				site = parts[i+1]
				break
			}
		}
		s.handleHealth(w, r, site)
		return
	}

	// Sysinfo endpoint: /proxy/network/api/s/{site}/stat/sysinfo
	if strings.Contains(path, "/stat/sysinfo") {
		// Extract site from path
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "s" && i+1 < len(parts) {
				site = parts[i+1]
				break
			}
		}
		s.handleSysInfo(w, r, site)
		return
	}

	writeNotFound(w)
}

// handleListSites returns all sites.
func (s *Server) handleListSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	sites := s.state.ListSites()

	// Convert to interface slice
	data := make([]interface{}, len(sites))
	for i, site := range sites {
		data[i] = site
	}

	writeAPIResponse(w, data)
}

// handleHealth returns health information for a site.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Return mock health data
	health := []types.HealthData{
		{
			Subsystem: "www",
			Status:    "ok",
		},
		{
			Subsystem: "wan",
			Status:    "ok",
			NumGw:     1,
		},
		{
			Subsystem: "lan",
			Status:    "ok",
			NumSta:    5,
		},
	}

	// Convert to interface slice
	data := make([]interface{}, len(health))
	for i, h := range health {
		data[i] = h
	}

	writeAPIResponse(w, data)
}

// handleSysInfo returns system information.
func (s *Server) handleSysInfo(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	sysInfo := &types.SysInfo{
		Hostname:   "UDM-Pro",
		Version:    "7.5.174",
		HTTPSPort:  443,
		Console:    true,
		UpdateAvailable: false,
	}

	writeAPIResponse(w, sysInfo)
}

// handleCreateSite creates a new site.
func (s *Server) handleCreateSite(w http.ResponseWriter, r *http.Request) {
	var req types.CreateSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID from name/desc
	id := req.Name
	if id == "" {
		id = strings.ToLower(strings.ReplaceAll(req.Desc, " ", "-"))
	}

	// Create site
	site := &types.Site{
		ID:   id,
		Name: id,
		Desc: req.Desc,
	}

	s.state.AddSite(site)

	writeAPIResponse(w, site)
}
