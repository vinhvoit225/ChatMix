package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	Logging   LoggingConfig   `yaml:"logging"`
	Auth      AuthConfig      `yaml:"auth"`
	Features  FeaturesConfig  `yaml:"features"`
	Chat      ChatConfig      `yaml:"chat"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	CORS         CORSConfig    `yaml:"cors"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

type DatabaseConfig struct {
	URI         string            `yaml:"uri"`
	Name        string            `yaml:"name"`
	Timeout     time.Duration     `yaml:"timeout"`
	Collections CollectionsConfig `yaml:"collections"`
}

type CollectionsConfig struct {
	Messages      string `yaml:"messages"`
	Users         string `yaml:"users"`
	RefreshTokens string `yaml:"refresh_tokens"`
	Sessions      string `yaml:"sessions"`
	Captchas      string `yaml:"captchas"`
}

type WebSocketConfig struct {
	ReadBufferSize  int  `yaml:"read_buffer_size"`
	WriteBufferSize int  `yaml:"write_buffer_size"`
	CheckOrigin     bool `yaml:"check_origin"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type AuthConfig struct {
	JWTSecret          string `yaml:"jwt_secret"`
	AccessTokenExpiry  int    `yaml:"access_token_expiry"`  // hours
	RefreshTokenExpiry int    `yaml:"refresh_token_expiry"` // hours
}

type FeaturesConfig struct {
	MaxUsernameLength int  `yaml:"max_username_length"`
	RequireAuth       bool `yaml:"require_auth"`
	CaptchaEnabled    bool `yaml:"captcha_enabled"`
}

type ChatConfig struct {
	MaxRooms            int           `yaml:"max_rooms"`
	QueueTimeout        time.Duration `yaml:"queue_timeout"`
	RoomCleanupInterval time.Duration `yaml:"room_cleanup_interval"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "configs/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if c.Database.URI == "" {
		return fmt.Errorf("database URI is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Features.MaxUsernameLength <= 0 {
		return fmt.Errorf("max username length must be positive")
	}

	if c.Chat.MaxRooms <= 0 {
		return fmt.Errorf("max rooms must be positive")
	}

	if c.Chat.QueueTimeout <= 0 {
		return fmt.Errorf("queue timeout must be positive")
	}

	if c.Chat.RoomCleanupInterval <= 0 {
		return fmt.Errorf("room cleanup interval must be positive")
	}

	return nil
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
