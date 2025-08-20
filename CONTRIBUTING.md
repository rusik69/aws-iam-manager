# Contributing to AWS IAM Manager

Thank you for your interest in contributing to AWS IAM Manager! This document provides guidelines for contributing to the project.

## Code of Conduct

Please note that this project adheres to security-first principles. All contributions must be related to defensive security practices.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Docker (optional, for containerized development)

### Setting up the development environment

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd aws-iam-manager
   ```

2. **Backend setup:**
   ```bash
   cd backend
   go mod download
   make install-linter
   ```

3. **Frontend setup:**
   ```bash
   cd frontend
   npm install
   ```

## Development Workflow

### Before making changes

1. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Set up your AWS credentials (for testing):
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

### Backend Development

1. **Make your changes** in the appropriate package:
   - `internal/handlers/` - HTTP handlers
   - `internal/services/` - Business logic
   - `internal/models/` - Data structures
   - `pkg/config/` - Configuration

2. **Run tests:**
   ```bash
   cd backend
   make test
   ```

3. **Run linter:**
   ```bash
   make lint
   ```

4. **Run the complete CI pipeline:**
   ```bash
   make ci
   ```

### Frontend Development

1. **Start development server:**
   ```bash
   cd frontend
   npm run dev
   ```

2. **Run tests:**
   ```bash
   npm test
   ```

3. **Build for production:**
   ```bash
   npm run build
   ```

### Code Quality Standards

1. **Backend (Go):**
   - Follow standard Go conventions
   - Write meaningful package and function comments
   - Ensure all linter rules pass
   - Write unit tests for new functionality
   - Use dependency injection for testability

2. **Frontend (Vue):**
   - Follow Vue.js best practices
   - Write component tests using Vitest
   - Use proper TypeScript types where applicable

3. **Security:**
   - Never commit AWS credentials or other secrets
   - Follow defensive security practices
   - Validate all inputs
   - Handle errors gracefully without exposing sensitive information

## Testing

### Running Tests

**Backend:**
```bash
cd backend
make test
make test-coverage  # with coverage report
make test-verbose   # with verbose output
```

**Frontend:**
```bash
cd frontend
npm test
npm test:ui  # with UI interface
```

**Integration:**
```bash
# Build and test the complete application
cd backend
make build
./bin/server &  # Start server
# Test endpoints...
kill %1  # Stop server
```

### Writing Tests

- Write unit tests for all new functionality
- Include both positive and negative test cases
- Mock external dependencies (AWS API calls)
- Test error conditions and edge cases
- Maintain high test coverage (aim for >80%)

## Submitting Changes

1. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```
   
   Use conventional commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for adding tests
   - `refactor:` for code refactoring

2. **Push to your branch:**
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request:**
   - Provide a clear description of the changes
   - Include test results
   - Reference any related issues
   - Ensure all CI checks pass

## Pull Request Guidelines

- **Title:** Use a descriptive title that summarizes the change
- **Description:** Provide detailed information about:
  - What changes were made
  - Why the changes were necessary
  - How to test the changes
  - Any breaking changes or migration notes

- **Checklist:**
  - [ ] Tests added/updated
  - [ ] Documentation updated
  - [ ] Linting passes
  - [ ] No security vulnerabilities introduced
  - [ ] Backward compatibility maintained

## Getting Help

- **Issues:** Use GitHub Issues for bug reports and feature requests
- **Discussions:** Use GitHub Discussions for questions and general discussion
- **Security:** For security-related issues, please follow responsible disclosure practices

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.