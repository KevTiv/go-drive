package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go-drive/proto/user"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*pb.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int32, filterType *string, filterActive *bool) ([]*pb.User, int32, error) {
	args := m.Called(ctx, page, pageSize, filterType, filterActive)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*pb.User), args.Get(1).(int32), args.Error(2)
}

func (m *MockUserRepository) VerifyEmail(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.CreateUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.CreateUserResponse)
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
			mockSetup: func(repo *MockUserRepository) {
				expectedUser := &pb.User{
					Id:            "123e4567-e89b-12d3-a456-426614174000",
					FirstName:     "John",
					Surname:       "Doe",
					Email:         "john@example.com",
					Phone:         "+1234567890",
					Country:       "USA",
					Region:        "California",
					City:          "San Francisco",
					Type:          "premium",
					EmailVerified: false,
					IsActive:      true,
					CreatedAt:     timestamppb.Now(),
					UpdatedAt:     timestamppb.Now(),
				}
				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).
					Return(expectedUser, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.CreateUserResponse) {
				assert.NotNil(t, resp.User)
				assert.Equal(t, "John", resp.User.FirstName)
				assert.Equal(t, "Doe", resp.User.Surname)
				assert.Equal(t, "User created successfully", resp.Message)
			},
		},
		{
			name: "missing first name",
			request: &pb.CreateUserRequest{
				FirstName: "",
				Surname:   "Doe",
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "missing surname",
			request: &pb.CreateUserRequest{
				FirstName: "John",
				Surname:   "",
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "repository error",
			request: &pb.CreateUserRequest{
				FirstName: "John",
				Surname:   "Doe",
				Email:     "john@example.com",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).
					Return(nil, errors.New("database connection failed"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.CreateUser(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.GetUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.GetUserResponse)
	}{
		{
			name: "successful user retrieval",
			request: &pb.GetUserRequest{
				Id: "123e4567-e89b-12d3-a456-426614174000",
			},
			mockSetup: func(repo *MockUserRepository) {
				expectedUser := &pb.User{
					Id:            "123e4567-e89b-12d3-a456-426614174000",
					FirstName:     "John",
					Surname:       "Doe",
					Email:         "john@example.com",
					EmailVerified: true,
					IsActive:      true,
				}
				repo.On("GetByID", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(expectedUser, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.GetUserResponse) {
				assert.NotNil(t, resp.User)
				assert.Equal(t, "John", resp.User.FirstName)
			},
		},
		{
			name: "missing user id",
			request: &pb.GetUserRequest{
				Id: "",
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "user not found",
			request: &pb.GetUserRequest{
				Id: "nonexistent-id",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, "nonexistent-id").
					Return(nil, errors.New("user not found"))
			},
			expectedError: true,
			errorCode:     codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.GetUser(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	firstName := "Jane"
	email := "jane@example.com"

	tests := []struct {
		name          string
		request       *pb.UpdateUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.UpdateUserResponse)
	}{
		{
			name: "successful user update",
			request: &pb.UpdateUserRequest{
				Id:        "123e4567-e89b-12d3-a456-426614174000",
				FirstName: &firstName,
				Email:     &email,
			},
			mockSetup: func(repo *MockUserRepository) {
				updatedUser := &pb.User{
					Id:        "123e4567-e89b-12d3-a456-426614174000",
					FirstName: "Jane",
					Surname:   "Doe",
					Email:     "jane@example.com",
				}
				repo.On("Update", mock.Anything, mock.AnythingOfType("*user.UpdateUserRequest")).
					Return(updatedUser, nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.UpdateUserResponse) {
				assert.NotNil(t, resp.User)
				assert.Equal(t, "Jane", resp.User.FirstName)
				assert.Equal(t, "User updated successfully", resp.Message)
			},
		},
		{
			name: "missing user id",
			request: &pb.UpdateUserRequest{
				Id:        "",
				FirstName: &firstName,
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "repository error",
			request: &pb.UpdateUserRequest{
				Id:        "123e4567-e89b-12d3-a456-426614174000",
				FirstName: &firstName,
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Update", mock.Anything, mock.AnythingOfType("*user.UpdateUserRequest")).
					Return(nil, errors.New("update failed"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.UpdateUser(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.DeleteUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.DeleteUserResponse)
	}{
		{
			name: "successful user deletion",
			request: &pb.DeleteUserRequest{
				Id: "123e4567-e89b-12d3-a456-426614174000",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Delete", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.DeleteUserResponse) {
				assert.Equal(t, "User deleted successfully", resp.Message)
			},
		},
		{
			name: "missing user id",
			request: &pb.DeleteUserRequest{
				Id: "",
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "repository error",
			request: &pb.DeleteUserRequest{
				Id: "123e4567-e89b-12d3-a456-426614174000",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Delete", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(errors.New("delete failed"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.DeleteUser(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.ListUsersRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.ListUsersResponse)
	}{
		{
			name: "successful user listing",
			request: &pb.ListUsersRequest{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func(repo *MockUserRepository) {
				users := []*pb.User{
					{Id: "id1", FirstName: "John", Surname: "Doe"},
					{Id: "id2", FirstName: "Jane", Surname: "Smith"},
				}
				repo.On("List", mock.Anything, int32(1), int32(20), (*string)(nil), (*bool)(nil)).
					Return(users, int32(2), nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.ListUsersResponse) {
				assert.Len(t, resp.Users, 2)
				assert.Equal(t, int32(2), resp.TotalCount)
				assert.Equal(t, int32(1), resp.Page)
				assert.Equal(t, int32(20), resp.PageSize)
			},
		},
		{
			name: "default page and page size",
			request: &pb.ListUsersRequest{
				Page:     0,
				PageSize: 0,
			},
			mockSetup: func(repo *MockUserRepository) {
				users := []*pb.User{}
				repo.On("List", mock.Anything, int32(1), int32(20), (*string)(nil), (*bool)(nil)).
					Return(users, int32(0), nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.ListUsersResponse) {
				assert.Equal(t, int32(1), resp.Page)
				assert.Equal(t, int32(20), resp.PageSize)
			},
		},
		{
			name: "page size exceeds limit",
			request: &pb.ListUsersRequest{
				Page:     1,
				PageSize: 150,
			},
			mockSetup: func(repo *MockUserRepository) {
				users := []*pb.User{}
				repo.On("List", mock.Anything, int32(1), int32(20), (*string)(nil), (*bool)(nil)).
					Return(users, int32(0), nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.ListUsersResponse) {
				assert.Equal(t, int32(20), resp.PageSize)
			},
		},
		{
			name: "repository error",
			request: &pb.ListUsersRequest{
				Page:     1,
				PageSize: 20,
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, int32(1), int32(20), (*string)(nil), (*bool)(nil)).
					Return(nil, int32(0), errors.New("list failed"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.ListUsers(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_VerifyEmail(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.VerifyEmailRequest
		mockSetup     func(*MockUserRepository)
		expectedError bool
		errorCode     codes.Code
		validate      func(*testing.T, *pb.VerifyEmailResponse)
	}{
		{
			name: "successful email verification",
			request: &pb.VerifyEmailRequest{
				Id:                "123e4567-e89b-12d3-a456-426614174000",
				VerificationToken: "valid-token",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("VerifyEmail", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(nil)
			},
			expectedError: false,
			validate: func(t *testing.T, resp *pb.VerifyEmailResponse) {
				assert.True(t, resp.Success)
				assert.Equal(t, "Email verified successfully", resp.Message)
			},
		},
		{
			name: "missing user id",
			request: &pb.VerifyEmailRequest{
				Id:                "",
				VerificationToken: "token",
			},
			mockSetup:     func(repo *MockUserRepository) {},
			expectedError: true,
			errorCode:     codes.InvalidArgument,
		},
		{
			name: "repository error",
			request: &pb.VerifyEmailRequest{
				Id:                "123e4567-e89b-12d3-a456-426614174000",
				VerificationToken: "valid-token",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("VerifyEmail", mock.Anything, "123e4567-e89b-12d3-a456-426614174000").
					Return(errors.New("verification failed"))
			},
			expectedError: true,
			errorCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			resp, err := service.VerifyEmail(context.Background(), tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validate != nil {
					tt.validate(t, resp)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
