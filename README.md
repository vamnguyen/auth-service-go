# Auth Service - Clean Architecture

A production-ready authentication service built with Go following Clean Architecture principles.

## 🏗️ Architecture

```
auth-service/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/             # Enterprise Business Rules
│   │   ├── entity/        # Domain entities
│   │   ├── repository/    # Repository interfaces
│   │   └── error/         # Domain errors
│   ├── application/       # Application Business Rules
│   │   ├── usecase/       # Use cases
│   │   └── dto/           # Data Transfer Objects
│   ├── infrastructure/    # Frameworks & Drivers
│   │   ├── persistence/   # Database implementations
│   │   ├── security/      # JWT, password, crypto
│   │   ├── logger/        # Logging
│   │   └── config/        # Configuration
│   └── presentation/      # Interface Adapters
│       ├── http/          # HTTP handlers & router
│       └── middleware/    # HTTP middleware
└── pkg/                   # Public packages
    └── response/          # Standardized API responses
```

## ✨ Features

### Core Features
- ✅ User registration with email/password
- ✅ User login with JWT access token
- ✅ Refresh token rotation (stored as hashed in DB)
- ✅ Logout (single session & all sessions)
- ✅ Password change
- ✅ Account locking after failed login attempts
- ✅ Audit logging for security events

### Security Features
- ✅ Password hashing with bcrypt
- ✅ Refresh token hashing with SHA256
- ✅ JWT with expiration
- ✅ HttpOnly cookies for refresh tokens
- ✅ Password strength validation
- ✅ Rate limiting
- ✅ CORS protection
- ✅ Account lockout mechanism

### Infrastructure
- ✅ Structured logging with Zap
- ✅ Database connection pooling
- ✅ Graceful shutdown
- ✅ Health check endpoint
- ✅ Request logging middleware
- ✅ Panic recovery middleware
- ✅ Clean error handling

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Docker & Docker Compose (optional)

### Installation

1. Clone the repository
```bash
git clone <repo-url>
cd auth-service
```

2. Copy environment file
```bash
cp .env.example .env
```

3. Update `.env` with your configuration
```bash
# Change JWT_SECRET to a strong random string
JWT_SECRET=your-secret-key-min-32-characters
DATABASE_URL=postgres://user:pass@localhost:5432/auth_db?sslmode=disable
```

4. Install dependencies
```bash
make deps
```

5. Run database migrations
```bash
make run
```

### Running

**Local development:**
```bash
make run
```

**With hot reload (requires air):**
```bash
make dev
```

**With Docker:**
```bash
make docker-up
```

## 📡 API Endpoints

### Public Endpoints

#### Health Check
```http
GET /health
```

#### Register
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "role": "user",
      "is_verified": false,
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Cookie: refresh_token=<token>
```

### Protected Endpoints (Requires Bearer Token)

#### Get Current User
```http
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

#### Change Password
```http
POST /api/v1/auth/change-password
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "old_password": "OldPass123!",
  "new_password": "NewPass123!"
}
```

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

#### Logout All Sessions
```http
POST /api/v1/auth/logout-all
Authorization: Bearer <access_token>
```

## 🔒 Security Best Practices

1. **JWT Secret**: Use a strong, random secret (min 32 characters)
2. **HTTPS**: Always use HTTPS in production (`COOKIE_SECURE=true`)
3. **CORS**: Configure `ALLOWED_ORIGINS` properly
4. **Rate Limiting**: Adjust `RATE_LIMIT_PER_MINUTE` based on your needs
5. **Database**: Use connection pooling settings appropriate for your load
6. **Passwords**: Enforces 8+ characters with uppercase, lowercase, number, and special character

## 🧪 Testing

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:9001/health
```

### Logs
Structured JSON logs in production, pretty-printed in development.

## 🐳 Docker

Build image:
```bash
make docker-build
```

Start services:
```bash
make docker-up
```

Stop services:
```bash
make docker-down
```

## 📝 Environment Variables

See `.env.example` for all available configuration options.

Key variables:
- `ENVIRONMENT`: development/production
- `JWT_SECRET`: Secret key for JWT signing
- `DATABASE_URL`: PostgreSQL connection string
- `ALLOWED_ORIGINS`: CORS allowed origins
- `RATE_LIMIT_PER_MINUTE`: Rate limit threshold

## 🛠️ Development

### Project Structure

- **Domain Layer**: Core business logic, entities, and repository interfaces
- **Application Layer**: Use cases and application-specific business rules
- **Infrastructure Layer**: External dependencies (database, security, logging)
- **Presentation Layer**: HTTP handlers, routing, and middleware

### Adding New Features

1. Define entity in `internal/domain/entity/`
2. Define repository interface in `internal/domain/repository/`
3. Implement repository in `internal/infrastructure/persistence/`
4. Create use case in `internal/application/usecase/`
5. Create handler in `internal/presentation/http/handler/`
6. Register routes in `internal/presentation/http/router/`

## 📈 Performance

- Database connection pooling configured
- Request/response logging
- Graceful shutdown handling
- Rate limiting per IP

## 🔄 Future Enhancements

- [ ] Email verification flow
- [ ] Password reset flow
- [ ] OAuth2 integration (Google, GitHub)
- [ ] Two-factor authentication (2FA)
- [ ] Redis caching layer
- [ ] Prometheus metrics
- [ ] OpenAPI/Swagger documentation
- [ ] Kubernetes manifests
- [ ] CI/CD pipeline

## 📄 License

MIT License
