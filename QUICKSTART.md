# Quick Start Guide

## 🚀 Setup trong 5 phút

### 1. Prerequisites
```bash
# Check Go version
go version  # Cần Go 1.21+

# Check PostgreSQL
psql --version  # Cần PostgreSQL 14+
```

### 2. Clone & Setup
```bash
cd auth-service
cp .env.example .env

# Chỉnh sửa .env
# QUAN TRỌNG: Đổi JWT_SECRET thành random string dài ít nhất 32 ký tự
```

### 3. Start Database
```bash
# Option 1: Docker
make docker-up

# Option 2: Local PostgreSQL
createdb auth_db
```

### 4. Run Service
```bash
# Development mode
make run

# Or with hot reload (cần cài air)
make dev

# Build binary
make build
```

## 📡 Test API

### Register User
```bash
curl -X POST http://localhost:9001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

### Login
```bash
curl -X POST http://localhost:9001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'

# Response sẽ có access_token và set cookie refresh_token
```

### Get Current User (Protected)
```bash
# Thay YOUR_ACCESS_TOKEN bằng token từ login response
curl http://localhost:9001/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Refresh Token
```bash
# Cookie refresh_token tự động gửi từ login
curl -X POST http://localhost:9001/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

### Change Password
```bash
curl -X POST http://localhost:9001/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "SecurePass123!",
    "new_password": "NewSecurePass123!"
  }'
```

### Logout
```bash
curl -X POST http://localhost:9001/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -b cookies.txt
```

### Logout All Sessions
```bash
curl -X POST http://localhost:9001/api/v1/auth/logout-all \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -b cookies.txt
```

### Health Check
```bash
curl http://localhost:9001/health
```

## 🐛 Common Issues

### Issue: "DATABASE_URL is required"
**Fix**: Copy `.env.example` to `.env` và cập nhật DATABASE_URL

### Issue: "Connection refused" when starting
**Fix**: 
1. Kiểm tra PostgreSQL đang chạy: `pg_isready`
2. Kiểm tra port 9001 chưa được dùng: `lsof -i :9001`

### Issue: Build fails với Go version error
**Fix**: Update Go version ít nhất 1.21: `go version`

### Issue: "weak password" error
**Fix**: Password phải có:
- Ít nhất 8 ký tự
- Ít nhất 1 chữ hoa
- Ít nhất 1 chữ thường
- Ít nhất 1 số
- Ít nhất 1 ký tự đặc biệt

## 📂 Project Structure Overview

```
auth-service/
├── cmd/server/              # Entry point
├── internal/
│   ├── domain/             # Business entities & rules
│   ├── application/        # Use cases
│   ├── infrastructure/     # External dependencies
│   └── presentation/       # HTTP layer
├── pkg/                    # Public packages
├── .env                    # Configuration
├── Makefile               # Build commands
└── docker-compose.yml     # Docker setup
```

## 🔧 Development Commands

```bash
# Run tests
make test

# Run tests với coverage
make test-coverage

# Run linter
make lint

# Build Docker image
make docker-build

# Start với Docker
make docker-up

# Stop Docker
make docker-down

# Clean build artifacts
make clean

# Download dependencies
make deps
```

## 📚 Next Steps

1. **Read Documentation**
   - [README.md](./README.md) - Full documentation
   - [ARCHITECTURE.md](./ARCHITECTURE.md) - Architecture deep dive

2. **Customize**
   - Update `.env` với production values
   - Configure CORS origins
   - Adjust rate limiting
   - Set proper JWT secret

3. **Add Features**
   - Email verification
   - Password reset
   - OAuth2 integration
   - Two-factor authentication

4. **Deploy**
   - Build Docker image
   - Deploy to Kubernetes
   - Setup monitoring
   - Configure CI/CD

## 💡 Tips

1. **Development**: Use `make dev` với air cho hot reload
2. **Testing**: Use Postman collection hoặc cURL scripts
3. **Debugging**: Check logs trong terminal, structured JSON logs
4. **Security**: Never commit `.env` file, use strong JWT_SECRET
5. **Performance**: Monitor với `/health` endpoint

## 🆘 Need Help?

- Architecture questions: See [ARCHITECTURE.md](./ARCHITECTURE.md)
- API documentation: See [README.md](./README.md#api-endpoints)
- Contributing: Follow Clean Architecture principles

## 🎯 Production Checklist

- [ ] Update JWT_SECRET to strong random string (32+ chars)
- [ ] Set COOKIE_SECURE=true for HTTPS
- [ ] Configure proper ALLOWED_ORIGINS for CORS
- [ ] Set appropriate RATE_LIMIT_PER_MINUTE
- [ ] Configure database connection pool limits
- [ ] Enable Redis caching (optional)
- [ ] Setup monitoring & alerting
- [ ] Configure proper logging
- [ ] Setup backup strategy
- [ ] Run security scan
- [ ] Load testing
- [ ] Setup CI/CD pipeline
