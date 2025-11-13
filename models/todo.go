package models

import (
	"time"

	"gorm.io/gorm"
)

// Todo is the model for todo items
type Todo struct {
	// Replaced gorm.Model with explicit fields for Swagger compatibility
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Core model fields
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}
