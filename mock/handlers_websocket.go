package mock

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/unifi-go/gofi/types"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsConnection struct {
	conn *websocket.Conn
	site string
}

// handleWebSocket handles WebSocket connections for event streaming.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract site from path
	site := extractSiteFromPath(r.URL.Path)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	wsConn := &wsConnection{
		conn: conn,
		site: site,
	}

	s.addWebSocketConnection(wsConn)
	defer s.removeWebSocketConnection(wsConn)

	// Keep connection alive and handle messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

var (
	wsConnections   = make(map[*wsConnection]bool)
	wsConnectionsMu sync.RWMutex
)

func (s *Server) addWebSocketConnection(conn *wsConnection) {
	wsConnectionsMu.Lock()
	defer wsConnectionsMu.Unlock()
	wsConnections[conn] = true
}

func (s *Server) removeWebSocketConnection(conn *wsConnection) {
	wsConnectionsMu.Lock()
	defer wsConnectionsMu.Unlock()
	delete(wsConnections, conn)
}

// BroadcastEvent broadcasts an event to all connected WebSocket clients.
func (s *Server) BroadcastEvent(event *types.Event) {
	wsConnectionsMu.RLock()
	connections := make([]*wsConnection, 0, len(wsConnections))
	for conn := range wsConnections {
		connections = append(connections, conn)
	}
	wsConnectionsMu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	for _, conn := range connections {
		// Only send to connections for the same site
		if conn.site == "" || conn.site == event.SiteID {
			_ = conn.conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// SimulateClientConnect simulates a client connection event.
func (s *Server) SimulateClientConnect(site string, client *types.Client) {
	event := &types.Event{
		Key:     "EVT_WU_Connected",
		SiteID:  site,
		Time:    0,
		Message: "User connected",
	}
	s.BroadcastEvent(event)
}

// SimulateClientDisconnect simulates a client disconnection event.
func (s *Server) SimulateClientDisconnect(site, mac string) {
	event := &types.Event{
		Key:     "EVT_WU_Disconnected",
		SiteID:  site,
		Time:    0,
		Message: "User disconnected",
	}
	s.BroadcastEvent(event)
}

// SimulateDeviceUpdate simulates a device update event.
func (s *Server) SimulateDeviceUpdate(site string, device *types.Device) {
	event := &types.Event{
		Key:     "EVT_AP_Updated",
		SiteID:  site,
		Time:    0,
		Message: "Device updated",
	}
	s.BroadcastEvent(event)
}

// SimulateAlarm simulates an alarm event.
func (s *Server) SimulateAlarm(site string, alarm *types.Alarm) {
	event := &types.Event{
		Key:     "EVT_AD_Alarm",
		SiteID:  site,
		Time:    0,
		Message: "Alarm triggered",
	}
	s.BroadcastEvent(event)
}

func extractSiteFromPath(path string) string {
	// Extract site from /proxy/network/wss/s/{site}/events
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "default"
}
