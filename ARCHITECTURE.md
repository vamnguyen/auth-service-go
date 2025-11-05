# Auth Service - Clean Architecture Documentation

## 📐 Architecture Overview

Dự án được xây dựng theo **Clean Architecture** (Uncle Bob), đảm bảo tách biệt rõ ràng giữa các layers và dễ dàng test, maintain, scale.

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│              (HTTP Handlers, Middleware, Router)             │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                   Application Layer                          │
│              (Use Cases, DTOs, Interfaces)                   │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                     Domain Layer                             │
│         (Entities, Repository Interfaces, Errors)            │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  Infrastructure Layer                        │
│    (Database, Security, Logging, External Services)          │
└─────────────────────────────────────────────────────────────┘
```

## 🏗️ Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
**Enterprise Business Rules** - Core business logic, không phụ thuộc vào bất kỳ layer nào khác.

```
domain/
├── entity/              # Business entities
│   ├── user.go         # User entity với business logic
│   ├── refresh_token.go
│   └── audit_log.go
├── repository/          # Repository interfaces (abstractions)
│   ├── user_repository.go
│   ├── refresh_token_repository.go
│   └── audit_log_repository.go
└── error/              # Domain-specific errors
    └── errors.go
```

**Key Principles:**
- Entities chứa business logic (password verification, account locking, etc.)
- Repository interfaces định nghĩa contract, không implement
- Không import bất kỳ package nào từ layers khác
- Pure business rules, không biết về database, HTTP, etc.

**Example:**
```go
// Entity với business logic
type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    // ... other fields
}

func (u *User) VerifyPassword(password string) error {
    return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (u *User) IsAccountLocked() bool {
    // Business logic for account locking
}
```

### 2. Application Layer (`internal/application/`)
**Application Business Rules** - Orchestrate data flow, implement use cases.

```
application/
├── usecase/            # Use cases (business flows)
│   └── auth_usecase.go # Login, Register, Refresh, etc.
└── dto/                # Data Transfer Objects
    └── auth_dto.go     # Request/Response structures
```

**Key Principles:**
- Use cases orchestrate domain entities
- Định nghĩa interfaces cho dependencies (TokenService, PasswordService)
- Không biết về HTTP, database specifics
- Input/Output qua DTOs

**Example:**
```go
func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
    // 1. Get user from repository
    user, err := uc.userRepo.FindByEmail(ctx, req.Email)
    
    // 2. Verify password (domain logic)
    if err := user.VerifyPassword(req.Password); err != nil {
        // Handle failed attempt
        user.IncrementFailedLoginAttempts(...)
    }
    
    // 3. Generate tokens
    accessToken, _ := uc.tokenService.GenerateAccessToken(user.ID.String())
    
    // 4. Create audit log
    // 5. Return DTO
}
```

### 3. Infrastructure Layer (`internal/infrastructure/`)
**Frameworks & Drivers** - External dependencies implementation.

```
infrastructure/
├── persistence/postgres/   # Database implementations
│   ├── database.go        # DB connection setup
│   ├── user_repository.go # Implement domain repository interface
│   ├── refresh_token_repository.go
│   └── audit_log_repository.go
├── security/              # Security implementations
│   ├── jwt_service.go    # JWT generation/validation
│   └── password_service.go # Password strength validation
├── logger/               # Logging implementation
│   └── logger.go        # Structured logging với Zap
└── config/              # Configuration management
    └── config.go
```

**Key Principles:**
- Implement interfaces từ Domain/Application layers
- Handle external dependencies (database, cache, email, etc.)
- Có thể swap implementations dễ dàng

**Example:**
```go
// Implement domain repository interface
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    // Database-specific implementation
    var model UserModel
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
    return r.toEntity(&model), err
}
```

### 4. Presentation Layer (`internal/presentation/`)
**Interface Adapters** - Convert data từ external format (HTTP) sang internal format (use cases).

```
presentation/
├── http/
│   ├── handler/           # HTTP handlers
│   │   ├── auth_handler.go
│   │   └── health_handler.go
│   └── router/           # Route definitions
│       └── router.go
└── middleware/           # HTTP middleware
    ├── auth_middleware.go
    ├── cors_middleware.go
    ├── rate_limit_middleware.go
    ├── logger_middleware.go
    └── recovery_middleware.go
```

**Key Principles:**
- Convert HTTP requests thành DTOs
- Convert DTOs thành HTTP responses
- Handle HTTP-specific concerns (headers, cookies, status codes)
- Middleware cho cross-cutting concerns

**Example:**
```go
func (h *AuthHandler) Login(c *gin.Context) {
    // 1. Parse HTTP request
    var req dto.LoginRequest
    c.ShouldBindJSON(&req)
    
    // 2. Call use case
    result, err := h.authUseCase.Login(c.Request.Context(), req, c.ClientIP(), c.GetHeader("User-Agent"))
    
    // 3. Set HTTP cookie
    h.setRefreshCookie(c, result.RefreshToken)
    
    // 4. Return HTTP response
    response.Success(c, http.StatusOK, result)
}
```

## 🔄 Dependency Flow

```
main.go
  │
  ├──> Load Config
  │
  ├──> Initialize Infrastructure
  │     ├── Database Connection
  │     ├── Logger
  │     └── Security Services (JWT, Password)
  │
  ├──> Initialize Repositories (Infra implements Domain interfaces)
  │
  ├──> Initialize Use Cases (Application uses Domain & Infra)
  │
  ├──> Initialize Handlers (Presentation uses Application)
  │
  └──> Setup Router & Start Server
```

**Dependency Rule:**
- Domain: không depend vào ai
- Application: depend vào Domain
- Infrastructure: implement Domain/Application interfaces
- Presentation: depend vào Application

## 🎯 Key Design Patterns

### 1. Repository Pattern
Tách biệt business logic khỏi data access logic.

```go
// Domain defines interface
type UserRepository interface {
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}

// Infrastructure implements
type PostgresUserRepository struct {
    db *gorm.DB
}
```

### 2. Dependency Injection
Constructor injection cho tất cả dependencies.

```go
func NewAuthUseCase(
    userRepo repository.UserRepository,
    refreshRepo repository.RefreshTokenRepository,
    // ... other dependencies
) *AuthUseCase {
    return &AuthUseCase{...}
}
```

### 3. Strategy Pattern
Interfaces cho interchangeable implementations.

```go
type TokenService interface {
    GenerateAccessToken(userID string) (string, error)
    ValidateAccessToken(token string) (string, error)
}

// Có thể swap JWT implementation with OAuth2, etc.
```

### 4. DTO Pattern
Separate internal models từ external representation.

```go
// External DTO
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Internal Entity
type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    // ... business fields
}
```

## 🔐 Security Design

### 1. Password Security
- **Hashing**: bcrypt với DefaultCost (10)
- **Validation**: Password strength requirements
- **Storage**: Never store plain text passwords

### 2. Token Security
- **Access Token**: JWT, short-lived (15 minutes)
- **Refresh Token**: Random 32 bytes, hashed với SHA256, long-lived (30 days)
- **Storage**: Refresh token hash stored in DB, plain token in HttpOnly cookie
- **Rotation**: Refresh token rotated on each use

### 3. Account Protection
- **Rate Limiting**: Per-IP rate limiting
- **Account Lockout**: Lock after N failed attempts
- **Audit Logging**: Log all security events

## 📊 Database Schema

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    is_verified BOOLEAN DEFAULT false,
    is_locked BOOLEAN DEFAULT false,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_verified ON users(is_verified);

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR UNIQUE NOT NULL,  -- SHA256 hash
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked);

-- Composite index for common queries
CREATE INDEX idx_refresh_tokens_lookup ON refresh_tokens(user_id, revoked, expires_at);

-- Audit logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

## 🚀 Scalability Considerations

### 1. Horizontal Scaling
- **Stateless Design**: Không có session state trên server
- **Load Balancer Ready**: Multiple instances có thể chạy behind load balancer
- **Cookie-based Refresh Token**: Không cần shared session storage

### 2. Database Optimization
- **Connection Pooling**: Configured với reasonable limits
- **Indexes**: Proper indexes cho common queries
- **Query Optimization**: Use context for query cancellation

### 3. Future Enhancements
- **Redis Caching**: Cache user sessions, rate limit counters
- **Read Replicas**: Separate read/write operations
- **Event-Driven**: Publish domain events (user registered, logged in)
- **CQRS**: Separate read/write models if needed

## 🧪 Testing Strategy

### 1. Unit Tests
- Domain entities business logic
- Use cases with mocked repositories
- Service implementations

### 2. Integration Tests
- Repository implementations với test database
- Use cases với real repositories
- HTTP handlers với test server

### 3. E2E Tests
- Full flow: register → login → protected routes
- Token refresh flow
- Account lockout scenarios

## 📝 Adding New Features

### Example: Add Email Verification

1. **Domain Layer**
```go
// entity/verification_token.go
type VerificationToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Token     string
    ExpiresAt time.Time
}

// repository/verification_token_repository.go
type VerificationTokenRepository interface {
    Create(ctx context.Context, token *VerificationToken) error
    FindByToken(ctx context.Context, token string) (*VerificationToken, error)
}
```

2. **Application Layer**
```go
// usecase/auth_usecase.go
func (uc *AuthUseCase) SendVerificationEmail(ctx context.Context, userID string) error {
    // Generate token
    // Send email
    // Store token
}

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
    // Find token
    // Mark user as verified
    // Delete token
}
```

3. **Infrastructure Layer**
```go
// persistence/postgres/verification_token_repository.go
type VerificationTokenRepository struct {
    db *gorm.DB
}
// Implement methods

// email/email_service.go
type EmailService struct {}
func (s *EmailService) SendVerificationEmail(to, token string) error {
    // SMTP implementation
}
```

4. **Presentation Layer**
```go
// handler/auth_handler.go
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
    token := c.Query("token")
    err := h.authUseCase.VerifyEmail(c.Request.Context(), token)
    // Handle response
}

// router/router.go
auth.GET("/verify-email", authHandler.VerifyEmail)
```

## 🎓 Best Practices

1. **Never break the Dependency Rule**: Inner layers không depend vào outer layers
2. **Use interfaces**: Định nghĩa contracts, swap implementations dễ dàng
3. **Context everywhere**: Pass context.Context cho cancellation và tracing
4. **Error handling**: Return errors, don't panic (except startup)
5. **Logging**: Structured logging cho observability
6. **Configuration**: Environment-based config, validate at startup
7. **Graceful shutdown**: Handle signals, close connections properly
8. **Security first**: Validate input, sanitize output, audit actions

## 📚 References

- [Clean Architecture by Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
