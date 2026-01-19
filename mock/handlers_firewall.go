package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleFirewall routes firewall-related requests.
func (s *Server) handleFirewall(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Firewall Group endpoints: /rest/firewallgroup
	if strings.Contains(path, "/rest/firewallgroup") {
		s.handleFirewallGroups(w, r, site)
		return
	}

	// Firewall Rule endpoints: /rest/firewallrule
	if strings.Contains(path, "/rest/firewallrule") {
		// Check if this is a reorder request
		if strings.Contains(path, "/reorder") && r.Method == "POST" {
			s.handleFirewallReorder(w, r, site)
			return
		}

		// Extract ID if present
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "firewallrule" && i+1 < len(parts) && parts[i+1] != "" && parts[i+1] != "reorder" {
				id = parts[i+1]
				break
			}
		}

		switch r.Method {
		case "GET":
			if id != "" {
				s.handleGetFirewallRule(w, r, site, id)
			} else {
				s.handleListFirewallRules(w, r, site)
			}
		case "POST":
			s.handleCreateFirewallRule(w, r, site)
		case "PUT":
			if id != "" {
				s.handleUpdateFirewallRule(w, r, site, id)
			} else {
				writeBadRequest(w, "Firewall rule ID required for update")
			}
		case "DELETE":
			if id != "" {
				s.handleDeleteFirewallRule(w, r, site, id)
			} else {
				writeBadRequest(w, "Firewall rule ID required for delete")
			}
		default:
			writeNotFound(w)
		}
		return
	}

	writeNotFound(w)
}

// handleListFirewallRules returns all firewall rules for a site.
func (s *Server) handleListFirewallRules(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	rules := s.state.ListFirewallRules()

	data := make([]interface{}, len(rules))
	for i, rule := range rules {
		data[i] = *rule
	}

	writeAPIResponse(w, data)
}

// handleGetFirewallRule returns a specific firewall rule by ID.
func (s *Server) handleGetFirewallRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "GET" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	rule := s.state.GetFirewallRule(id)
	if rule == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*rule})
}

// handleCreateFirewallRule creates a new firewall rule.
func (s *Server) handleCreateFirewallRule(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	var rule types.FirewallRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Validate required fields
	if rule.Name == "" {
		writeBadRequest(w, "Firewall rule name is required")
		return
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = generateID()
	}
	rule.SiteID = site

	s.state.AddFirewallRule(&rule)

	writeAPIResponse(w, []interface{}{rule})
}

// handleUpdateFirewallRule updates an existing firewall rule.
func (s *Server) handleUpdateFirewallRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "PUT" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetFirewallRule(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var rule types.FirewallRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Preserve ID and site
	rule.ID = id
	rule.SiteID = site

	s.state.UpdateFirewallRule(&rule)

	writeAPIResponse(w, []interface{}{rule})
}

// handleDeleteFirewallRule deletes a firewall rule.
func (s *Server) handleDeleteFirewallRule(w http.ResponseWriter, r *http.Request, site, id string) {
	if r.Method != "DELETE" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	existing := s.state.GetFirewallRule(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteFirewallRule(id)

	writeAPIResponse(w, []interface{}{})
}

// handleFirewallReorder handles reordering firewall rules.
func (s *Server) handleFirewallReorder(w http.ResponseWriter, r *http.Request, site string) {
	if r.Method != "POST" {
		writeBadRequest(w, "Method not allowed")
		return
	}

	var updates []types.FirewallRuleIndexUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Update rule indices
	for _, update := range updates {
		rule := s.state.GetFirewallRule(update.ID)
		if rule != nil {
			rule.RuleIndex = update.RuleIndex
			s.state.UpdateFirewallRule(rule)
		}
	}

	writeAPIResponse(w, []interface{}{})
}

// handleFirewallGroups routes firewall group requests.
func (s *Server) handleFirewallGroups(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Extract ID if present
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "firewallgroup" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetFirewallGroup(w, r, site, id)
		} else {
			s.handleListFirewallGroups(w, r, site)
		}
	case "POST":
		s.handleCreateFirewallGroup(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdateFirewallGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "Firewall group ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeleteFirewallGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "Firewall group ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListFirewallGroups returns all firewall groups for a site.
func (s *Server) handleListFirewallGroups(w http.ResponseWriter, r *http.Request, site string) {
	groups := s.state.ListFirewallGroups()

	data := make([]interface{}, len(groups))
	for i, group := range groups {
		data[i] = *group
	}

	writeAPIResponse(w, data)
}

// handleGetFirewallGroup returns a specific firewall group by ID.
func (s *Server) handleGetFirewallGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	group := s.state.GetFirewallGroup(id)
	if group == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*group})
}

// handleCreateFirewallGroup creates a new firewall group.
func (s *Server) handleCreateFirewallGroup(w http.ResponseWriter, r *http.Request, site string) {
	var group types.FirewallGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Validate required fields
	if group.Name == "" {
		writeBadRequest(w, "Firewall group name is required")
		return
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = generateID()
	}
	group.SiteID = site

	s.state.AddFirewallGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleUpdateFirewallGroup updates an existing firewall group.
func (s *Server) handleUpdateFirewallGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetFirewallGroup(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var group types.FirewallGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid JSON")
		return
	}

	// Preserve ID and site
	group.ID = id
	group.SiteID = site

	s.state.UpdateFirewallGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleDeleteFirewallGroup deletes a firewall group.
func (s *Server) handleDeleteFirewallGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetFirewallGroup(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteFirewallGroup(id)

	writeAPIResponse(w, []interface{}{})
}
