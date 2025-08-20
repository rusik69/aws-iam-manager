# AWS IAM Manager - Backend

A Go-based backend service for managing AWS IAM users and access keys across multiple AWS accounts.

## Project Structure

This project follows the standard Go project layout conventions:

```
backend/
├── cmd/
│   └── server/           # Main application entry point
│       └── main.go       # Application bootstrap
├── internal/             # Private application code
│   ├── handlers/         # HTTP handlers (controllers)
│   │   ├── handlers.go
│   │   └── handlers_test.go
│   ├── models/           # Data models/structs
│   │   └── user.go
│   ├── server/           # Server setup and configuration
│   │   ├── server.go
│   │   └── frontend/     # Embedded frontend files
│   └── services/         # Business logic
│       ├── aws_service.go
│       ├── aws_service_test.go
│       └── interfaces.go
├── pkg/                  # Public library code
│   └── config/           # Configuration management
│       ├── config.go
│       └── config_test.go
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # This file
```

## Architecture

### Separation of Concerns

- **`cmd/`**: Contains the main application entry points
- **`internal/`**: Private application code that cannot be imported by other projects
  - **`handlers/`**: HTTP request handlers (like controllers in MVC)
  - **`models/`**: Data structures and models
  - **`server/`**: Server setup, routing, and middleware configuration
  - **`services/`**: Business logic and external service integrations
- **`pkg/`**: Public library code that can be imported by other projects

### Key Components

1. **Models** (`internal/models/`): Define data structures for Account, User, and AccessKey
2. **Services** (`internal/services/`): Handle AWS API interactions and business logic
3. **Handlers** (`internal/handlers/`): Process HTTP requests and responses
4. **Server** (`internal/server/`): Configure routes, middleware, and embedded frontend
5. **Config** (`pkg/config/`): Manage application configuration

## Building and Running

### Build the application:
```bash
go build ./cmd/server
```

### Run the application:
```bash
./server
```

### Run tests:
```bash
# Run all tests
go test ./...

# Run tests for specific packages
go test ./internal/handlers
go test ./internal/services
go test ./pkg/config
```

### Environment Variables

- `PORT`: Server port (default: 8080)
- `AWS_REGION`: AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID`: AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key

## Features

- List AWS accounts in an organization
- List IAM users for specific accounts
- View detailed user information
- Create new access keys
- Delete existing access keys
- Rotate access keys (create new, delete old)

## API Endpoints

- `GET /api/accounts` - List all AWS accounts
- `GET /api/accounts/:accountId/users` - List users in an account
- `GET /api/accounts/:accountId/users/:username` - Get specific user details
- `POST /api/accounts/:accountId/users/:username/keys` - Create new access key
- `DELETE /api/accounts/:accountId/users/:username/keys/:keyId` - Delete access key
- `PUT /api/accounts/:accountId/users/:username/keys/:keyId/rotate` - Rotate access key

## Security

This application implements defensive security practices:
- Only provides read access to user information
- Access key operations are logged and controlled
- No secrets are stored or logged
- Proper error handling to prevent information leakage