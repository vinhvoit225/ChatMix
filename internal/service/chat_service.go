package service

import (
	"chatmix-backend/internal/config"
	"chatmix-backend/internal/model"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Simple chat matching service
// Logic: User tries to join existing waiting room, or creates new room

type ChatService interface {
	StartChat(username string) (*model.ChatStartResponse, error)
	JoinRoom(roomCode, username string) error
	LeaveRoom(roomCode, username string)
	GetRoom(roomCode string) (*model.ChatRoom, bool)
	GetWaitingRooms() []*model.ChatRoom
	GetQueuePosition(username string) int
	GetQueueSize() int
}

type chatService struct {
	rooms     map[string]*model.ChatRoom
	roomsLock sync.RWMutex
	queue     []model.QueueEntry
	queueLock sync.RWMutex
	config    *config.ChatConfig
	logger    *logrus.Logger
}

func NewChatService(cfg *config.Config, logger *logrus.Logger) ChatService {
	cs := &chatService{
		rooms:  make(map[string]*model.ChatRoom),
		queue:  make([]model.QueueEntry, 0),
		config: &cfg.Chat,
		logger: logger,
	}

	// Start background queue processor
	go cs.processQueue()
	go cs.cleanupExpiredQueueEntries()
	go cs.cleanupLonelyRooms()

	return cs
}

// StartChat finds a waiting room and joins it, creates a new room, or adds to queue
func (s *chatService) StartChat(username string) (*model.ChatStartResponse, error) {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	// First, check if user is already in a room
	for _, room := range s.rooms {
		if room.HasUser(username) {
			return &model.ChatStartResponse{
				Status:   "room_assigned",
				RoomCode: room.Code,
				Message:  "Already in room",
			}, nil
		}
	}

	// Try to find a waiting room (exactly 1 user)
	for _, room := range s.rooms {
		if room.IsWaiting() {
			room.AddUser(username)
			return &model.ChatStartResponse{
				Status:   "room_assigned",
				RoomCode: room.Code,
				Message:  "Joined existing room",
			}, nil
		}
	}

	// Check if we can create a new room (under limit)
	if len(s.rooms) < s.config.MaxRooms {
		// Create new room
		code := s.generateRoomCode()
		room := &model.ChatRoom{
			Code:      code,
			Users:     []string{username},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.rooms[code] = room

		log.Println("List of rooms:")
		for code, room := range s.rooms {
			log.Printf("Room %s: %v", code, room)
		}
		return &model.ChatStartResponse{
			Status:   "room_assigned",
			RoomCode: room.Code,
			Message:  "Created new room",
		}, nil

	}

	// Room limit reached, add to queue
	return s.addToQueue(username)
}

// JoinRoom allows user to join specific room if space available
func (s *chatService) JoinRoom(roomCode, username string) error {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	// log list room
	log.Println("List of rooms:")
	for code, room := range s.rooms {
		log.Printf("Room %s: %v", code, room)
	}

	room, exists := s.rooms[roomCode]
	if !exists {
		return fmt.Errorf("room not found")
	}

	if room.HasUser(username) {
		return nil // already in room
	}

	if room.IsFull() {
		return fmt.Errorf("room is full")
	}

	room.AddUser(username)
	return nil
}

// LeaveRoom removes user from room, deletes room if empty
func (s *chatService) LeaveRoom(roomCode, username string) {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	room, exists := s.rooms[roomCode]
	if !exists {
		return
	}

	room.RemoveUser(username)

	// Delete room if empty
	if len(room.Users) == 0 {
		delete(s.rooms, roomCode)
	}
}

// GetRoom returns room by code
func (s *chatService) GetRoom(roomCode string) (*model.ChatRoom, bool) {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()

	room, exists := s.rooms[roomCode]
	if !exists {
		return nil, false
	}

	return s.cloneRoom(room), true
}

// GetWaitingRooms returns all rooms waiting for a second user
func (s *chatService) GetWaitingRooms() []*model.ChatRoom {
	s.roomsLock.RLock()
	defer s.roomsLock.RUnlock()

	var waitingRooms []*model.ChatRoom
	for _, room := range s.rooms {
		if room.IsWaiting() {
			waitingRooms = append(waitingRooms, s.cloneRoom(room))
		}
	}

	return waitingRooms
}

// GetQueuePosition returns user's position in queue (1-based), 0 if not in queue
func (s *chatService) GetQueuePosition(username string) int {
	s.queueLock.RLock()
	defer s.queueLock.RUnlock()

	for i, entry := range s.queue {
		if entry.Username == username {
			return i + 1 // 1-based position
		}
	}
	return 0
}

// GetQueueSize returns total number of users in queue
func (s *chatService) GetQueueSize() int {
	s.queueLock.RLock()
	defer s.queueLock.RUnlock()

	return len(s.queue)
}

// addToQueue adds user to queue and returns response
func (s *chatService) addToQueue(username string) (*model.ChatStartResponse, error) {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	// Check if user already in queue
	for i, entry := range s.queue {
		if entry.Username == username {
			return &model.ChatStartResponse{
				Status:   "queued",
				Position: i + 1,
				Message:  "Already in queue",
			}, nil
		}
	}

	// Add to queue
	s.queue = append(s.queue, model.QueueEntry{
		Username: username,
		QueuedAt: time.Now(),
	})

	return &model.ChatStartResponse{
		Status:   "queued",
		Position: len(s.queue),
		Message:  fmt.Sprintf("Added to queue. Position: %d", len(s.queue)),
	}, nil
}

// removeFromQueue removes user from queue
func (s *chatService) removeFromQueue(username string) {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	for i, entry := range s.queue {
		if entry.Username == username {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			break
		}
	}
}

// processQueue runs in background to assign rooms to queued users
func (s *chatService) processQueue() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for range ticker.C {
		s.tryAssignQueuedUsers()
	}
}

// tryAssignQueuedUsers tries to assign rooms to users in queue
func (s *chatService) tryAssignQueuedUsers() {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	if len(s.queue) == 0 {
		return
	}

	// Try to find available spots
	for i := 0; i < len(s.queue); i++ {
		user := s.queue[i]

		// Try to find a waiting room
		roomAssigned := false
		for _, room := range s.rooms {
			if room.IsWaiting() {
				room.AddUser(user.Username)
				roomAssigned = true
				break
			}
		}

		// If no waiting room and we can create new room
		if !roomAssigned && len(s.rooms) < s.config.MaxRooms {
			code := s.generateRoomCode()
			room := &model.ChatRoom{
				Code:      code,
				Users:     []string{user.Username},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			s.rooms[code] = room
			roomAssigned = true
		}

		// Remove from queue if assigned
		if roomAssigned {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			i-- // Adjust index after removal
		}
	}
}

// cleanupExpiredQueueEntries removes users who have been in queue too long
func (s *chatService) cleanupExpiredQueueEntries() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		s.queueLock.Lock()
		now := time.Now()

		var validEntries []model.QueueEntry
		for _, entry := range s.queue {
			if now.Sub(entry.QueuedAt) < s.config.QueueTimeout {
				validEntries = append(validEntries, entry)
			}
		}

		s.queue = validEntries
		s.queueLock.Unlock()
	}
}

// cleanupLonelyRooms removes rooms where a single user has been waiting too long
func (s *chatService) cleanupLonelyRooms() {
	log.Println("Cleaning up lonely rooms...")
	ticker := time.NewTicker(s.config.RoomCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.roomsLock.Lock()
		now := time.Now()

		var roomsToDelete []string
		for code, room := range s.rooms {
			// Check if room has exactly 1 user and has been waiting longer than cleanup interval
			if len(room.Users) == 1 && now.Sub(room.UpdatedAt) >= s.config.RoomCleanupInterval {
				log.Printf("Room %s is lonely and will be deleted", code)
				log.Printf("Room %s was created at %s", code, room.CreatedAt)
				log.Printf("Room %s was updated at %s", code, room.UpdatedAt)
				log.Printf("RoomCleanupInterval: %s", s.config.RoomCleanupInterval)
				roomsToDelete = append(roomsToDelete, code)
			}
		}

		// Delete the lonely rooms
		for _, code := range roomsToDelete {
			delete(s.rooms, code)
		}

		s.roomsLock.Unlock()
	}
}

// Helper methods

func (s *chatService) cloneRoom(room *model.ChatRoom) *model.ChatRoom {
	clone := &model.ChatRoom{
		Code:      room.Code,
		CreatedAt: room.CreatedAt,
		UpdatedAt: room.UpdatedAt,
		Users:     make([]string, len(room.Users)),
	}
	copy(clone.Users, room.Users)
	return clone
}

func (s *chatService) generateRoomCode() string {
	for {
		code := generateRandomCode(8)
		if _, exists := s.rooms[code]; !exists {
			return code
		}
	}
}

func generateRandomCode(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	encoded := base32.StdEncoding.EncodeToString(bytes)
	code := strings.ToUpper(strings.TrimRight(encoded, "="))
	if len(code) >= n {
		return code[:n]
	}
	return code
}
