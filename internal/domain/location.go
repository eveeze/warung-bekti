package domain

import (
	"time"

	"github.com/google/uuid"
)

// LocationType represents the categorical type of a location
type LocationType string

const (
	LocationTypeShelf       LocationType = "shelf"
	LocationTypeFridge      LocationType = "fridge"
	LocationTypeShowcase    LocationType = "showcase"
	LocationTypeFloor       LocationType = "floor"
	LocationTypeWarehouse   LocationType = "warehouse"
	LocationTypeCashierArea LocationType = "cashier_area"
)

// Location represents a physical location in the store (for 2D/3D layouts)
type Location struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Category    LocationType `json:"category"`
	XCoordinate float64      `json:"x_coordinate"`
	YCoordinate float64      `json:"y_coordinate"`
	ZCoordinate float64      `json:"z_coordinate"`
	Width       float64      `json:"width"`
	Depth       float64      `json:"depth"`
	Height      float64      `json:"height"`
	Description *string      `json:"description,omitempty"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// LocationCreateInput is the input payload for creating a location
type LocationCreateInput struct {
	Name        string       `json:"name" validate:"required"`
	Category    LocationType `json:"category" validate:"required"`
	XCoordinate float64      `json:"x_coordinate"`
	YCoordinate float64      `json:"y_coordinate"`
	ZCoordinate float64      `json:"z_coordinate"`
	Width       float64      `json:"width"`
	Depth       float64      `json:"depth"`
	Height      float64      `json:"height"`
	Description *string      `json:"description,omitempty"`
}

// LocationUpdateInput is the input payload for updating a location
type LocationUpdateInput struct {
	Name        *string       `json:"name,omitempty"`
	Category    *LocationType `json:"category,omitempty"`
	XCoordinate *float64      `json:"x_coordinate,omitempty"`
	YCoordinate *float64      `json:"y_coordinate,omitempty"`
	ZCoordinate *float64      `json:"z_coordinate,omitempty"`
	Width       *float64      `json:"width,omitempty"`
	Depth       *float64      `json:"depth,omitempty"`
	Height      *float64      `json:"height,omitempty"`
	Description *string       `json:"description,omitempty"`
	IsActive    *bool         `json:"is_active,omitempty"`
}
