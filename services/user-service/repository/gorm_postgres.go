package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-drive/internal/database"
	"go-drive/internal/domain"
	pb "go-drive/proto/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type gormUserRepository struct {
	conn *database.GormConnection
}

// NewGormUserRepository creates a new user repository using GORM
func NewGormUserRepository(cfg database.Config) (UserRepository, error) {
	conn, err := database.NewGormConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &gormUserRepository{conn: conn}, nil
}

// NewGormUserRepositoryFromConnection creates a repository from an existing GORM connection
func NewGormUserRepositoryFromConnection(conn *database.GormConnection) UserRepository {
	return &gormUserRepository{conn: conn}
}

func (r *gormUserRepository) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	user := &domain.User{
		ID:            uuid.New(),
		FirstName:     req.FirstName,
		Surname:       req.Surname,
		Email:         req.Email,
		Phone:         req.Phone,
		Country:       req.Country,
		Region:        req.Region,
		City:          req.City,
		Type:          req.Type,
		EmailVerified: false,
		IsActive:      true,
	}

	if err := r.conn.DB.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return domainUserToProto(user), nil
}

func (r *gormUserRepository) GetByID(ctx context.Context, id string) (*pb.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user domain.User
	if err := r.conn.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return domainUserToProto(&user), nil
}

func (r *gormUserRepository) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// First get the existing user
	var user domain.User
	if err := r.conn.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["firstname"] = *req.FirstName
	}
	if req.Surname != nil {
		updates["surname"] = *req.Surname
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}

	// Update user
	if err := r.conn.DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Reload user to get updated values
	if err := r.conn.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	return domainUserToProto(&user), nil
}

func (r *gormUserRepository) Delete(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Soft delete
	if err := r.conn.DB.WithContext(ctx).Delete(&domain.User{}, userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *gormUserRepository) List(ctx context.Context, page, pageSize int32, filterType *string, filterActive *bool) ([]*pb.User, int32, error) {
	offset := (page - 1) * pageSize

	query := r.conn.DB.WithContext(ctx).Model(&domain.User{})

	// Apply filters
	if filterType != nil {
		query = query.Where("type = ?", *filterType)
	}
	if filterActive != nil {
		query = query.Where("is_active = ?", *filterActive)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	var users []domain.User
	if err := query.
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to proto
	protoUsers := make([]*pb.User, len(users))
	for i, user := range users {
		protoUsers[i] = domainUserToProto(&user)
	}

	return protoUsers, int32(totalCount), nil
}

func (r *gormUserRepository) VerifyEmail(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if err := r.conn.DB.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error; err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

func (r *gormUserRepository) Close() error {
	return r.conn.Close()
}

func (r *gormUserRepository) HealthCheck(ctx context.Context) error {
	return r.conn.HealthCheck(ctx)
}

// domainUserToProto converts a domain.User to pb.User
func domainUserToProto(user *domain.User) *pb.User {
	pbUser := &pb.User{
		Id:            user.ID.String(),
		FirstName:     user.FirstName,
		Surname:       user.Surname,
		Email:         user.Email,
		Phone:         user.Phone,
		Country:       user.Country,
		Region:        user.Region,
		City:          user.City,
		Type:          user.Type,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
	}

	return pbUser
}
