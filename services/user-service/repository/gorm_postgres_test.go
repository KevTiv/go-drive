package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-drive/internal/database"
	"go-drive/internal/domain"
	pb "go-drive/proto/user"
)

func setupGormMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create mock database")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err, "failed to open gorm connection")

	cleanup := func() {
		sqlDB, _ := gormDB.DB()
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestGormUserRepository_Create(t *testing.T) {
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
				// No transaction because SkipDefaultTransaction: true
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
					WithArgs(
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
						nil,              // deleted_at
						sqlmock.AnyArg(), // id
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
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
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupGormMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := &gormUserRepository{
				conn: &database.GormConnection{DB: gormDB},
			}

			user, err := repo.Create(context.Background(), tt.request)

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

func TestGormUserRepository_GetByID(t *testing.T) {
	fixedTime := time.Now()
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		validate      func(*testing.T, *pb.User)
	}{
		{
			name:   "successful retrieval",
			userID: userID.String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					userID, "John", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", true,
					true, fixedTime, fixedTime, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedError: false,
			validate: func(t *testing.T, user *pb.User) {
				assert.Equal(t, userID.String(), user.Id)
				assert.Equal(t, "John", user.FirstName)
				assert.Equal(t, "Doe", user.Surname)
				assert.Equal(t, "john@example.com", user.Email)
				assert.True(t, user.EmailVerified)
				assert.True(t, user.IsActive)
			},
		},
		{
			name:   "user not found",
			userID: uuid.New().String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupGormMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := &gormUserRepository{
				conn: &database.GormConnection{DB: gormDB},
			}

			user, err := repo.GetByID(context.Background(), tt.userID)

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

func TestGormUserRepository_Update(t *testing.T) {
	fixedTime := time.Now()
	userID := uuid.New()
	newFirstName := "Jane"

	tests := []struct {
		name          string
		request       *pb.UpdateUserRequest
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "successful update",
			request: &pb.UpdateUserRequest{
				Id:        userID.String(),
				FirstName: &newFirstName,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// First query to get existing user
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					userID, "John", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", false,
					true, fixedTime, fixedTime, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
					WithArgs(userID).
					WillReturnRows(rows)

				// Update query (no transaction)
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "firstname"=$1,"updated_at"=$2 WHERE "id" = $3`)).
					WithArgs("Jane", sqlmock.AnyArg(), userID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Reload user query
				rows2 := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					userID, "Jane", "Doe", "john@example.com", "+1234567890",
					"USA", "California", "San Francisco", "premium", false,
					true, fixedTime, fixedTime, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
					WithArgs(userID).
					WillReturnRows(rows2)
			},
			expectedError: false,
		},
		{
			name: "user not found",
			request: &pb.UpdateUserRequest{
				Id:        uuid.New().String(),
				FirstName: &newFirstName,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupGormMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := &gormUserRepository{
				conn: &database.GormConnection{DB: gormDB},
			}

			_, err := repo.Update(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGormUserRepository_Delete(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name:   "successful deletion",
			userID: userID.String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE "users"."id" = $2`)).
					WithArgs(sqlmock.AnyArg(), userID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			userID: userID.String(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1`)).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupGormMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := &gormUserRepository{
				conn: &database.GormConnection{DB: gormDB},
			}

			err := repo.Delete(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGormUserRepository_List(t *testing.T) {
	fixedTime := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	tests := []struct {
		name          string
		page          int32
		pageSize      int32
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
		expectedCount int32
	}{
		{
			name:     "successful list with results",
			page:     1,
			pageSize: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
					WillReturnRows(countRows)

				// List query
				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at", "deleted_at",
				}).
					AddRow(id1, "John", "Doe", "john@example.com", "+1234567890",
						"USA", "California", "SF", "premium", true, true, fixedTime, fixedTime, nil).
					AddRow(id2, "Jane", "Smith", "jane@example.com", "+0987654321",
						"Canada", "Ontario", "Toronto", "standard", false, true, fixedTime, fixedTime, nil)

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL ORDER BY created_at DESC LIMIT $1`)).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:     "empty result",
			page:     1,
			pageSize: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"id", "firstname", "surname", "email", "phone",
					"country", "region", "city", "type", "email_verified",
					"is_active", "created_at", "updated_at", "deleted_at",
				})
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
					WillReturnRows(rows)
			},
			expectedError: false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormDB, mock, cleanup := setupGormMock(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := &gormUserRepository{
				conn: &database.GormConnection{DB: gormDB},
			}

			_, totalCount, err := repo.List(context.Background(), tt.page, tt.pageSize, nil, nil)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, totalCount)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDomainUserToProto(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	domainUser := &domain.User{
		ID:            userID,
		FirstName:     "John",
		Surname:       "Doe",
		Email:         "john@example.com",
		Phone:         "+1234567890",
		Country:       "USA",
		Region:        "California",
		City:          "San Francisco",
		Type:          "premium",
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	protoUser := domainUserToProto(domainUser)

	assert.Equal(t, userID.String(), protoUser.Id)
	assert.Equal(t, "John", protoUser.FirstName)
	assert.Equal(t, "Doe", protoUser.Surname)
	assert.Equal(t, "john@example.com", protoUser.Email)
	assert.Equal(t, "+1234567890", protoUser.Phone)
	assert.Equal(t, "USA", protoUser.Country)
	assert.Equal(t, "California", protoUser.Region)
	assert.Equal(t, "San Francisco", protoUser.City)
	assert.Equal(t, "premium", protoUser.Type)
	assert.True(t, protoUser.EmailVerified)
	assert.True(t, protoUser.IsActive)
	assert.NotNil(t, protoUser.CreatedAt)
	assert.NotNil(t, protoUser.UpdatedAt)
}
