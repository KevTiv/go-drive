package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-drive/internal/models"
	"net/http"
)

// UserHandler contains dependencies for user operations
type UserHandler struct {
	db *sql.DB
}

type CreateUserRequest struct {
	models.Users
}
type UpdateUserRequest struct {
	FirstName string
	Surname   string
	Email     string
	Phone     string
	Country   string
	Region    string
	Type      string
}

type UserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// NewCrateUserHanlder created a new handler with dependencies
func NewCreateUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func handleCreateUser(w http.ResponseWriter, r *http.Request, h *UserHandler) {
	var req CreateUserRequest

	// decode request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// validate input
	if err := req.ValidatePayload(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into database
	result, err := h.db.Exec(
		"INSERT INTO users (id, firstname, surname, email, phone, country, region, city) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.ID,
		req.FirstName,
		req.Surname,
		req.Email,
		req.Phone,
		req.Country,
		req.Region,
		req.City,
	)
	if err != nil {
		// Check for duplicate email error
		if err == sql.ErrNoRows {
			h.respondError(w, "Email already exists", http.StatusConflict)
			return
		}
		h.respondError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		h.respondError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, http.StatusCreated, UserResponse{
		ID:      id,
		Message: "User created successfully",
	})
}

// ServeHTTP implements http.Handler interface
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateUser(w, r, h)
	default:
		return
	}
}

// Helper methods
func (h *UserHandler) respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *UserHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (u *CreateUserRequest) ValidatePayload() error {
	if u.FirstName == "" || u.Surname == "" {
		return fmt.Errorf("first_name and surname are required")
	}
	return nil
}
