# Testing Guide for Go-Drive Microservices

This document describes the testing strategy and how to run tests for the Go-Drive microservices application.

## Testing Structure

The project follows a layered testing approach:

```
go-drive/
├── internal/handlers/                  # Legacy HTTP handlers
│   └── handler_users_test.go         # HTTP handler tests with sqlmock
├── services/
│   ├── user-service/
│   │   ├── repository/
│   │   │   └── postgres_test.go      # Repository layer tests with sqlmock
│   │   └── service/
│   │       └── user_service_test.go   # gRPC service tests with mocks
│   └── api-gateway/
│       └── (integration tests - future)
└── internal/database/
    └── database_test.go               # Database connection tests
```

## Test Coverage

### 1. Repository Layer Tests (`services/user-service/repository/postgres_test.go`)

Tests the data access layer with SQL mocking:

**Tested Functions:**
- ✅ `Create` - User creation with validation
- ✅ `GetByID` - User retrieval by ID
- ✅ `Update` - User update with partial fields
- ✅ `Delete` - Soft delete functionality
- ✅ `List` - Pagination and filtering
- ✅ `VerifyEmail` - Email verification

**Test Cases per Function:**
- **Create:**
  - Successful user creation
  - Database error handling
  - Duplicate email handling

- **GetByID:**
  - Successful retrieval
  - User not found
  - Database errors

- **Update:**
  - Successful update with partial fields
  - User not found
  - Database errors during update

- **Delete:**
  - Successful soft delete
  - Database errors

- **List:**
  - Successful listing with results
  - Filtering by type
  - Filtering by active status
  - Empty results
  - Pagination
  - Database errors

- **VerifyEmail:**
  - Successful verification
  - Database errors

### 2. Service Layer Tests (`services/user-service/service/user_service_test.go`)

Tests the gRPC business logic layer with repository mocking:

**Tested gRPC Methods:**
- ✅ `CreateUser` - User creation with validation
- ✅ `GetUser` - User retrieval
- ✅ `UpdateUser` - User updates
- ✅ `DeleteUser` - User deletion
- ✅ `ListUsers` - User listing with pagination
- ✅ `VerifyEmail` - Email verification

**Test Cases per Method:**
- **CreateUser:**
  - Successful creation
  - Missing first name (InvalidArgument)
  - Missing surname (InvalidArgument)
  - Repository errors (Internal)

- **GetUser:**
  - Successful retrieval
  - Missing ID (InvalidArgument)
  - User not found (NotFound)

- **UpdateUser:**
  - Successful update
  - Missing ID (InvalidArgument)
  - Repository errors (Internal)

- **DeleteUser:**
  - Successful deletion
  - Missing ID (InvalidArgument)
  - Repository errors (Internal)

- **ListUsers:**
  - Successful listing
  - Default pagination values
  - Page size capping (max 100)
  - Repository errors (Internal)

- **VerifyEmail:**
  - Successful verification
  - Missing ID (InvalidArgument)
  - Repository errors (Internal)

### 3. Legacy Handler Tests (`internal/handlers/handler_users_test.go`)

Original HTTP handler tests (still functional):
- Successful user creation
- Invalid JSON body
- Database errors
- Duplicate email handling

## Running Tests

### Run All Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Run Specific Test Suites

```bash
# Repository layer tests
go test -v ./services/user-service/repository

# Service layer tests
go test -v ./services/user-service/service

# Legacy handler tests
go test -v ./internal/handlers

# Database tests
go test -v ./internal/database
```

### Run Specific Tests

```bash
# Run a specific test function
go test -v -run TestUserService_CreateUser ./services/user-service/service

# Run tests matching a pattern
go test -v -run "TestPostgresUserRepository_.*" ./services/user-service/repository
```

### Run Tests with Race Detection

```bash
# Detect race conditions
go test -race ./...
```

### Run Short Tests Only

```bash
# Skip long-running integration tests
go test -short ./...
```

## Using the Makefile

The project includes a Makefile with convenient test commands:

```bash
# Run all tests
make -f Makefile.microservices test

# Run unit tests only
make -f Makefile.microservices test-unit

# Run integration tests
make -f Makefile.microservices test-integration

# Generate coverage report
make -f Makefile.microservices test-coverage
```

## Test Dependencies

The project uses the following testing libraries:

- **testify/assert** - Fluent assertions
- **testify/require** - Assertions that stop test execution on failure
- **testify/mock** - Mock object framework
- **go-sqlmock** - SQL driver mock for testing database interactions
- **testcontainers-go** - Docker containers for integration tests

## Writing New Tests

### Repository Layer Test Example

```go
func TestPostgresUserRepository_YourFunction(t *testing.T) {
    tests := []struct {
        name          string
        input         interface{}
        mockSetup     func(sqlmock.Sqlmock)
        expectedError bool
        validate      func(*testing.T, *Result)
    }{
        {
            name: "successful case",
            input: someInput,
            mockSetup: func(mock sqlmock.Sqlmock) {
                mock.ExpectQuery("SELECT (.+)").
                    WillReturnRows(sqlmock.NewRows(...))
            },
            expectedError: false,
            validate: func(t *testing.T, result *Result) {
                assert.NotNil(t, result)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, mock, err := sqlmock.New()
            require.NoError(t, err)
            defer db.Close()

            tt.mockSetup(mock)

            repo := &postgresUserRepository{db: db}
            result, err := repo.YourFunction(context.Background(), tt.input)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                if tt.validate != nil {
                    tt.validate(t, result)
                }
            }

            assert.NoError(t, mock.ExpectationsWereMet())
        })
    }
}
```

### Service Layer Test Example

```go
func TestUserService_YourMethod(t *testing.T) {
    tests := []struct {
        name          string
        request       *pb.YourRequest
        mockSetup     func(*MockUserRepository)
        expectedError bool
        errorCode     codes.Code
    }{
        {
            name: "successful case",
            request: &pb.YourRequest{...},
            mockSetup: func(repo *MockUserRepository) {
                repo.On("YourMethod", mock.Anything, mock.Anything).
                    Return(expectedResult, nil)
            },
            expectedError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockUserRepository)
            tt.mockSetup(mockRepo)

            service := NewUserService(mockRepo)
            resp, err := service.YourMethod(context.Background(), tt.request)

            if tt.expectedError {
                assert.Error(t, err)
                st, ok := status.FromError(err)
                assert.True(t, ok)
                assert.Equal(t, tt.errorCode, st.Code())
            } else {
                assert.NoError(t, err)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

## Integration Tests

For full end-to-end testing:

```bash
# Start services with docker-compose
docker-compose -f docker-compose.microservices.yml up -d

# Run API tests
make -f Makefile.microservices test-api

# Or manually:
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "surname": "User",
    "email": "test@example.com"
  }'
```

## Continuous Integration

Example GitHub Actions workflow:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Install protobuf
        run: |
          sudo apt-get update
          sudo apt-get install -y protobuf-compiler

      - name: Generate protobuf
        run: cd proto && make install-tools && make all

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Best Practices

1. **Use Table-Driven Tests** - Group related test cases together
2. **Test Error Paths** - Always test both success and failure scenarios
3. **Mock External Dependencies** - Use mocks for databases, external services
4. **Verify Mock Expectations** - Always call `AssertExpectations` or `ExpectationsWereMet`
5. **Use Descriptive Names** - Test names should clearly describe what they test
6. **Keep Tests Independent** - Tests should not depend on each other
7. **Clean Up Resources** - Always defer cleanup (e.g., `defer db.Close()`)
8. **Test Edge Cases** - Empty inputs, nil values, boundary conditions

## Troubleshooting

### Tests Failing to Connect to Database

This is expected for unit tests - they use mocks. For integration tests:
```bash
# Make sure Supabase credentials are set
export DB_HOST=db.xxxxx.supabase.co
export DB_PASSWORD=your-password
```

### Protobuf Import Errors

Generate protobuf code first:
```bash
cd proto
make install-tools
make all
```

### Mock Expectations Not Met

Check that:
1. Mock setup matches the actual call parameters
2. You're calling `AssertExpectations(t)` at the end
3. Function signatures match between mock and implementation

### Race Condition Warnings

Run with race detector:
```bash
go test -race ./...
```

Fix any reported issues before committing.

## Coverage Goals

- **Repository Layer:** 80%+ coverage
- **Service Layer:** 85%+ coverage
- **Critical Paths:** 95%+ coverage

Check current coverage:
```bash
go test -cover ./... | grep coverage
```

## Next Steps

1. Add integration tests for API Gateway
2. Add end-to-end tests with testcontainers
3. Add performance/load tests
4. Add benchmark tests for critical paths
5. Set up coverage reporting in CI/CD
