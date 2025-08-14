package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"chatmix-backend/internal/service"

	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	chatService service.ChatService
	authService service.AuthService
	upgrader    websocket.Upgrader
	connections map[string]map[string]*websocket.Conn // connections maps roomCode -> username -> websocket connection
	connLock    sync.RWMutex
}

type ChatMessage struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

func NewChatHandler(chatService service.ChatService, authService service.AuthService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		authService: authService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		connections: make(map[string]map[string]*websocket.Conn),
	}
}

func (h *ChatHandler) HandleStartChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		WriteError(w, http.StatusBadRequest, "username required")
		return
	}

	response, err := h.chatService.StartChat(username)
	if err != nil {
		log.Printf("Error starting chat: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed to start chat")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) HandleQueueStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		WriteError(w, http.StatusBadRequest, "username required")
		return
	}

	position := h.chatService.GetQueuePosition(username)
	queueSize := h.chatService.GetQueueSize()

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"in_queue":   position > 0,
		"position":   position,
		"queue_size": queueSize,
	})
}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	username := r.URL.Query().Get("username")
	token := r.URL.Query().Get("token")

	if roomCode == "" || username == "" {
		WriteError(w, http.StatusBadRequest, "room and username required")
		return
	}

	if token == "" {
		WriteError(w, http.StatusUnauthorized, "authentication token required")
		return
	}

	// Verify room exists and user can join
	if err := h.chatService.JoinRoom(roomCode, username); err != nil {
		log.Printf("Error joining room: %v", err)
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Add connection
	h.addConnection(roomCode, username, conn)

	h.handleConnection(roomCode, username, conn)
}

func (h *ChatHandler) addConnection(roomCode, username string, conn *websocket.Conn) {
	h.connLock.Lock()
	defer h.connLock.Unlock()

	if h.connections[roomCode] == nil {
		h.connections[roomCode] = make(map[string]*websocket.Conn)
	}

	// Close existing connection if any
	if oldConn := h.connections[roomCode][username]; oldConn != nil {
		oldConn.Close()
	}

	h.connections[roomCode][username] = conn
}

func (h *ChatHandler) removeConnection(roomCode, username string) {
	h.connLock.Lock()
	defer h.connLock.Unlock()

	if roomConns := h.connections[roomCode]; roomConns != nil {
		delete(roomConns, username)
		if len(roomConns) == 0 {
			delete(h.connections, roomCode)
		}
	}

	// Remove user from room in service
	h.chatService.LeaveRoom(roomCode, username)
}

func (h *ChatHandler) handleConnection(roomCode, username string, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		h.removeConnection(roomCode, username)
	}()

	// Set connection limits
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping routine
	go h.pingRoutine(conn)

	// Send welcome message
	h.broadcastToRoom(roomCode, ChatMessage{
		Type:      "system",
		Text:      username + " đã vào phòng chat",
		Timestamp: time.Now().UnixMilli(),
	})

	// Message reading loop
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Broadcast message to room
		message := ChatMessage{
			Type:      "message",
			From:      username,
			Text:      string(messageBytes),
			Timestamp: time.Now().UnixMilli(),
		}

		h.broadcastToRoom(roomCode, message)
	}

	// Send leave message
	h.broadcastToRoom(roomCode, ChatMessage{
		Type:      "system",
		Text:      username + " đã rời khỏi phòng chat",
		Timestamp: time.Now().UnixMilli(),
	})
}

func (h *ChatHandler) broadcastToRoom(roomCode string, message ChatMessage) {
	h.connLock.RLock()
	roomConns := h.connections[roomCode]
	h.connLock.RUnlock()

	if roomConns == nil {
		return
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Send to all connections in room
	for username, conn := range roomConns {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("Error sending message to %s: %v", username, err)
			conn.Close()
			h.removeConnection(roomCode, username)
		}
	}
}

func (h *ChatHandler) pingRoutine(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
			return
		}
	}
}
