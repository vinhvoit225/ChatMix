package model

import "time"

type ChatStartResponse struct {
	Status   string `json:"status"` // "room_assigned", "queued"
	RoomCode string `json:"room,omitempty"`
	Position int    `json:"position,omitempty"` // position in queue
	Message  string `json:"message,omitempty"`
}

type QueueEntry struct {
	Username string
	QueuedAt time.Time
}

type ChatRoom struct {
	Code      string
	Users     []string // max 2 users
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *ChatRoom) IsFull() bool {
	return len(r.Users) >= 2
}

func (r *ChatRoom) HasUser(username string) bool {
	for _, user := range r.Users {
		if user == username {
			return true
		}
	}
	return false
}

func (r *ChatRoom) IsWaiting() bool {
	return len(r.Users) == 1
}

func (r *ChatRoom) AddUser(username string) {
	if !r.HasUser(username) && !r.IsFull() {
		r.Users = append(r.Users, username)
		r.UpdatedAt = time.Now()
	}
}

func (r *ChatRoom) RemoveUser(username string) {
	for i, user := range r.Users {
		if user == username {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			r.UpdatedAt = time.Now()
			break
		}
	}
}
