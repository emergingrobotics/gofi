package mock

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/unifi-go/gofi/types"
)

// handleUsers routes user-related requests.
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// User group endpoints
	if strings.Contains(path, "/rest/usergroup") {
		s.handleUserGroups(w, r, site)
		return
	}

	// User endpoints: /rest/user
	if strings.Contains(path, "/rest/user") {
		parts := strings.Split(path, "/")
		var id string
		for i, part := range parts {
			if part == "user" && i+1 < len(parts) && parts[i+1] != "" {
				id = parts[i+1]
				break
			}
		}

		switch r.Method {
		case "GET":
			if id != "" {
				s.handleGetUser(w, r, site, id)
			} else {
				s.handleListUsers(w, r, site)
			}
		case "POST":
			s.handleCreateUser(w, r, site)
		case "PUT":
			if id != "" {
				s.handleUpdateUser(w, r, site, id)
			} else {
				writeBadRequest(w, "User ID required for update")
			}
		case "DELETE":
			if id != "" {
				s.handleDeleteUser(w, r, site, id)
			} else {
				writeBadRequest(w, "User ID required for delete")
			}
		default:
			writeNotFound(w)
		}
		return
	}

	writeNotFound(w)
}

// handleListUsers returns all users.
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request, site string) {
	users := s.state.ListKnownClients()

	data := make([]interface{}, len(users))
	for i, user := range users {
		data[i] = *user
	}

	writeAPIResponse(w, data)
}

// handleGetUser returns a specific user by ID.
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request, site, id string) {
	user := s.state.GetKnownClient(id)
	if user == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*user})
}

// handleCreateUser creates a new user.
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request, site string) {
	var user types.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = generateID()
	}

	// Set site ID
	if user.SiteID == "" {
		user.SiteID = site
	}

	s.state.AddKnownClient(&user)

	writeAPIResponse(w, []interface{}{user})
}

// handleUpdateUser updates an existing user.
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetKnownClient(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var user types.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID
	user.ID = id
	user.SiteID = site

	s.state.UpdateKnownClient(&user)

	writeAPIResponse(w, []interface{}{user})
}

// handleDeleteUser deletes a user.
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request, site, id string) {
	user := s.state.GetKnownClient(id)
	if user == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteKnownClient(id)

	writeAPIResponse(w, []interface{}{})
}

// handleUserGroups routes user group requests.
func (s *Server) handleUserGroups(w http.ResponseWriter, r *http.Request, site string) {
	path := r.URL.Path

	// Extract ID if present
	parts := strings.Split(path, "/")
	var id string
	for i, part := range parts {
		if part == "usergroup" && i+1 < len(parts) && parts[i+1] != "" {
			id = parts[i+1]
			break
		}
	}

	switch r.Method {
	case "GET":
		if id != "" {
			s.handleGetUserGroup(w, r, site, id)
		} else {
			s.handleListUserGroups(w, r, site)
		}
	case "POST":
		s.handleCreateUserGroup(w, r, site)
	case "PUT":
		if id != "" {
			s.handleUpdateUserGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "User group ID required for update")
		}
	case "DELETE":
		if id != "" {
			s.handleDeleteUserGroup(w, r, site, id)
		} else {
			writeBadRequest(w, "User group ID required for delete")
		}
	default:
		writeNotFound(w)
	}
}

// handleListUserGroups returns all user groups.
func (s *Server) handleListUserGroups(w http.ResponseWriter, r *http.Request, site string) {
	groups := s.state.ListUserGroups()

	data := make([]interface{}, len(groups))
	for i, group := range groups {
		data[i] = *group
	}

	writeAPIResponse(w, data)
}

// handleGetUserGroup returns a specific user group by ID.
func (s *Server) handleGetUserGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	group := s.state.GetUserGroup(id)
	if group == nil {
		writeNotFound(w)
		return
	}

	writeAPIResponse(w, []interface{}{*group})
}

// handleCreateUserGroup creates a new user group.
func (s *Server) handleCreateUserGroup(w http.ResponseWriter, r *http.Request, site string) {
	var group types.UserGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = generateID()
	}

	// Set site ID
	if group.SiteID == "" {
		group.SiteID = site
	}

	s.state.AddUserGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleUpdateUserGroup updates an existing user group.
func (s *Server) handleUpdateUserGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	existing := s.state.GetUserGroup(id)
	if existing == nil {
		writeNotFound(w)
		return
	}

	var group types.UserGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		writeBadRequest(w, "Invalid request body")
		return
	}

	// Preserve ID
	group.ID = id
	group.SiteID = site

	s.state.UpdateUserGroup(&group)

	writeAPIResponse(w, []interface{}{group})
}

// handleDeleteUserGroup deletes a user group.
func (s *Server) handleDeleteUserGroup(w http.ResponseWriter, r *http.Request, site, id string) {
	group := s.state.GetUserGroup(id)
	if group == nil {
		writeNotFound(w)
		return
	}

	s.state.DeleteUserGroup(id)

	writeAPIResponse(w, []interface{}{})
}
