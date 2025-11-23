package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go-drive/proto/user"
)

// MockUserServiceClient is a mock implementation of the gRPC user service client
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.CreateUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UpdateUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*pb.DeleteUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) ListUsers(ctx context.Context, in *pb.ListUsersRequest, opts ...grpc.CallOption) (*pb.ListUsersResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ListUsersResponse), args.Error(1)
}

func (m *MockUserServiceClient) VerifyEmail(ctx context.Context, in *pb.VerifyEmailRequest, opts ...grpc.CallOption) (*pb.VerifyEmailResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.VerifyEmailResponse), args.Error(1)
}

func TestAPIGateway_HandleCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserServiceClient)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful user creation",
			requestBody: pb.CreateUserRequest{
				FirstName: "John",
				Surname:   "Doe",
				Email:     "john@example.com",
				Country:   "USA",
			},
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("CreateUser", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).
					Return(&pb.CreateUserResponse{
						User: &pb.User{
							Id:        "test-id",
							FirstName: "John",
							Surname:   "Doe",
							Email:     "john@example.com",
							Country:   "USA",
							IsActive:  true,
							CreatedAt: timestamppb.Now(),
							UpdatedAt: timestamppb.Now(),
						},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp pb.CreateUserResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotNil(t, resp.User)
				assert.Equal(t, "John", resp.User.FirstName)
				assert.Equal(t, "john@example.com", resp.User.Email)
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(mockClient *MockUserServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "grpc service error",
			requestBody: pb.CreateUserRequest{
				FirstName: "Jane",
				Surname:   "Doe",
				Email:     "jane@example.com",
			},
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("CreateUser", mock.Anything, mock.AnythingOfType("*user.CreateUserRequest")).
					Return(nil, status.Error(codes.Internal, "database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockUserServiceClient)
			tt.mockSetup(mockClient)

			gw := &APIGateway{
				userClient: mockClient,
			}

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			gw.handleCreateUser(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, rec)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAPIGateway_HandleGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserServiceClient)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful user retrieval",
			userID: "test-id",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("GetUser", mock.Anything, &pb.GetUserRequest{Id: "test-id"}).
					Return(&pb.GetUserResponse{
						User: &pb.User{
							Id:        "test-id",
							FirstName: "John",
							Surname:   "Doe",
							Email:     "john@example.com",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp pb.GetUserResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotNil(t, resp.User)
				assert.Equal(t, "test-id", resp.User.Id)
			},
		},
		{
			name:           "missing user id",
			userID:         "",
			mockSetup:      func(mockClient *MockUserServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("GetUser", mock.Anything, &pb.GetUserRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockUserServiceClient)
			tt.mockSetup(mockClient)

			gw := &APIGateway{
				userClient: mockClient,
			}

			url := "/api/v1/users"
			if tt.userID != "" {
				url += "?id=" + tt.userID
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			gw.handleGetUser(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, rec)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAPIGateway_HandleListUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockUserServiceClient)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful user listing",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("ListUsers", mock.Anything, mock.AnythingOfType("*user.ListUsersRequest")).
					Return(&pb.ListUsersResponse{
						Users: []*pb.User{
							{
								Id:        "id1",
								FirstName: "John",
								Surname:   "Doe",
								Email:     "john@example.com",
							},
							{
								Id:        "id2",
								FirstName: "Jane",
								Surname:   "Smith",
								Email:     "jane@example.com",
							},
						},
						TotalCount: 2,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp pb.ListUsersResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Len(t, resp.Users, 2)
				assert.Equal(t, int32(2), resp.TotalCount)
			},
		},
		{
			name: "grpc service error",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("ListUsers", mock.Anything, mock.AnythingOfType("*user.ListUsersRequest")).
					Return(nil, status.Error(codes.Internal, "database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockUserServiceClient)
			tt.mockSetup(mockClient)

			gw := &APIGateway{
				userClient: mockClient,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
			rec := httptest.NewRecorder()

			gw.handleListUsers(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, rec)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAPIGateway_HandleUpdateUser(t *testing.T) {
	firstName := "Jane"

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		mockSetup      func(*MockUserServiceClient)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful user update",
			method: http.MethodPut,
			requestBody: pb.UpdateUserRequest{
				Id:        "test-id",
				FirstName: &firstName,
			},
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("UpdateUser", mock.Anything, mock.AnythingOfType("*user.UpdateUserRequest")).
					Return(&pb.UpdateUserResponse{
						User: &pb.User{
							Id:        "test-id",
							FirstName: "Jane",
							Surname:   "Doe",
							Email:     "jane@example.com",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp pb.UpdateUserResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotNil(t, resp.User)
				assert.Equal(t, "Jane", resp.User.FirstName)
			},
		},
		{
			name:           "invalid request body",
			method:         http.MethodPut,
			requestBody:    "invalid json",
			mockSetup:      func(mockClient *MockUserServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "supports PATCH method",
			method: http.MethodPatch,
			requestBody: pb.UpdateUserRequest{
				Id:        "test-id",
				FirstName: &firstName,
			},
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("UpdateUser", mock.Anything, mock.AnythingOfType("*user.UpdateUserRequest")).
					Return(&pb.UpdateUserResponse{
						User: &pb.User{
							Id:        "test-id",
							FirstName: "Jane",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockUserServiceClient)
			tt.mockSetup(mockClient)

			gw := &APIGateway{
				userClient: mockClient,
			}

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/users", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			gw.handleUpdateUser(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, rec)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAPIGateway_HandleDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserServiceClient)
		expectedStatus int
	}{
		{
			name:   "successful user deletion",
			userID: "test-id",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("DeleteUser", mock.Anything, &pb.DeleteUserRequest{Id: "test-id"}).
					Return(&pb.DeleteUserResponse{Message: "User deleted successfully"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user id",
			userID:         "",
			mockSetup:      func(mockClient *MockUserServiceClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "grpc service error",
			userID: "test-id",
			mockSetup: func(mockClient *MockUserServiceClient) {
				mockClient.On("DeleteUser", mock.Anything, &pb.DeleteUserRequest{Id: "test-id"}).
					Return(nil, status.Error(codes.Internal, "database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockUserServiceClient)
			tt.mockSetup(mockClient)

			gw := &APIGateway{
				userClient: mockClient,
			}

			url := "/api/v1/users"
			if tt.userID != "" {
				url += "?id=" + tt.userID
			}

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()

			gw.handleDeleteUser(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestAPIGateway_HandleHealth(t *testing.T) {
	gw := &APIGateway{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	gw.handleHealth(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
}

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := corsMiddleware(handler)

	t.Run("OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("Regular request with CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestAPIGateway_MethodNotAllowed(t *testing.T) {
	gw := &APIGateway{}

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		method  string
	}{
		{
			name:    "CreateUser with GET",
			handler: gw.handleCreateUser,
			method:  http.MethodGet,
		},
		{
			name:    "GetUser with POST",
			handler: gw.handleGetUser,
			method:  http.MethodPost,
		},
		{
			name:    "ListUsers with POST",
			handler: gw.handleListUsers,
			method:  http.MethodPost,
		},
		{
			name:    "UpdateUser with GET",
			handler: gw.handleUpdateUser,
			method:  http.MethodGet,
		},
		{
			name:    "DeleteUser with POST",
			handler: gw.handleDeleteUser,
			method:  http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/users", nil)
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		})
	}
}
