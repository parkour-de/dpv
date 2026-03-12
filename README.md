# DPV Mitgliederverwaltung

The **Deutscher Parkour Verband (DPV) Membership Management API** is a backend service designed to manage users, parkour sports clubs, organizations, and their memberships within the German Parkour Association infrastructure.

## About

This API serves as the backbone for a comprehensive membership management platform that allows:

- **User Management**: Registration, authentication, and profile management for individuals
- **Club & Organization Management**: Creation and administration of parkour clubs and organizations  
- **Membership Processing**: Applications, approvals, and hierarchical membership structures
- **Verification Systems**: Email verification and password reset workflows
- **Graph-Based Relationships**: Support for complex organizational hierarchies (e.g., Landesverbände)

## Features

- ✅ **User Registration & Authentication**: Secure user accounts with strong password requirements
- ✅ **HTTP Basic Authentication**: Stateless authentication for API requests
- ✅ **Email Verification**: Secure email verification and change workflows
- ✅ **Password Reset**: Self-service password reset with secure token-based links
- ✅ **Club Management**: Create and manage parkour clubs and organizations
- ✅ **Membership Applications**: Apply for and process DPV memberships
- 🚧 **Graph Relationships**: Handle complex organizational hierarchies (planned)

## Technology Stack

- **Language**: Go 1.25
- **Router**: [httprouter](https://github.com/julienschmidt/httprouter) - Fast HTTP routing
- **Database**: [ArangoDB](https://arangodb.com/) - Multi-model database (documents + graphs)
- **Authentication**: HTTP Basic Auth with bcrypt password hashing
- **Configuration**: YAML-based configuration management
- **Testing**: Go's built-in testing with test database support

## Project Structure

```
src/
├── cmd/membership/         # Application entry point
├── api/                    # Core API utilities and response helpers
├── domain/
│   └── entities/          # Data models (User, Club, etc.)
├── endpoints/             # HTTP handlers and request/response logic
│   └── users/            # User-related endpoints
├── middleware/            # HTTP middleware (auth, CORS, etc.)
├── repository/            # Data access layer
│   ├── dpv/              # Configuration management
│   ├── graph/            # ArangoDB connection and queries
│   ├── security/         # Password hashing and token generation
│   ├── storage/          # Structured document storage
│   └── t/                # Translation and error handling
├── router/               # HTTP routing setup
└── service/              # Business logic layer
    ├── club/            # Club and membership logic
    └── user/            # User business logic
```

## Prerequisites

- **Go**: Version 1.25 or higher
- **ArangoDB**: Version 3.x (local or remote instance)
- **Make**: For build automation

## Getting Started

### 1. Database Setup

Start ArangoDB locally using Docker:

```
docker run -d \
--name arangodb \
-p 8529:8529 \
-e ARANGO_ROOT_PASSWORD=change-me \
arangodb/arangodb:latest
```

### 2. Configuration

Copy the example configuration and customize it:

```
cp cfg/config.example.yml cfg/config.yml
```

Update `cfg/config.yml` with your ArangoDB credentials:

```
db:
  host: localhost
  port: 8529
  user: root
  pass: change-me
auth:
  dpv_secret_key: your-secret-key-here
```

### 3. Build and Run

Build the application:

```
make build
```

Run tests:

```
make test
```

Start the server:

```
make run
# Or with custom port:
PORT=3000 make run
# Or with Unix socket:
UNIX=/tmp/dpv.sock make run
```

The API will be available at `http://localhost:8080` (or your specified port).

## API Endpoints

### Public Endpoints

- `GET /dpv/version` - Get API version
- `POST /dpv/users` - Register a new user

### Authenticated Endpoints (require HTTP Basic Auth)

- `GET /dpv/users/me` - Get current user profile
- `PATCH /dpv/user/:key/roles` - Update user roles (Admin only)
- `GET /dpv/clubs` - List clubs (with pagination/filtering)
- `POST /dpv/clubs` - Create a new club
- `GET /dpv/club/:key` - Get club details
- `PATCH /dpv/club/:key` - Update club details
- `DELETE /dpv/club/:key` - Delete a club
- `POST /dpv/club/:key/apply` - Apply for membership
- `POST /dpv/club/:key/approve` - Approve membership (Admin only)
- `POST /dpv/club/:key/deny` - Deny membership (Admin only)
- `POST /dpv/club/:key/cancel` - Cancel/reset membership
- `POST /dpv/club/:key/documents` - Upload club documents

### Example Usage

**Register a new user:**
```
curl -X POST http://localhost:8080/dpv/users \
-H "Content-Type: application/json" \
-d '{
"email": "user@example.com",
"password": "SecurePass123!",
"lastname": "Doe",
"firstname": "John",
"dateOfBirth": "1990-01-01",
"consent_privacy": true
}'
```

**Get current user (with authentication):**
```
curl -X GET http://localhost:8080/dpv/users/me \
-u "user@example.com:SecurePass123!"
```

## Password Requirements

Passwords must meet the following criteria:
- Minimum 10 characters
- At least 8 different character types
- Cannot be only digits, only upper case, only lower case, etc.

## Development

### Available Make Commands

```
make help          # Show all available commands
make build         # Build the binary
make test          # Run all tests
make run           # Run the application
make docker-build  # Build Docker image
make docker-run    # Run in Docker container
make raml          # Generate API documentation
make strings       # Update translatable strings
```

### Running Tests

```
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific test package
go test ./src/service/user/
```

### Docker Support

Build and run with Docker:

```
make docker-build
make docker-run
```

Stop the container:

```
make docker-stop
```

## API Documentation

API documentation is available in RAML format in the `docs/` directory:

- `docs/api.raml` - Main API specification
- `docs/api.html` - Generated HTML documentation

Generate updated documentation:

```
make raml
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add tests for new functionality
- Keep commit messages descriptive

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

### Phase 1: Core User Management ✅
- [x] User registration and authentication
- [x] Password security and validation
- [x] Basic HTTP API structure

### Phase 2: Extended User Features 🚧
- [x] Email verification system
- [x] Password reset workflows
- [x] User profile updates

### Phase 3: Club Management ✅
- [x] Club/organization creation
- [x] Membership roles and permissions
- [x] Document upload and verification

### Phase 4: Membership Processing ✅
- [x] Membership applications
- [x] Approval workflows
- [x] Fee calculation and management

### Phase 5: Graph Relationships 📋
- [ ] Hierarchical organization support (Landesverbände)
- [ ] Complex membership structures
- [ ] Automated member counting and voting rights

## Support

For questions or issues, please:

1. Check existing [GitHub Issues](https://github.com/parkour-de/dpv/issues)
2. Create a new issue with detailed description
3. Contact the development team

---

**Deutscher Parkour Verband** - Building the infrastructure for parkour in Germany 🏃‍♂️