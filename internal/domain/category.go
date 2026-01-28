package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	
	// Relations (populated when needed)
	Parent   *Category   `json:"parent,omitempty"`
	Children []Category  `json:"children,omitempty"`
}

// CategoryCreateInput is the input for creating a category
type CategoryCreateInput struct {
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// CategoryUpdateInput is the input for updating a category
type CategoryUpdateInput struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}
