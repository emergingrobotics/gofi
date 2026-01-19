package mock

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/unifi-go/gofi/types"
)

// handleClients routes client/station requests.
func (s *Server) handleClients(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Client stat endpoints
	if strings.Contains(path, "/stat/sta") {
		s.handleClientStat(w, r, site)
		return
	}

	if strings.Contains(path, "/stat/alluser") {
		s.handleAllUserStat(w, r, site)
		return
	}

	// Client commands
	if strings.Contains(path, "/cmd/stamgr") {
		s.handleClientCommand(w, r, site)
		return
	}

	writeNotFound(w)
}

// handleClientStat returns active clients.
func (s *Server) handleClientStat(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	clients := s.state.ListClients()

	// Filter active clients (seen in last 5 minutes)
	now := time.Now().Unix()
	activeClients := make([]interface{}, 0)
	for _, client := range clients {
		if client.LastSeen > 0 && now-client.LastSeen < 300 {
			activeClients = append(activeClients, *client)
		}
	}

	writeAPIResponse(w, activeClients)
}

// handleAllUserStat returns all clients with history.
func (s *Server) handleAllUserStat(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Parse within_hours parameter
	withinHours := 8760 // Default: 1 year
	if hours := r.URL.Query().Get("within"); hours != "" {
		if h, err := strconv.Atoi(hours); err == nil {
			withinHours = h
		}
	}

	clients := s.state.ListClients()

	// Filter by time window
	cutoff := time.Now().Unix() - int64(withinHours*3600)
	filteredClients := make([]interface{}, 0)
	for _, client := range clients {
		if client.LastSeen >= cutoff {
			filteredClients = append(filteredClients, *client)
		}
	}

	writeAPIResponse(w, filteredClients)
}

// handleClientCommand processes client management commands.
func (s *Server) handleClientCommand(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	// Parse command
	var cmd struct {
		CMD  string `json:"cmd"`
		MAC  string `json:"mac"`
		// Guest authorization options
		Minutes int    `json:"minutes,omitempty"`
		Up      int    `json:"up,omitempty"`
		Down    int    `json:"down,omitempty"`
		Bytes   int    `json:"bytes,omitempty"`
		APMAC   string `json:"ap_mac,omitempty"`
		// Device fingerprint
		DevID   int    `json:"dev_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	if cmd.MAC == "" {
		writeBadRequest(w, "MAC address required")
		return
	}

	// Get or create client
	client := s.state.GetClient(cmd.MAC)
	if client == nil {
		// For some commands, client must exist
		if cmd.CMD != "authorize-guest" {
			writeNotFound(w)
			return
		}
		// Create guest client
		client = &types.Client{
			MAC:       cmd.MAC,
			IsGuest:   true,
			FirstSeen: time.Now().Unix(),
			LastSeen:  time.Now().Unix(),
		}
		s.state.AddClient(client)
	}

	// Execute command
	switch cmd.CMD {
	case "block-sta":
		client.Blocked = true
		s.state.UpdateClient(client)
	case "unblock-sta":
		client.Blocked = false
		s.state.UpdateClient(client)
	case "kick-sta":
		client.GuestKicked = true
		// In real controller, this would disconnect the client
		s.state.UpdateClient(client)
	case "forget-sta":
		s.state.DeleteClient(cmd.MAC)
	case "authorize-guest":
		client.GuestAuthorized = true
		client.Authorized = true
		if cmd.Minutes > 0 {
			// Set expiration (not fully modeled in mock)
			client.LastSeen = time.Now().Unix()
		}
		s.state.UpdateClient(client)
	case "unauthorize-guest":
		client.GuestAuthorized = false
		client.Authorized = false
		s.state.UpdateClient(client)
	case "set-sta-dev-id":
		if cmd.DevID > 0 {
			client.DeviceIDOverride = cmd.DevID
			s.state.UpdateClient(client)
		}
	default:
		writeBadRequest(w, "Unknown command: "+cmd.CMD)
		return
	}

	writeAPIResponse(w, []interface{}{})
}
