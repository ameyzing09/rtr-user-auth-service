# RTR User Authentication Service

A multi-tenant user authentication and authorization service built with Go, Gin, and MySQL. This service provides secure user management, tenant isolation, and JWT-based authentication for the RTR platform.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [API Documentation](#api-documentation)
- [Permission System](#permission-system)
- [Security](#security)
- [Development](#development)
- [Deployment](#deployment)
- [Contributing](#contributing)

## Features

- **Multi-tenant Architecture**: Complete tenant isolation with secure tenant context validation
- **JWT Authentication**: Secure token-based authentication with configurable secrets
- **User Management**: Create, list, and manage users within tenant boundaries
- **Role-based Access Control**: Support for ADMIN, HR, INTERVIEWER, and CANDIDATE roles
- **Password Management**: Secure password hashing with bcrypt and forced password change support
- **Tenant Settings**: Configurable tenant-specific settings management
- **Database Migrations**: Automated database schema management with golang-migrate
- **CORS Support**: Configurable Cross-Origin Resource Sharing
- **Rate Limiting**: Tenant-specific rate limiting capabilities
- **Concurrent Access Control**: Tenant-level concurrency management

## Architecture

### Project Structure

```
rtr-user-auth-service/
├── cmd/server/           # Application entry point
├── domain/               # Domain-specific error definitions
├── handlers/             # HTTP request handlers and DTOs
├── internal/db/          # Database connection and migrations
├── middleware/           # HTTP middleware (auth, CORS, tenant context)
├── models/               # Data models and business entities
├── repositories/         # Data access layer
├── routes/               # HTTP route definitions
├── services/             # Business logic layer
└── utils/                # Utility functions (JWT, password hashing)
```

### Technology Stack

- **Language**: Go 1.24.4
- **Web Framework**: Gin
- **Database**: MySQL with GORM ORM
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt
- **Migrations**: golang-migrate
- **Environment**: godotenv

## Prerequisites

- Go 1.24.4 or later
- MySQL 8.0 or later
- golang-migrate CLI tool
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd rtr-user-auth-service
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install golang-migrate (if not already installed)**
   ```bash
   make migrate-install
   # or manually:
   go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

## Configuration

Create a `.env` file in the project root with the following variables:

```env
# Database Configuration
DB_USER=root
DB_PASSWORD=your_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=recrutr

# Alternative: Use MYSQL_DSN for complete connection string
# MYSQL_DSN=mysql://user:pass@tcp(127.0.0.1:3306)/authdb?multiStatements=true&parseTime=true

# JWT Configuration
JWT_SECRET=your_jwt_secret_key

# Tenant Context Security
TENANT_CTX_SECRET=your_tenant_context_secret
TENANT_CTX_SECRET_PREV=your_previous_secret_for_rotation

# Application Configuration
GIN_MODE=release  # or debug for development
ENV=production    # or local for development
```

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DB_USER` | MySQL username | Yes | - |
| `DB_PASSWORD` | MySQL password | Yes | - |
| `DB_HOST` | MySQL host | Yes | - |
| `DB_PORT` | MySQL port | Yes | - |
| `DB_NAME` | Database name | Yes | - |
| `JWT_SECRET` | JWT signing secret | Yes | - |
| `TENANT_CTX_SECRET` | Tenant context validation secret | Yes | - |
| `TENANT_CTX_SECRET_PREV` | Previous secret for rotation | No | - |
| `GIN_MODE` | Gin framework mode | No | debug |
| `ENV` | Application environment | No | production |

## Database Setup

1. **Create the database**
   ```sql
   CREATE DATABASE recrutr;
   ```

2. **Run migrations**
   ```bash
   make migrate-up
   ```

3. **Verify migration status**
   ```bash
   make migrate-version
   ```

### Available Migration Commands

```bash
# Apply all pending migrations
make migrate-up

# Roll back N steps (default: 1)
make migrate-down STEPS=2

# Force database to specific version
make migrate-force VERSION=2

# Check current migration version
make migrate-version

# Print environment configuration
make env-print
```

## API Documentation

### Base URL
```
http://localhost:8082
```

### Authentication

The API uses JWT tokens for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

### Tenant Context

All requests must include tenant context headers:

- `X-Tenant-Id`: UUID for the tenant (signed)
- `X-Tenant-Domain`: Optional domain (must match tenant record if provided)
- `X-Tenant-Ts`: Epoch minutes (UTC) for staleness checks
- `X-Tenant-Sig`: Base64url encoded HMAC-SHA256 signature

### Endpoints

#### Public Endpoints

| Method | Endpoint | Description | Headers Required |
|--------|----------|-------------|------------------|
| GET | `/` | Health check | None |
| POST | `/login` | User login | Tenant Context |
| GET | `/tenant/settings` | Get tenant settings | Tenant Context |

#### Protected Endpoints

| Method | Endpoint | Description | Headers Required |
|--------|----------|-------------|------------------|
| GET | `/me` | Get current user info | JWT + Tenant Context |
| POST | `/me/change-password` | Change user password | JWT + Tenant Context |
| GET | `/users` | List users in tenant | JWT + Tenant Context |
| POST | `/users` | Create new user | JWT + Tenant Context |
| PUT | `/tenant/settings` | Update tenant settings | JWT + Tenant Context |

### Request/Response Examples

#### Login
```bash
POST /login
Content-Type: application/json
X-Tenant-Id: 123e4567-e89b-12d3-a456-426614174000
X-Tenant-Domain: example.com
X-Tenant-Ts: 1640995200
X-Tenant-Sig: <signature>

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Create User
```bash
POST /users
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-Id: 123e4567-e89b-12d3-a456-426614174000

{
  "email": "newuser@example.com",
  "name": "New User",
  "role": "CANDIDATE"
}
```

## Permission System

The service implements a comprehensive server-side permission system with role-based access control (RBAC). The backend is the source of truth for all permissions - UI elements may be hidden, but server validation is always enforced.

### Roles

- **SUPERADMIN**: Full system access, can manage tenants and bypass tenant boundaries on control-plane routes only
- **ADMIN**: Tenant-level administrator, can manage users and tenant settings  
- **HR**: Human resources role, can list and create users, view tenant settings
- **INTERVIEWER**: Limited access, can only manage their own profile
- **CANDIDATE**: Limited access, can only manage their own profile

### Route Protection

- **Public Routes**: No authentication required (login, public tenant settings)
- **Protected Routes**: Require authentication + tenant context + role permissions
- **Admin Routes**: Require SUPERADMIN role, no tenant context (control-plane operations)

### Permission Enforcement

The system uses middleware-based role gates and policy-based action checking:

```go
// Role-based route protection
admin.POST("/tenant/create", middleware.RequireRole(models.RoleSuperAdmin), handler.Create)
protected.GET("/users", middleware.RequireAny(models.RoleAdmin, models.RoleHR), handler.ListUsers)

// Policy-based action checking
if !policy.Can(actor.Role, policy.ActionUserDelete) {
    return 403 Forbidden
}
```

For detailed permission matrix and implementation examples, see [docs/permissions.md](docs/permissions.md).

## Security

### Tenant Context Validation

The service validates tenant context using HMAC-SHA256 signatures:

1. **Signature Generation**: `HMAC-SHA256(tenantID + "." + domain + "." + timestamp, secret)`
2. **Validation**: Verifies signature and timestamp staleness
3. **Secret Rotation**: Supports dual secrets for zero-downtime rotation
4. **Local Development**: Allows unsigned headers when `ENV=local`

### JWT Security

- **Secret Management**: Configurable JWT secrets via environment variables
- **Token Validation**: Comprehensive token validation with claims verification
- **Tenant Isolation**: Tokens are validated against tenant context
- **Role-based Access**: JWT includes user role for authorization

### Password Security

- **Hashing**: bcrypt with configurable cost
- **Force Change**: Support for mandatory password changes
- **Validation**: Minimum length and complexity requirements

## Development

### Running the Application

1. **Start the database** (using Docker)
   ```bash
   docker run --name mysql-auth -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=recrutr -p 3306:3306 -d mysql:8.0
   ```

2. **Run migrations**
   ```bash
   make migrate-up
   ```

3. **Start the application**

   **For development with hot reload:**
   ```bash
   make dev
   ```
   This uses `air` to automatically restart the server when Go files change.

   **For manual runs:**
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./middleware -v
```

### Code Structure Guidelines

- **Handlers**: HTTP request/response handling and validation
- **Services**: Business logic and orchestration
- **Repositories**: Data access and persistence
- **Models**: Domain entities and data structures
- **Middleware**: Cross-cutting concerns (auth, CORS, logging)

## Deployment

### Docker Deployment

```dockerfile
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Environment-Specific Configuration

- **Development**: `GIN_MODE=debug`, `ENV=local`
- **Staging**: `GIN_MODE=release`, `ENV=staging`
- **Production**: `GIN_MODE=release`, `ENV=production`

### Health Checks

The service provides a health check endpoint at `/` that returns `200 OK` when the service is running.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards and best practices
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all migrations are reversible
- Use meaningful commit messages

## License

This project is licensed under the MIT License - see the LICENSE file for details.
