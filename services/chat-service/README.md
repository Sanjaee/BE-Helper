# Chat Service - Real-time Messaging

Real-time chat service untuk komunikasi antara client dan provider menggunakan WebSocket dan event-driven architecture.

## 🎯 Features

- ✅ Real-time messaging menggunakan WebSocket
- ✅ Event-driven dengan RabbitMQ
- ✅ Chat history tersimpan di database
- ✅ Unread message counter
- ✅ Mark messages as read
- ✅ Efficient resource usage (WebSocket only ketika diperlukan)
- ✅ Scalable architecture

## 🏗️ Architecture

### Backend (Go)
```
chat-service/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── config/                 # Configuration
│   ├── database/               # Database connection & migrations
│   ├── events/                 # RabbitMQ integration
│   ├── handlers/               # HTTP & WebSocket handlers
│   ├── models/                 # Data models
│   ├── repository/             # Database operations
│   ├── services/               # Business logic
│   └── websocket/              # WebSocket hub & client management
├── Dockerfile
├── go.mod
└── env.example
```

### Frontend (Flutter)
```
mobile_helper/
├── lib/
│   ├── data/
│   │   ├── models/
│   │   │   └── chat_model.dart         # Chat message model
│   │   └── services/
│   │       └── chat_service.dart       # Chat API service
│   └── presentation/
│       └── chat/
│           └── chat_page.dart          # Chat UI (universal)
```

## 🔄 How It Works

### 1. **Order Acceptance**
Ketika provider menerima order, chat channel terbuka secara otomatis.

### 2. **WebSocket Connection**
- Client/Provider hanya connect ke WebSocket ketika membuka chat page
- Connection closed ketika meninggalkan chat page
- Tidak ada background WebSocket (efficient!)

### 3. **Message Flow**
```
User sends message → HTTP POST → Database → WebSocket Broadcast → All connected clients
                              ↓
                         RabbitMQ Event (for notifications)
```

### 4. **Notification System**
- Message events dikirim ke RabbitMQ
- Notification service bisa consume untuk push notifications
- Chat history tetap tersimpan untuk nanti

## 📡 API Endpoints

### REST API

#### Send Message
```http
POST /api/v1/chats/messages
Content-Type: application/json

{
  "order_id": "uuid",
  "sender_id": "uuid",
  "sender_type": "client" | "provider",
  "message": "Hello!"
}
```

#### Get Chat History
```http
GET /api/v1/chats/order/:order_id
```

#### Get Unread Count
```http
GET /api/v1/chats/order/:order_id/unread?user_id=uuid
```

#### Mark as Read
```http
PATCH /api/v1/chats/order/:order_id/read
Content-Type: application/json

{
  "user_id": "uuid"
}
```

### WebSocket

#### Connect to Chat
```
ws://localhost:5000/api/v1/ws/chat/:order_id?user_id=:user_id
```

#### Message Format
```json
{
  "type": "new_message",
  "message": {
    "id": "uuid",
    "order_id": "uuid",
    "sender_id": "uuid",
    "sender_type": "client",
    "message": "Hello!",
    "is_read": false,
    "created_at": "2025-01-01T10:00:00Z"
  }
}
```

## 🚀 Setup & Run

### Prerequisites
- Go 1.21+
- PostgreSQL
- RabbitMQ
- Docker (optional)

### Environment Variables
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123
DB_NAME=chatdb

RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=admin
RABBITMQ_PASSWORD=secret123

PORT=5005
```

### Run with Docker Compose
```bash
cd be-helper
docker-compose up -d chat-service
```

### Run Standalone
```bash
cd services/chat-service
go mod download
go run cmd/main.go
```

## 💾 Database Schema

### Table: chat_messages
```sql
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('client', 'provider')),
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP
);

CREATE INDEX idx_chat_order_id ON chat_messages(order_id);
CREATE INDEX idx_chat_created_at ON chat_messages(created_at);
CREATE INDEX idx_chat_unread ON chat_messages(order_id, is_read) WHERE is_read = FALSE;
```

## 📱 Flutter Integration

### 1. Install Dependencies
```yaml
dependencies:
  web_socket_channel: ^2.4.0
```

### 2. Usage Example
```dart
// Open chat
Navigator.push(
  context,
  MaterialPageRoute(
    builder: (context) => ChatPage(
      orderId: order.id,
      orderNumber: order.orderNumber,
      userType: 'client', // or 'provider'
    ),
  ),
);
```

### 3. Chat Button Integration
Chat button sudah ditambahkan di:
- **Client Side**: `provider_on_the_way_page.dart` (samping "Provider is on the way")
- **Provider Side**: `navigation_page.dart` (samping order number)

## 🎨 UI Features

### Chat Page
- ✅ Real-time messaging
- ✅ Message bubbles (different colors for sender/receiver)
- ✅ Auto-scroll to bottom
- ✅ Timestamp formatting
- ✅ Typing indicator ready
- ✅ Read receipts ready
- ✅ Clean and modern UI

### Chat Icon
- ✅ Icon: `chat_bubble_outline`
- ✅ Color matches theme (green for client, teal for provider)
- ✅ Tooltip untuk UX yang lebih baik

## ⚡ Performance Optimization

### 1. **Lazy WebSocket Connection**
- WebSocket hanya connect saat chat page dibuka
- Tidak ada background connection
- Auto disconnect saat page closed

### 2. **Efficient Polling**
- Tidak ada polling untuk unread count saat tidak diperlukan
- Unread count di-fetch on-demand

### 3. **Database Indexing**
- Index pada order_id untuk fast lookup
- Index pada created_at untuk chronological ordering
- Partial index untuk unread messages

### 4. **Message Broadcasting**
- WebSocket hub mengelola connections per order
- Broadcast hanya ke clients yang terhubung di order tersebut
- No global broadcasts

## 🔐 Security

- ✅ Authentication required (JWT token)
- ✅ User dapat mengakses chat untuk order mereka saja
- ✅ Message validation
- ✅ XSS protection (message escaping)
- ✅ Rate limiting ready (di API Gateway)

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:5005/health
```

### Logs
```bash
docker logs chat-service
```

### RabbitMQ Management
```
http://localhost:15672
Username: admin
Password: secret123
```

## 🐛 Troubleshooting

### WebSocket tidak connect
1. Pastikan chat service running
2. Check API Gateway WebSocket proxy
3. Verify user_id parameter

### Message tidak realtime
1. Check WebSocket connection
2. Verify RabbitMQ running
3. Check browser console untuk errors

### Database error
1. Pastikan UUID extension enabled
2. Verify database migrations
3. Check connection string

## 🚦 Status

✅ Backend: Complete
✅ Database: Complete
✅ WebSocket: Complete
✅ Frontend: Complete
✅ Integration: Complete
✅ Docker: Complete

## 📝 Next Steps (Optional Enhancements)

- [ ] Push notifications untuk offline users
- [ ] Typing indicators
- [ ] Read receipts visualization
- [ ] Image/file sharing
- [ ] Message reactions
- [ ] Search in chat history
- [ ] Delete messages
- [ ] Chat export

## 📞 Support

Untuk pertanyaan atau issues, silakan buat issue di repository atau hubungi developer team.

---

**Built with ❤️ using Go, Flutter, PostgreSQL, RabbitMQ, and WebSocket**

