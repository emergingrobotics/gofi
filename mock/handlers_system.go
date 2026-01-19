package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/unifi-go/gofi/types"
)

// handleSystem routes system-related requests.
func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Reboot endpoint: /api/cmd/system
	if strings.Contains(path, "/api/cmd/system") {
		s.handleReboot(w, r)
		return
	}

	// Backup endpoints: /api/cmd/backup
	if strings.Contains(path, "/api/cmd/backup") {
		s.handleBackups(w, r)
		return
	}

	// Admin list: /api/stat/admin
	if strings.Contains(path, "/api/stat/admin") {
		s.handleAdminList(w, r)
		return
	}

	// Speed test endpoints (site-specific)
	if strings.Contains(path, "/cmd/speedtest") {
		s.handleSpeedTest(w, r, site)
		return
	}

	if strings.Contains(path, "/stat/speedtest") {
		s.handleSpeedTestStatus(w, r, site)
		return
	}

	writeNotFound(w)
}

// handleReboot handles system reboot command.
func (s *Server) handleReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeNotFound(w)
		return
	}

	// Check CSRF token even though we might have disabled it for testing
	// This is for realistic simulation
	if s.requireCSRF {
		if !s.validateCSRF(r) {
			writeForbidden(w, "Invalid CSRF token")
			return
		}
	}

	// Parse request body to check for cmd=reboot
	var req struct {
		Cmd string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	if req.Cmd != "reboot" {
		writeBadRequest(w, "Invalid command")
		return
	}

	// Simulate reboot success
	writeAPIResponse(w, []interface{}{})
}

// handleBackups routes backup-related requests.
func (s *Server) handleBackups(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var filename string
	for i, part := range parts {
		if part == "backup" && i+1 < len(parts) && parts[i+1] != "" {
			filename = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		s.handleBackupList(w, r)
	case "POST":
		s.handleBackupCreate(w, r)
	case "DELETE":
		if filename != "" {
			s.handleBackupDelete(w, r, filename)
		} else {
			writeBadRequest(w, "Filename required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleBackupList returns all backups.
func (s *Server) handleBackupList(w http.ResponseWriter, r *http.Request) {
	backups := s.state.ListBackups()

	data := make([]interface{}, len(backups))
	for i, backup := range backups {
		data[i] = *backup
	}

	writeAPIResponse(w, data)
}

// handleBackupCreate creates a new backup.
func (s *Server) handleBackupCreate(w http.ResponseWriter, r *http.Request) {
	// Generate backup filename
	now := time.Now()
	filename := fmt.Sprintf("backup_%d_%d%02d%02d_%02d%02d.unf",
		now.Unix(), now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

	backup := &types.Backup{
		Filename: filename,
		Size:     1024 * 1024, // 1MB mock size
		Time:     now.Unix(),
		Datetime: now.Format(time.RFC3339),
	}

	s.state.AddBackup(backup)
	writeAPIResponse(w, []interface{}{*backup})
}

// handleBackupDelete deletes a backup.
func (s *Server) handleBackupDelete(w http.ResponseWriter, r *http.Request, filename string) {
	s.state.DeleteBackup(filename)
	writeAPIResponse(w, []interface{}{})
}

// handleAdminList returns all admin users.
func (s *Server) handleAdminList(w http.ResponseWriter, r *http.Request) {
	admins := s.state.ListAdmins()

	data := make([]interface{}, len(admins))
	for i, admin := range admins {
		data[i] = *admin
	}

	writeAPIResponse(w, data)
}

// handleSpeedTest initiates a speed test.
func (s *Server) handleSpeedTest(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeNotFound(w)
		return
	}

	// Simulate speed test (immediately complete for testing)
	s.state.SimulateSpeedTest()

	writeAPIResponse(w, []interface{}{})
}

// handleSpeedTestStatus returns the speed test status.
func (s *Server) handleSpeedTestStatus(w http.ResponseWriter, r *http.Request, site string) {
	status := s.state.GetSpeedTestStatus()
	if status == nil {
		// No speed test run yet
		writeAPIResponse(w, []interface{}{})
		return
	}

	writeAPIResponse(w, []interface{}{*status})
}
