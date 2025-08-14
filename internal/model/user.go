package model

import (
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderOther   Gender = "other"
	GenderPrivate Gender = "private"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username"`
	Email        string             `json:"email" bson:"email"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	Age          int                `json:"age,omitempty" bson:"age,omitempty"`
	Gender       Gender             `json:"gender,omitempty" bson:"gender,omitempty"`
	Bio          string             `json:"bio,omitempty" bson:"bio,omitempty"`
	IsOnline     bool               `json:"is_online" bson:"is_online"`
	IsVerified   bool               `json:"is_verified" bson:"is_verified"`
	LastSeen     time.Time          `json:"last_seen" bson:"last_seen"`
	JoinedAt     time.Time          `json:"joined_at" bson:"joined_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
	RoomID       string             `json:"room_id,omitempty" bson:"room_id,omitempty"`
}

type OnlineUser struct {
	*User
	Conn   *websocket.Conn `json:"-"`
	SendCh chan []byte     `json:"-"`
}

// NewUser creates a new user
func NewUser(username, email string) *User {
	now := time.Now()
	return &User{
		ID:        primitive.NewObjectID(),
		Username:  username,
		Email:     email,
		IsOnline:  false,
		LastSeen:  now,
		JoinedAt:  now,
		UpdatedAt: now,
	}
}

func NewUserWithProfile(username, email string, age int, gender Gender, bio string) *User {
	user := NewUser(username, email)
	user.Age = age
	user.Gender = gender
	user.Bio = bio
	return user
}

func NewOnlineUser(user *User, conn *websocket.Conn) *OnlineUser {
	return &OnlineUser{
		User:   user,
		Conn:   conn,
		SendCh: make(chan []byte, 256),
	}
}

func (u *User) IsValid(maxUsernameLength int) bool {
	if len(u.Username) == 0 || len(u.Username) > maxUsernameLength {
		return false
	}
	if len(u.Email) == 0 || !isValidEmail(u.Email) {
		return false
	}
	if u.Age < 0 || u.Age > 150 {
		return false
	}
	if len(u.Bio) > 500 {
		return false
	}
	return true
}

func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 320 {
		return false
	}
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false
			}
			atIndex = i
		}
	}
	return atIndex > 0 && atIndex < len(email)-1
}

func (u *User) UpdateLastSeen() {
	u.LastSeen = time.Now()
}

func (u *User) SetOnline(online bool) {
	u.IsOnline = online
	if online {
		u.UpdateLastSeen()
	}
}

func (u *User) ToPublicUser() map[string]interface{} {
	public := map[string]interface{}{
		"id":          u.ID,
		"username":    u.Username,
		// nickname field removed - using username only
		"is_online":   u.IsOnline,
		"is_verified": u.IsVerified,
		"last_seen":   u.LastSeen,
		"joined_at":   u.JoinedAt,
	}

	if u.Age > 0 {
		public["age"] = u.Age
	}
	if u.Gender != "" && u.Gender != GenderPrivate {
		public["gender"] = u.Gender
	}
	if u.Bio != "" {
		public["bio"] = u.Bio
	}

	return public
}

func (u *User) ToPrivateUser() map[string]interface{} {
	private := u.ToPublicUser()
	private["email"] = u.Email
	if u.Gender == GenderPrivate {
		private["gender"] = u.Gender
	}
	return private
}
