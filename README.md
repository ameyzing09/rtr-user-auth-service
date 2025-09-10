# Recrutr Auth Service

A production-ready, multi-tenant authentication service built with Go, Gin, GORM, and MySQL. This service provides comprehensive user authentication and authorization for the Recrutr platform with strict tenant isolation and role-based access control (RBAC).

## Features

### 🔐 Authentication & Authorization
- **JWT Authentication** with HS256 signing algorithm
- **Multi-tenant architecture** with strict tenant isolation
- **Role-Based Access Control (RBAC)** with 4 distinct roles:
  - `ADMIN`: Full system access across all tenants
  - `HR`: Manage users within their tenant
  - `INTERVIEWER`: Limited access within their tenant
  - `CANDIDATE`: Basic access within their tenant
- **Secure password hashing** using bcrypt
- **Refresh token rotation** for enhanced security

### 🏢 Multi-Tenancy
- **Tenant isolation** enforced at database and API level
- **Unique email per tenant** constraint
- **Tenant-specific user management**
- **Domain-based tenant identification**

### 🛡️ Security
- **Input validation** on all endpoints
- **CORS protection** with configurable origins
- **Rate limiting** ready infrastructure
- **Secure headers** and error handling
- **No `any` types** - fully typed codebase

### 🏗️ Architecture
- **Clean Architecture** with separation of concerns
- **Repository Pattern** for data access
- **Service Layer** for business logic
- **Middleware-based** authentication and authorization
- **Comprehensive error handling**

### 📊 Production Ready
- **Docker containerization** with multi-stage builds
- **Database migrations** with GORM AutoMigrate
- **Health checks** and monitoring endpoints
- **Graceful shutdown** handling
- **Structured logging**
- **Environment-based configuration**

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **ORM**: GORM
- **Database**: MySQL 8.0
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### 1. Clone the Repository

```bash
git clone https://github.com/ameyzing09/rtr-user-auth-service.git
cd rtr-user-auth-service
```

### 2. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (important: change JWT_SECRET!)
nano .env
```

### 3. Run with Docker Compose (Recommended)

```bash
# Start all services (MySQL + Auth Service + Adminer)
make docker-compose-up

# Or manually:
docker-compose up --build -d
```

### 4. Verify Installation

```bash
# Check service health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2023-01-01T00:00:00Z",
  "service": "recrutr-auth-service"
}
```

## Development Setup

### 1. Install Dependencies

```bash
# Setup development environment
make setup-dev

# Or manually:
go mod download
go install github.com/air-verse/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. Start Database

```bash
# Start only MySQL
docker-compose up -d mysql
```

### 3. Run Development Server

```bash
# Run with hot reload
make dev

# Or run normally
make run
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
Most endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Core Endpoints

#### 🔐 Authentication
```http
POST /api/v1/auth/login           # User login
POST /api/v1/auth/refresh         # Refresh access token
POST /api/v1/auth/logout          # User logout
GET  /api/v1/auth/profile         # Get current user profile
```

#### 🏢 Tenants (Admin Only)
```http
POST /api/v1/tenants              # Create tenant
GET  /api/v1/tenants              # List tenants
GET  /api/v1/tenants/{id}         # Get tenant by ID
PUT  /api/v1/tenants/{id}         # Update tenant
DELETE /api/v1/tenants/{id}       # Delete tenant
GET  /api/v1/tenants/by-domain    # Get tenant by domain (public)
```

#### 👥 Users
```http
POST /api/v1/tenants/{tenantId}/users              # Create user (Admin/HR)
GET  /api/v1/tenants/{tenantId}/users              # List users (Admin/HR)
GET  /api/v1/tenants/{tenantId}/users/{userId}     # Get user (Owner/Admin/HR)
PUT  /api/v1/tenants/{tenantId}/users/{userId}     # Update user (Owner/Admin/HR)
DELETE /api/v1/tenants/{tenantId}/users/{userId}   # Delete user (Admin/HR)
```

### Example Requests

#### 1. Create Tenant (Admin Only)
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "domain": "acme.com"
  }'
```

#### 2. Login User
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@acme.com",
    "password": "password123"
  }'
```

#### 3. Create User (HR/Admin)
```bash
curl -X POST http://localhost:8080/api/v1/tenants/123e4567-e89b-12d3-a456-426614174000/users \
  -H "Authorization: Bearer <hr-or-admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "candidate@acme.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "role": "CANDIDATE"
  }'
```

## Role-Based Access Control

### Role Hierarchy
```
ADMIN > HR > INTERVIEWER > CANDIDATE
```

### Permissions Matrix

| Action | ADMIN | HR | INTERVIEWER | CANDIDATE |
|--------|-------|----|-----------  |-----------|
| Manage Tenants | ✅ | ❌ | ❌ | ❌ |
| Cross-tenant Access | ✅ | ❌ | ❌ | ❌ |
| Create Users | ✅ | ✅ | ❌ | ❌ |
| List All Users | ✅ | ✅ | ❌ | ❌ |
| View Own Profile | ✅ | ✅ | ✅ | ✅ |
| Update Own Profile | ✅ | ✅ | ✅ | ✅ |
| Update Other Users | ✅ | ✅ | ❌ | ❌ |
| Delete Users | ✅ | ✅ | ❌ | ❌ |

## Database Schema

### Tenants Table
```sql
CREATE TABLE tenants (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  domain VARCHAR(255) NOT NULL UNIQUE,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);
```

### Users Table
```sql
CREATE TABLE users (
  id CHAR(36) PRIMARY KEY,
  tenant_id CHAR(36) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  first_name VARCHAR(50) NOT NULL,
  last_name VARCHAR(50) NOT NULL,
  role ENUM('ADMIN', 'HR', 'INTERVIEWER', 'CANDIDATE') NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  UNIQUE KEY unique_tenant_email (tenant_id, email),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

### Refresh Tokens Table
```sql
CREATE TABLE refresh_tokens (
  id CHAR(36) PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  tenant_id CHAR(36) NOT NULL,
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  is_revoked BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

## Configuration

### Environment Variables

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=recrutr_auth
DB_USER=root
DB_PASSWORD=password

# JWT Configuration (CHANGE IN PRODUCTION!)
JWT_SECRET=your-super-secret-jwt-key-here-minimum-32-characters
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# Server Configuration
PORT=8080
GIN_MODE=debug

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Production Security Notes

1. **JWT Secret**: Use a strong, randomly generated secret (minimum 32 characters)
2. **Database Credentials**: Use strong passwords and dedicated database users
3. **CORS Origins**: Restrict to your actual frontend domains
4. **HTTPS**: Always use HTTPS in production
5. **Rate Limiting**: Implement rate limiting for production use

## Available Make Commands

```bash
make help                 # Show available commands
make build                # Build the application
make test                 # Run tests
make test-coverage        # Run tests with coverage
make run                  # Build and run the application
make dev                  # Run in development mode with hot reload
make docker-compose-up    # Start all services with Docker Compose
make docker-compose-down  # Stop all services
make fmt                  # Format Go code
make lint                 # Lint Go code
make setup-dev           # Setup development environment
```

## Testing

### Run Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/services -run TestAuthService_Login
```

### Test Coverage
The project maintains high test coverage across all layers:
- Repository layer tests with database mocking
- Service layer tests with repository mocking
- Handler tests with HTTP mocking
- Integration tests with test database

## Project Structure

```
.
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/
│   │   ├── entities/     # Domain models
│   │   └── repositories/ # Repository interfaces & implementations
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── services/         # Business logic services
│   └── utils/            # Utility functions
├── docs/                 # Swagger documentation
├── tests/                # Test files
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile            # Docker build configuration
├── Makefile              # Build automation
└── README.md             # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards
- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Use meaningful commit messages
- Document public APIs
- No `any` types allowed

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue in the GitHub repository
- Check the [API documentation](http://localhost:8080/swagger/index.html) when running locally
- Review the test files for usage examples