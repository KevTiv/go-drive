package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-drive/internal/database"
	pb "go-drive/proto/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserRepository interface {
	Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error)
	GetByID(ctx context.Context, id string) (*pb.User, error)
	Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int32, filterType *string, filterActive *bool) ([]*pb.User, int32, error)
	VerifyEmail(ctx context.Context, id string) error
	Close() error
	HealthCheck(ctx context.Context) error
}

type postgresUserRepository struct {
	conn *database.Connection
}

// NewPostgresUserRepository creates a new user repository using shared database connection
func NewPostgresUserRepository(cfg database.Config) (UserRepository, error) {
	conn, err := database.NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return &postgresUserRepository{conn: conn}, nil
}

// NewPostgresUserRepositoryFromConnection creates a repository from an existing connection
func NewPostgresUserRepositoryFromConnection(conn *database.Connection) UserRepository {
	return &postgresUserRepository{conn: conn}
}

func (r *postgresUserRepository) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	user := &pb.User{
		Id:            uuid.New().String(),
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
		CreatedAt:     timestamppb.Now(),
		UpdatedAt:     timestamppb.Now(),
	}

	query := `
		INSERT INTO users (id, firstname, surname, email, phone, country, region, city, type, email_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.conn.DB.ExecContext(ctx, query,
		user.Id, user.FirstName, user.Surname, user.Email, user.Phone,
		user.Country, user.Region, user.City, user.Type, user.EmailVerified,
		user.IsActive, user.CreatedAt.AsTime(), user.UpdatedAt.AsTime(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*pb.User, error) {
	query := `
		SELECT id, firstname, surname, email, phone, country, region, city, type, email_verified, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &pb.User{}
	var createdAt, updatedAt time.Time

	err := r.conn.DB.QueryRowContext(ctx, query, id).Scan(
		&user.Id, &user.FirstName, &user.Surname, &user.Email, &user.Phone,
		&user.Country, &user.Region, &user.City, &user.Type, &user.EmailVerified,
		&user.IsActive, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.CreatedAt = timestamppb.New(createdAt)
	user.UpdatedAt = timestamppb.New(updatedAt)

	return user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// First get the existing user
	user, err := r.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.Surname != nil {
		user.Surname = *req.Surname
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Country != nil {
		user.Country = *req.Country
	}
	if req.Region != nil {
		user.Region = *req.Region
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.Type != nil {
		user.Type = *req.Type
	}

	user.UpdatedAt = timestamppb.Now()

	query := `
		UPDATE users
		SET firstname = $2, surname = $3, email = $4, phone = $5, country = $6, region = $7, city = $8, type = $9, updated_at = $10
		WHERE id = $1
	`

	_, err = r.conn.DB.ExecContext(ctx, query,
		user.Id, user.FirstName, user.Surname, user.Email, user.Phone,
		user.Country, user.Region, user.City, user.Type, user.UpdatedAt.AsTime(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	_, err := r.conn.DB.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) List(ctx context.Context, page, pageSize int32, filterType *string, filterActive *bool) ([]*pb.User, int32, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id, firstname, surname, email, phone, country, region, city, type, email_verified, is_active, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	argCount := 1

	if filterType != nil {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filterType)
		argCount++
	}

	if filterActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *filterActive)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, pageSize, offset)

	rows, err := r.conn.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*pb.User
	for rows.Next() {
		user := &pb.User{}
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&user.Id, &user.FirstName, &user.Surname, &user.Email, &user.Phone,
			&user.Country, &user.Region, &user.City, &user.Type, &user.EmailVerified,
			&user.IsActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		user.CreatedAt = timestamppb.New(createdAt)
		user.UpdatedAt = timestamppb.New(updatedAt)
		users = append(users, user)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	var totalCount int32
	err = r.conn.DB.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, totalCount, nil
}

func (r *postgresUserRepository) VerifyEmail(ctx context.Context, id string) error {
	query := `UPDATE users SET email_verified = true, updated_at = $1 WHERE id = $2`
	_, err := r.conn.DB.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) Close() error {
	return r.conn.Close()
}

func (r *postgresUserRepository) HealthCheck(ctx context.Context) error {
	return r.conn.HealthCheck(ctx)
}
