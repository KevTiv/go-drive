package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "succesful user creation",
			requestBody: map[string]interface{}{
				"id":         "123e4567-e89b-12d3-a456-426614174000",
				"first_name": "John",
				"surname":    "Doe",
				"email":      "john@example.com",
				"phone":      "+1234567890",
				"country":    "USA",
				"region":     "California",
				"city":       "San Francisco",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("123e4567-e89b-12d3-a456-426614174000", "John", "Doe", "john@example.com", "+1234567890", "USA", "California", "San Francisco").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":      float64(1),
				"message": "User created successfully",
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid json",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database expectations - request should fail before DB call
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "database error",
			requestBody: map[string]interface{}{
				"id":         "223e4567-e89b-12d3-a456-426614174001",
				"first_name": "Jane",
				"surname":    "Doe",
				"email":      "jane@example.com",
				"phone":      "+0987654321",
				"country":    "Canada",
				"region":     "Ontario",
				"city":       "Toronto",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("223e4567-e89b-12d3-a456-426614174001", "Jane", "Doe", "jane@example.com", "+0987654321", "Canada", "Ontario", "Toronto").
					WillReturnError(sql.ErrConnDone)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "duplicate email error",
			requestBody: map[string]interface{}{
				"id":         "323e4567-e89b-12d3-a456-426614174002",
				"first_name": "Duplicate",
				"surname":    "User",
				"email":      "existing@example.com",
				"phone":      "+1122334455",
				"country":    "UK",
				"region":     "England",
				"city":       "London",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("323e4567-e89b-12d3-a456-426614174002", "Duplicate", "User", "existing@example.com", "+1122334455", "UK", "England", "London").
					WillReturnError(sql.ErrNoRows) // Or your specific duplicate error
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err, "failed to crate mock database")
			defer db.Close()

			// Setup mock expectations
			tt.mockSetup(mock)

			// Create request body
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err, "failed to marshal request body")
			}

			// Crate HTTP request
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-type", "application/json")

			// Crate response recorder
			w := httptest.NewRecorder()

			// Create handler with mock db
			handler := NewCreateUserHandler(db)
			handler.ServeHTTP(w, req)

			// Assert response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body if expected
			if tt.expectedBody != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "failed to unmarshal response")

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			// Ensure all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled mock expectations")
		})
	}
}
