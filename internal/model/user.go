package model

import "time"

type User struct {
	ID         string    `bson:"_id,omitempty" json:"id"`
	TenantID   string    `bson:"tenant_id" json:"tenant_id"`
	Email      string    `bson:"email" json:"email"`
	FirstName  string    `bson:"first_name" json:"first_name"`
	LastName   string    `bson:"last_name" json:"last_name"`
	Department string    `bson:"department" json:"department"`
	Status     string    `bson:"status" json:"status"` // active / inactive
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Department string `json:"department"`
	Status     string `json:"status"`
}

type ListUsersResponse struct {
	Users      []User `json:"users"`
	Page       int64  `json:"page"`
	Limit      int64  `json:"limit"`
	Total      int64  `json:"total"`
	TotalPages int64  `json:"total_pages"`
}

// CSVUploadResult is returned after a bulk CSV upload
type CSVUploadResult struct {
	SuccessCount int        `json:"success_count"`
	FailureCount int        `json:"failure_count"`
	Errors       []RowError `json:"errors,omitempty"`
}

// RowError describes why a particular CSV row failed
type RowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}
