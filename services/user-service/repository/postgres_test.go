package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-drive/internal/database"
	pb "go-drive/proto/user"
)

func TestPostgresUserRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.CreateUserRequest
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		validate      func(*testing.T, *pb.User)
	}{
		{
			name: "successful user creation",
			request: &pb.CreateUserRequest{
				FirstName: "John",
				Surname:   "Doe",
				Email:     "john@example.com",
				Phone:     "+1234567890",
				Country:   "USA",
				Region:    "California",
				City:      "San Francisco",
				Type:      "premium",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						sqlmock.AnyArg(), // id (UUID)
						"John",
						"Doe",
						"john@example.com",
						"+1234567890",
						"USA",
						"California",
						"San Francisco",
						"premium",
						false,            // email_verified
						true,             // is_active
						sqlmock.AnyArg(), // created_at
						sqlmock.AnyArg(), // updated_at
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
			validate: func(t *testing.T, user *pb.User) {
				assert.NotEmpty(t, user.Id)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.Surname)
				assert.Equal(t, "john@example.com", user.Email)
				assert.Equal(t, "USA", user.Country)
				assert.False(t, user.EmailVerified)
				assert.True(t, user.IsActive)
				assert.NotNil(t, user.CreatedAt)
				assert.NotNil(t, user.UpdatedAt)
			},
		},
		{
			name: "database error",
			request: &pb.CreateUserRequest{
				FirstName: "Jane",
				Surname:   "Smith",
				Email:     "jane@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
		{
			name: "duplicate email error",
			request: &pb.CreateUserRequest{
				FirstName: "Duplicate",
				Surname:   "User",
				Email:     "existing@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			db, mock, err := sqlmock.New()
			require.NoError(t, err, "failed to create mock database")
			defer db.Close()

			// Setup mock expectations
			tt.mockSetup(mock)

			// Create repository
			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}

			// Execute
			user, err := repo.Create(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	fixedTime := time.Now()
	userID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		validate      func(*testing.T, *pb.User)
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				}).AddRow(
					userID, "John", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", true,
					true, fixedTime, fixedTime,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedError: false,
			validate: func(t *testing.T, user *pb.User) {
				assert.Equal(t, userID, user.Id)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.Surname)
				assert.Equal(t, "john@example.com", user.Email)
				assert.True(t, user.EmailVerified)
				assert.True(t, user.IsActive)
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent-id",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs("nonexistent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
		},
		{
			name:   "database error",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(userID).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}
			user, err := repo.GetByID(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresUserRepository_Update(t *testing.T) {
	fixedTime := time.Now()
	userID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name          string
		request       *pb.UpdateUserRequest
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		validate      func(*testing.T, *pb.User)
	}{
		{
			name: "successful update",
			request: &pb.UpdateUserRequest{
				Id:        userID,
				FirstName: stringPtr("Jane"),
				Email:     stringPtr("jane@example.com"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// First query to get existing user
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				}).AddRow(
					userID, "John", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", false,
					true, fixedTime, fixedTime,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(userID).
					WillReturnRows(rows)

				// Update query
				mock.ExpectExec("UPDATE users SET").
					WithArgs(
						userID,
						"Jane", // updated first name
						"Doe",
						"jane@example.com", // updated email
						"+1234567890",
						"USA",
						"California",
						"San Francisco",
						"premium",
						sqlmock.AnyArg(), // updated_at
					).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
			validate: func(t *testing.T, user *pb.User) {
				assert.Equal(t, "Jane", user.FirstName)
				assert.Equal(t, "jane@example.com", user.Email)
				assert.Equal(t, "Doe", user.Surname) // unchanged
			},
		},
		{
			name: "user not found",
			request: &pb.UpdateUserRequest{
				Id:        "nonexistent-id",
				FirstName: stringPtr("Jane"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs("nonexistent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: true,
		},
		{
			name: "database error on update",
			request: &pb.UpdateUserRequest{
				Id:        userID,
				FirstName: stringPtr("Jane"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				}).AddRow(
					userID, "John", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", false,
					true, fixedTime, fixedTime,
				)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE id").
					WithArgs(userID).
					WillReturnRows(rows)

				mock.ExpectExec("UPDATE users SET").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}
			user, err := repo.Update(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validate != nil {
					tt.validate(t, user)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:   "successful deletion",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET deleted_at").
					WithArgs(sqlmock.AnyArg(), userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET deleted_at").
					WithArgs(sqlmock.AnyArg(), userID).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}
			err = repo.Delete(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresUserRepository_List(t *testing.T) {
	fixedTime := time.Now()

	tests := []struct {
		name          string
		page          int32
		pageSize      int32
		filterType    *string
		filterActive  *bool
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		expectedCount int32
		validate      func(*testing.T, []*pb.User)
	}{
		{
			name:     "successful list with results",
			page:     1,
			pageSize: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				}).
					AddRow("id1", "John", "Doe", "john@example.com", "+1234567890",
						"USA", "California", "SF", "premium", true, true, fixedTime, fixedTime).
					AddRow("id2", "Jane", "Smith", "jane@example.com", "+0987654321",
						"Canada", "Ontario", "Toronto", "standard", false, true, fixedTime, fixedTime)

				mock.ExpectQuery("SELECT (.+) FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC").
					WithArgs(int32(10), int32(0)).
					WillReturnRows(rows)

				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE deleted_at IS NULL").
					WillReturnRows(countRows)
			},
			expectedError: false,
			expectedCount: 2,
			validate: func(t *testing.T, users []*pb.User) {
				assert.Len(t, users, 2)
				assert.Equal(t, "John", users[0].FirstName)
				assert.Equal(t, "Jane", users[1].FirstName)
			},
		},
		{
			name:         "list with type filter",
			page:         1,
			pageSize:     10,
			filterType:   stringPtr("premium"),
			filterActive: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				}).
					AddRow("id1", "John", "Doe", "john@example.com", "+1234567890",
						"USA", "California", "SF", "premium", true, true, fixedTime, fixedTime)

				mock.ExpectQuery("SELECT (.+) FROM users WHERE deleted_at IS NULL AND type").
					WithArgs("premium", int32(10), int32(0)).
					WillReturnRows(rows)

				countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE deleted_at IS NULL").
					WillReturnRows(countRows)
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:     "empty result",
			page:     1,
			pageSize: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC").
					WithArgs(int32(10), int32(0)).
					WillReturnRows(rows)

				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE deleted_at IS NULL").
					WillReturnRows(countRows)
			},
			expectedError: false,
			expectedCount: 0,
			validate: func(t *testing.T, users []*pb.User) {
				assert.Empty(t, users)
			},
		},
		{
			name:     "database error",
			page:     1,
			pageSize: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}
			users, totalCount, err := repo.List(context.Background(), tt.page, tt.pageSize, tt.filterType, tt.filterActive)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, totalCount)
				if tt.validate != nil {
					tt.validate(t, users)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresUserRepository_VerifyEmail(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:   "successful email verification",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET email_verified").
					WithArgs(sqlmock.AnyArg(), userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			userID: userID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET email_verified").
					WithArgs(sqlmock.AnyArg(), userID).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			repo := &postgresUserRepository{conn: &database.Connection{DB: db}}
			err = repo.VerifyEmail(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
