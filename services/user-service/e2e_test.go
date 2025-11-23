//go:build e2e
// +build e2e

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "go-drive/proto/user"
)

// This test requires a running user service and database
// Run with: go test -tags=e2e -v ./services/user-service/...

func getTestClient(t *testing.T) (pb.UserServiceClient, *grpc.ClientConn) {
	addr := os.Getenv("USER_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err, "Failed to connect to user service")

	client := pb.NewUserServiceClient(conn)
	return client, conn
}

func TestE2E_UserLifecycle(t *testing.T) {
	client, conn := getTestClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate unique email for this test run
	testEmail := "e2e-test-" + time.Now().Format("20060102150405") + "@example.com"

	var userID string

	t.Run("Create User", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			FirstName: "E2E",
			Surname:   "Test",
			Email:     testEmail,
			Phone:     "+1234567890",
			Country:   "TestLand",
			Region:    "TestRegion",
			City:      "TestCity",
			Type:      "premium",
		}

		resp, err := client.CreateUser(ctx, req)
		require.NoError(t, err, "Failed to create user")
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)

		assert.NotEmpty(t, resp.User.Id)
		assert.Equal(t, "E2E", resp.User.FirstName)
		assert.Equal(t, "Test", resp.User.Surname)
		assert.Equal(t, testEmail, resp.User.Email)
		assert.Equal(t, "premium", resp.User.Type)
		assert.True(t, resp.User.IsActive)
		assert.False(t, resp.User.EmailVerified)

		userID = resp.User.Id
		t.Logf("Created user with ID: %s", userID)
	})

	t.Run("Get User", func(t *testing.T) {
		require.NotEmpty(t, userID, "User ID is required")

		req := &pb.GetUserRequest{
			Id: userID,
		}

		resp, err := client.GetUser(ctx, req)
		require.NoError(t, err, "Failed to get user")
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)

		assert.Equal(t, userID, resp.User.Id)
		assert.Equal(t, "E2E", resp.User.FirstName)
		assert.Equal(t, testEmail, resp.User.Email)
	})

	t.Run("Update User", func(t *testing.T) {
		require.NotEmpty(t, userID, "User ID is required")

		newFirstName := "UpdatedE2E"
		newPhone := "+9876543210"

		req := &pb.UpdateUserRequest{
			Id:        userID,
			FirstName: &newFirstName,
			Phone:     &newPhone,
		}

		resp, err := client.UpdateUser(ctx, req)
		require.NoError(t, err, "Failed to update user")
		require.NotNil(t, resp)
		require.NotNil(t, resp.User)

		assert.Equal(t, userID, resp.User.Id)
		assert.Equal(t, "UpdatedE2E", resp.User.FirstName)
		assert.Equal(t, "+9876543210", resp.User.Phone)
		assert.Equal(t, "Test", resp.User.Surname) // Unchanged
	})

	t.Run("List Users", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := client.ListUsers(ctx, req)
		require.NoError(t, err, "Failed to list users")
		require.NotNil(t, resp)

		// Should find at least our test user
		assert.Greater(t, len(resp.Users), 0)
		assert.Greater(t, resp.TotalCount, int32(0))

		// Verify our user is in the list
		found := false
		for _, user := range resp.Users {
			if user.Id == userID {
				found = true
				assert.Equal(t, "UpdatedE2E", user.FirstName)
				break
			}
		}
		assert.True(t, found, "Test user should be in the list")
	})

	t.Run("Verify Email", func(t *testing.T) {
		require.NotEmpty(t, userID, "User ID is required")

		req := &pb.VerifyEmailRequest{
			Id: userID,
		}

		resp, err := client.VerifyEmail(ctx, req)
		require.NoError(t, err, "Failed to verify email")
		require.NotNil(t, resp)
		assert.True(t, resp.Success)

		// Verify the email was actually verified
		getResp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: userID})
		require.NoError(t, err)
		assert.True(t, getResp.User.EmailVerified)
	})

	t.Run("Delete User", func(t *testing.T) {
		require.NotEmpty(t, userID, "User ID is required")

		req := &pb.DeleteUserRequest{
			Id: userID,
		}

		resp, err := client.DeleteUser(ctx, req)
		require.NoError(t, err, "Failed to delete user")
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Message)
		t.Logf("Delete message: %s", resp.Message)
	})

	t.Run("Get Deleted User Should Fail", func(t *testing.T) {
		require.NotEmpty(t, userID, "User ID is required")

		req := &pb.GetUserRequest{
			Id: userID,
		}

		_, err := client.GetUser(ctx, req)
		require.Error(t, err, "Getting deleted user should fail")

		st, ok := status.FromError(err)
		require.True(t, ok, "Error should be a gRPC status error")
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestE2E_ValidationErrors(t *testing.T) {
	client, conn := getTestClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("Create User Missing FirstName", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			Surname: "Test",
			Email:   "missing-firstname@example.com",
		}

		_, err := client.CreateUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Create User Missing Surname", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			FirstName: "Test",
			Email:     "missing-surname@example.com",
		}

		_, err := client.CreateUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Create User Missing Email", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			FirstName: "Test",
			Surname:   "User",
		}

		_, err := client.CreateUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Get User With Invalid ID", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Id: "nonexistent-id-12345",
		}

		_, err := client.GetUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Update User With Empty ID", func(t *testing.T) {
		firstName := "Test"
		req := &pb.UpdateUserRequest{
			Id:        "",
			FirstName: &firstName,
		}

		_, err := client.UpdateUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Delete User With Empty ID", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Id: "",
		}

		_, err := client.DeleteUser(ctx, req)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestE2E_ListUsersPagination(t *testing.T) {
	client, conn := getTestClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create multiple test users
	timestamp := time.Now().Format("20060102150405")
	createdIDs := []string{}

	for i := 0; i < 5; i++ {
		req := &pb.CreateUserRequest{
			FirstName: "PagTest",
			Surname:   "User",
			Email:     "pagination-test-" + timestamp + "-" + string(rune('a'+i)) + "@example.com",
		}

		resp, err := client.CreateUser(ctx, req)
		require.NoError(t, err)
		createdIDs = append(createdIDs, resp.User.Id)
	}

	// Cleanup
	defer func() {
		for _, id := range createdIDs {
			client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
		}
	}()

	t.Run("List with pagination", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Page:     1,
			PageSize: 2,
		}

		resp, err := client.ListUsers(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should get at most 2 users per page
		assert.LessOrEqual(t, len(resp.Users), 2)
		assert.Greater(t, resp.TotalCount, int32(0))
	})

	t.Run("List with default values", func(t *testing.T) {
		req := &pb.ListUsersRequest{}

		resp, err := client.ListUsers(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Greater(t, len(resp.Users), 0)
	})
}

func TestE2E_DuplicateEmail(t *testing.T) {
	client, conn := getTestClient(t)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	testEmail := "duplicate-test-" + time.Now().Format("20060102150405") + "@example.com"

	// Create first user
	req1 := &pb.CreateUserRequest{
		FirstName: "First",
		Surname:   "User",
		Email:     testEmail,
	}

	resp1, err := client.CreateUser(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	userID := resp1.User.Id

	// Cleanup
	defer func() {
		client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userID})
	}()

	// Try to create second user with same email
	req2 := &pb.CreateUserRequest{
		FirstName: "Second",
		Surname:   "User",
		Email:     testEmail,
	}

	_, err = client.CreateUser(ctx, req2)
	require.Error(t, err, "Creating user with duplicate email should fail")

	st, ok := status.FromError(err)
	require.True(t, ok)
	// Should be either AlreadyExists or Internal depending on DB error handling
	assert.Contains(t, []codes.Code{codes.AlreadyExists, codes.Internal}, st.Code())
}
