# ChatMix Backend

Ứng dụng chat với người lạ , không lưu trữ tin nhắn người dùng

## 🚀 Tính năng

- **Clean Architecture**: Phân tách rõ ràng các layer
- **WebSocket Real-time**: Chat thời gian thực
- **MongoDB**: Lưu trữ tin nhắn và người dùng
- **Configuration**: Đọc config từ file YAML
- **Logging**: Structured logging với Logrus
- **CORS**: Cross-origin resource sharing
- **Graceful Shutdown**: Tắt server một cách an toàn
- **Middleware**: Recovery, logging, CORS

## 📋 Yêu cầu

- Go 1.21+
- MongoDB 4.4+

## ⚙️ Cài đặt

### 1. Clone repository
```bash
git clone <repository-url>
cd ChatMix/backend
```

### 2. Cài đặt dependencies
```bash
make deps
# hoặc
go mod tidy
```

### 3. Cấu hình
```bash
# Tạo file config mặc định
make config

# Hoặc copy và chỉnh sửa
cp configs/config.yaml.example configs/config.yaml
```

### 4. Khởi động MongoDB
```bash
# Sử dụng Docker
make mongo-up

# Hoặc khởi động MongoDB locally
mongod
```

## 🎯 Chạy ứng dụng

### Development
```bash
# Chạy với live reload
make dev

# Hoặc chạy bình thường
make run
```

### Production
```bash
# Build
make build-prod

# Chạy
./bin/chatmix-backend
```

## 🔧 Makefile Commands

```bash
make deps          # Cài đặt dependencies
make build         # Build ứng dụng
make run           # Chạy ứng dụng
make dev           # Chạy với live reload
make test          # Chạy tests
make clean         # Dọn dẹp build artifacts
make lint          # Lint code
make fmt           # Format code
make docker-build  # Build Docker image
make mongo-up      # Khởi động MongoDB
make help          # Hiển thị help
```

## 📡 API Endpoints

### WebSocket
- `GET /ws?username=<username>` - Kết nối WebSocket chat
- `GET /ws/room?username=<username>&room=<room>` - Chat với room

### REST API

#### Messages
- `GET /api/messages` - Lấy tin nhắn gần đây
- `POST /api/messages` - Gửi tin nhắn (testing)
- `GET /api/messages/user/{username}` - Lấy tin nhắn theo user
- `DELETE /api/messages/{messageId}` - Xóa tin nhắn

#### Users
- `GET /api/users` - Lấy tất cả users
- `GET /api/users/online` - Lấy users online
- `GET /api/users/{username}` - Lấy thông tin user

#### System
- `GET /api/stats` - Thống kê hệ thống
- `GET /api/health` - Health check
- `GET /health` - Health check
- `GET /` - API documentation

## 🔐 Configuration

File `configs/config.yaml`:

```yaml
server:
  host: "localhost"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  cors:
    allowed_origins: 
      - "http://localhost:3000"

database:
  uri: "mongodb://localhost:27017"
  name: "chatmix"
  timeout: 10s

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  check_origin: true

logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json, text

features:
  max_message_length: 1000
  max_username_length: 50
  message_history_limit: 50
```

## 📝 Environment Variables

```bash
CONFIG_PATH=/path/to/config.yaml  # Path to config file
```

## 🧪 Testing

```bash
# Chạy tests
make test

# Test với coverage
make test-coverage

# Chạy benchmarks
make bench
```

## 📊 Logging

Server sử dụng structured logging với các level:
- `debug`: Chi tiết debug
- `info`: Thông tin chung
- `warn`: Cảnh báo
- `error`: Lỗi

## 🔒 Security

- Input validation
- CORS configuration
- Rate limiting (planned)
- Authentication (planned)

## 🐳 Docker

```bash
# Build image
make docker-build

# Run container
make docker-run

# Stop container
make docker-stop
```

## 📈 Performance

- Connection pooling với MongoDB
- Efficient WebSocket handling
- Graceful shutdown
- Memory optimization

## 🔧 Development Tools

```bash
# Cài đặt tools
make install-tools

# Format code
make fmt

# Lint code
make lint

# Security check
make security
```

## 🚀 Deployment

### Systemd Service

Tạo file `/etc/systemd/system/chatmix.service`:

```ini
[Unit]
Description=ChatMix Backend Server
After=network.target

[Service]
Type=simple
User=chatmix
WorkingDirectory=/opt/chatmix
ExecStart=/opt/chatmix/bin/chatmix-backend
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable chatmix
sudo systemctl start chatmix
```

## 📚 Documentation

- API documentation: `GET /`
- Code documentation: `make docs`

## 🤝 Contributing

1. Fork repository
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📄 License

MIT License
