package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/src/types"
)

// handleDNSRecords routes local (static) DNS record requests (v2 API).
//
// The v2 surface returns bare JSON rather than the meta/data envelope the v1
// endpoints use, so these handlers write with writeJSON instead of
// writeAPIResponse. DNSService unmarshals responses directly into records and
// would silently decode an envelope into a zero value.
func (s *Server) handleDNSRecords(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Extract ID if present: /v2/api/site/{site}/static-dns/{id}
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "static-dns" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			// Real controllers answer GET on an individual static-dns record
			// with 405; only the collection is readable. Mirror that so code
			// cannot pass here and fail against hardware.
			writeAPIError(w, http.StatusMethodNotAllowed, "error", "Method Not Allowed")
			return
		}
		s.handleListDNSRecords(w, r, site)
	case "POST":
		s.handleCreateDNSRecord(w, r, site)
	case "PUT":
		if id == "" {
			writeBadRequest(w, "DNS record ID required for update")
			return
		}
		s.handleUpdateDNSRecord(w, r, site, id)
	case "DELETE":
		if id == "" {
			writeBadRequest(w, "DNS record ID required for delete")
			return
		}
		s.handleDeleteDNSRecord(w, r, site, id)
	default:
		writeNotFound(w)
	}
}

// handleListDNSRecords returns all local DNS records as a bare array.
func (s *Server) handleListDNSRecords(w http.ResponseWriter, r *http.Request, site string) {
	records := s.state.ListDNSRecords()

	data := make([]types.DNSRecord, len(records))
	for i, record := range records {
		data[i] = *record
	}

	writeJSON(w, http.StatusOK, data)
}

// handleCreateDNSRecord creates a new local DNS record.
func (s *Server) handleCreateDNSRecord(w http.ResponseWriter, r *http.Request, site string) {
	var record types.DNSRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	if record.Key == "" {
		writeBadRequest(w, "DNS record key is required")
		return
	}
	if record.Value == "" {
		writeBadRequest(w, "DNS record value is required")
		return
	}

	if record.ID == "" {
		record.ID = generateID()
	}
	if record.RecordType == "" {
		record.RecordType = types.DNSRecordTypeA
	}

	s.state.AddDNSRecord(&record)

	writeJSON(w, http.StatusOK, record)
}

// handleUpdateDNSRecord updates an existing local DNS record.
func (s *Server) handleUpdateDNSRecord(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetDNSRecord(id) == nil {
		writeNotFound(w)
		return
	}

	var record types.DNSRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	record.ID = id
	s.state.UpdateDNSRecord(&record)

	writeJSON(w, http.StatusOK, record)
}

// handleDeleteDNSRecord deletes a local DNS record.
func (s *Server) handleDeleteDNSRecord(w http.ResponseWriter, r *http.Request, site, id string) {
	if s.state.GetDNSRecord(id) == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteDNSRecord(id)

	writeJSON(w, http.StatusOK, []types.DNSRecord{})
}
