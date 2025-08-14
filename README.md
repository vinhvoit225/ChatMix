# ChatMix Backend

Ứng dụng chat với người lạ , không lưu trữ tin nhắn người dùng

## 🚀 Tính năng

- **Clean Architecture**: Phân tách rõ ràng các layer
- **WebSocket Real-time**: Chat thời gian thực
- **MongoDB**: Lưu trữ thông tin người dùng
- **Configuration**: Đọc config từ file YAML
- **Logging**: Structured logging với Logrus
- **CORS**: Cross-origin resource sharing
- **Middleware**: Recovery, logging, CORS
- **Docker**: Tạo container với Docker

## 🚢 Triển khai (Deploy)

### 1) Chuẩn bị cấu hình

- Tạo file cấu hình runtime: sao chép từ mẫu `config-sample.yaml` thành `configs/config.yaml`, sau đó chỉnh sửa theo môi trường:

```yaml
database:
  uri: "mongodb://localhost:27017"
  name: "chatmix"
```

- Tạo file môi trường MongoDB `.env.mongodb` (đặt ở thư mục gốc repo, cùng cấp `docker-compose.yml`):

```
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=1
```


### 2) Chạy bằng Docker Compose (local/VM)

```bash
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

- Backend sẽ lắng nghe trong container ở cổng `8080`, được publish ra ngoài tại `http://localhost:8082` (theo `docker-compose.yml`).
- Health check: `GET http://localhost:8082/api/health`.
- Log backend được mount ra máy host tại `~/Documents/chatmix/logs` (có thể đổi trong `docker-compose.yml`).

### 3) Frontend (tùy chọn cho local)

- Cập nhật `frontend/src/config/api.js` cho môi trường local:

```js
const CONFIG = {
  API_BASE_URL: 'http://localhost:8082/api',
  WS_BASE_URL: 'ws://localhost:8082',
};
export default CONFIG;
```

- Chạy dev: `npm start` (trong thư mục `frontend/`), hoặc build production: `npm run build`.

### 4) Troubleshooting

- **MongoDB AuthenticationFailed**:
  - Đảm bảo `.env.mongodb` khởi tạo đúng user/pass.
  - Cập nhật `database.uri` trong `configs/config.yaml` kèm `authSource=admin` nếu bật `--auth`.
  - Khởi động lại: `docker-compose down && docker-compose up -d`.

- **Không copy được config khi build Docker**:
  - Đã dùng đúng đường dẫn `configs/` trong Dockerfile: `COPY --from=builder /app/configs .`.

- **Frontend không kết nối**:
  - Kiểm tra `frontend/src/config/api.js` trùng khớp domain/port thực tế.